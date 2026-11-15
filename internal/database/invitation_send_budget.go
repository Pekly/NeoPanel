package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Every committed delivery attempt (including SMTP failure) consumes all three
// rolling budgets. This ledger deliberately has no foreign keys: revocation and
// account/server deletion must never refund sends.
const (
	InvitationSendWindow     = 2 * time.Hour
	InvitationSenderLimit    = 5
	InvitationServerLimit    = 10
	InvitationRecipientLimit = 1
)

type InvitationSendAvailability struct {
	AvailableAt time.Time `json:"available_at"`
	Scope       string    `json:"scope,omitempty"`
}

type InvitationSendBudgetError struct{ InvitationSendAvailability }

func (e *InvitationSendBudgetError) Error() string {
	return "Invitation " + e.Scope + " send limit reached (rolling 2 hours); wait until the next available time"
}
func (e *InvitationSendBudgetError) Unwrap() error { return ErrInvitationCooldown }

func (db *DB) migrateInvitationSendBudget() error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// DDL obtains the SQLite writer lock before reading the migration marker.
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS invitation_send_budget_migrations (version INTEGER PRIMARY KEY);
CREATE TABLE IF NOT EXISTS invitation_send_attempts (
 id INTEGER PRIMARY KEY, invitation_id TEXT NOT NULL, actor_id INTEGER NOT NULL,
 node_id TEXT NOT NULL, server_id TEXT NOT NULL, email TEXT NOT NULL, attempted_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS invitation_send_actor ON invitation_send_attempts(actor_id,attempted_at);
CREATE INDEX IF NOT EXISTS invitation_send_server ON invitation_send_attempts(node_id,server_id,attempted_at);
CREATE INDEX IF NOT EXISTS invitation_send_recipient ON invitation_send_attempts(email,attempted_at);
CREATE INDEX IF NOT EXISTS invitation_send_time ON invitation_send_attempts(attempted_at);
INSERT OR IGNORE INTO invitation_send_budget_migrations(version) VALUES (0);`)
	if err != nil {
		return err
	}
	var done int
	if err = tx.QueryRow(`SELECT count(*) FROM invitation_send_budget_migrations WHERE version=1`).Scan(&done); err != nil {
		return err
	}
	if done == 0 {
		// The old schema knows only the most recent attempt and original inviter.
		// Preserve that evidence for every status, including failed/revoked rows.
		rows, e := tx.Query(`SELECT id,inviter_id,node_id,server_id,email,last_sent_at FROM team_invitations`)
		if e != nil {
			return e
		}
		type historical struct {
			id, node, server, email string
			actor                   int
			at                      time.Time
		}
		var recent []historical
		cutoff := time.Now().UTC().Add(-InvitationSendWindow)
		for rows.Next() {
			var h historical
			if e = rows.Scan(&h.id, &h.actor, &h.node, &h.server, &h.email, &h.at); e != nil {
				rows.Close()
				return e
			}
			if h.at.After(cutoff) {
				recent = append(recent, h)
			}
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		for _, h := range recent {
			if _, err = tx.Exec(`INSERT INTO invitation_send_attempts(invitation_id,actor_id,node_id,server_id,email,attempted_at) VALUES (?,?,?,?,?,?)`, h.id, h.actor, h.node, h.server, normalizeInvitationRecipient(h.email), h.at.UnixNano()); err != nil {
				return err
			}
		}
		if _, err = tx.Exec(`INSERT INTO invitation_send_budget_migrations(version) VALUES (1)`); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func normalizeInvitationRecipient(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Recheck local authorization under the same write lock as reservation. Remote
// permissions and requested grants are verified by the HTTP handler on the node.
func invitationSendActor(tx *sql.Tx, i *Invitation, actorID int) error {
	var role string
	if err := tx.QueryRow(`SELECT role FROM users WHERE id=?`, actorID).Scan(&role); err != nil {
		return ErrInvitationForbidden
	}
	if i.NodeID != "local" {
		return nil
	}
	if err := invitationPermission(tx, i.ServerID, actorID, role); err != nil {
		return err
	}
	var owner int
	if err := tx.QueryRow(`SELECT owner_id FROM servers WHERE id=?`, i.ServerID).Scan(&owner); err != nil {
		return err
	}
	if role == "admin" || owner == actorID {
		return nil
	}
	var raw string
	if err := tx.QueryRow(`SELECT permissions FROM server_subusers WHERE server_id=? AND user_id=?`, i.ServerID, actorID).Scan(&raw); err != nil {
		return err
	}
	var grants []string
	if err := json.Unmarshal([]byte(raw), &grants); err != nil {
		return err
	}
	allowed := make(map[string]bool)
	for _, p := range grants {
		allowed[p] = true
	}
	for _, p := range i.Permissions {
		if !allowed[p] {
			return ErrInvitationForbidden
		}
	}
	return nil
}

// Nth newest expiry is when enough attempts have expired to permit one more,
// even if a conservative upgrade backfill is already above the new limit.
func invitationSendAvailability(tx *sql.Tx, actorID int, node, server, email string, now time.Time) (InvitationSendAvailability, error) {
	result := InvitationSendAvailability{AvailableAt: now}
	checks := []struct {
		scope, where string
		limit        int
		args         []any
	}{
		{"sender", "actor_id=?", InvitationSenderLimit, []any{actorID}},
		{"server", "node_id=? AND server_id=?", InvitationServerLimit, []any{node, server}},
	}
	if email != "" {
		checks = append(checks, struct {
			scope, where string
			limit        int
			args         []any
		}{"recipient", "email=?", InvitationRecipientLimit, []any{normalizeInvitationRecipient(email)}})
	}
	for _, c := range checks {
		args := append(c.args, now.Add(-InvitationSendWindow).UnixNano(), c.limit-1)
		var at int64
		err := tx.QueryRow(`SELECT attempted_at FROM invitation_send_attempts WHERE `+c.where+` AND attempted_at>? ORDER BY attempted_at DESC LIMIT 1 OFFSET ?`, args...).Scan(&at)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return result, err
		}
		available := time.Unix(0, at).UTC().Add(InvitationSendWindow)
		if available.After(result.AvailableAt) {
			result.AvailableAt, result.Scope = available, c.scope
		}
	}
	return result, nil
}

// Caller holds the writer lock acquired before any invitation/quota reads.
func reserveInvitationSend(tx *sql.Tx, i *Invitation, actorID int, now, cooldown time.Time) error {
	if _, err := tx.Exec(`DELETE FROM invitation_send_attempts WHERE attempted_at<=?`, now.Add(-InvitationSendWindow).UnixNano()); err != nil {
		return err
	}
	availability, err := invitationSendAvailability(tx, actorID, i.NodeID, i.ServerID, i.Email, now)
	if err != nil {
		return err
	}
	if cooldown.After(now) && !availability.AvailableAt.After(cooldown) {
		return &InvitationCooldownError{AvailableAt: cooldown}
	}
	if availability.AvailableAt.After(now) {
		return &InvitationSendBudgetError{availability}
	}
	_, err = tx.Exec(`INSERT INTO invitation_send_attempts(invitation_id,actor_id,node_id,server_id,email,attempted_at) VALUES (?,?,?,?,?,?)`, i.ID, actorID, i.NodeID, i.ServerID, normalizeInvitationRecipient(i.Email), now.UnixNano())
	return err
}

// InvitationSendAvailability exposes no ledger/account/recipient identities. It
// is advisory only; the delivery transaction always rechecks all budgets.
func (db *DB) InvitationSendAvailability(actorID int, node, server string, now time.Time) (InvitationSendAvailability, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return InvitationSendAvailability{}, err
	}
	defer tx.Rollback()
	result, err := invitationSendAvailability(tx, actorID, node, server, "", now)
	if err != nil {
		return result, err
	}
	return result, tx.Commit()
}
