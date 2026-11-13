package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Invitation secrets and recovery account IDs are deliberately never serialized.
type Invitation struct {
	ID             string    `json:"id"`
	ServerID       string    `json:"server_id"`
	NodeID         string    `json:"node_id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Permissions    []string  `json:"permissions"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	DeliveryStatus string    `json:"delivery_status"`
	InviterID      int       `json:"-"`
	TokenHash      string    `json:"-"`
	AccountID      int       `json:"-"`
	LastSentAt     time.Time `json:"-"`
}

var (
	ErrInvitationGone      = errors.New("invitation expired, revoked or already accepted")
	ErrInvitationConflict  = errors.New("invitation conflicts with an existing account or pending invitation; use the account's exact username and email, or manage the pending invitation")
	ErrInvitationIdentity  = errors.New("sign in to the existing account with the invited email; invitations cannot change passwords")
	ErrInvitationForbidden = errors.New("inviter no longer has permission or server was deleted")
	ErrInvitationBusy      = errors.New("remote acceptance is unresolved; retry acceptance or revoke before resending")
	ErrInvitationCooldown  = errors.New("wait at least one minute before resending")
	ErrInvitationMember    = fmt.Errorf("%w: this account already has server access; manage its existing permissions instead", ErrInvitationConflict)
)

func (db *DB) MigrateInvitations() error {
	_, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS team_invitations (
 id TEXT PRIMARY KEY, server_id TEXT NOT NULL, node_id TEXT NOT NULL,
 username TEXT NOT NULL, email TEXT NOT NULL, permissions TEXT NOT NULL,
 inviter_id INTEGER NOT NULL, token_hash TEXT NOT NULL UNIQUE,
 status TEXT NOT NULL, delivery_status TEXT NOT NULL, expires_at DATETIME NOT NULL,
 created_at DATETIME NOT NULL, last_sent_at DATETIME NOT NULL, account_id INTEGER NOT NULL DEFAULT 0
 );
 CREATE UNIQUE INDEX IF NOT EXISTS invitation_pending_email ON team_invitations(node_id,server_id,email) WHERE status IN ('pending','granting');
 CREATE TABLE IF NOT EXISTS invitation_grants (
 id TEXT PRIMARY KEY, server_id TEXT NOT NULL, user_id INTEGER NOT NULL, permissions TEXT NOT NULL,
 status TEXT NOT NULL, membership_id INTEGER NOT NULL DEFAULT 0
 );`)
	return err
}

const invitationColumns = `id,server_id,node_id,username,email,permissions,inviter_id,token_hash,status,delivery_status,expires_at,created_at,last_sent_at,account_id`

func scanInvitation(row interface{ Scan(...any) error }) (*Invitation, error) {
	i := &Invitation{}
	var permissions string
	err := row.Scan(&i.ID, &i.ServerID, &i.NodeID, &i.Username, &i.Email, &permissions, &i.InviterID, &i.TokenHash, &i.Status, &i.DeliveryStatus, &i.ExpiresAt, &i.CreatedAt, &i.LastSentAt, &i.AccountID)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(permissions), &i.Permissions)
	return i, err
}
func (db *DB) GetInvitation(id string) (*Invitation, error) {
	return scanInvitation(db.conn.QueryRow(`SELECT `+invitationColumns+` FROM team_invitations WHERE id=?`, id))
}
func (db *DB) InvitationByToken(hash string) (*Invitation, error) {
	return scanInvitation(db.conn.QueryRow(`SELECT `+invitationColumns+` FROM team_invitations WHERE token_hash=?`, hash))
}
func (db *DB) ListInvitations(server, node string) ([]*Invitation, error) {
	rows, err := db.conn.Query(`SELECT `+invitationColumns+` FROM team_invitations WHERE server_id=? AND node_id=? ORDER BY created_at DESC`, server, node)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Invitation{}
	for rows.Next() {
		i, e := scanInvitation(rows)
		if e != nil {
			return nil, e
		}
		if i.Status == "pending" && !i.ExpiresAt.After(time.Now()) {
			i.Status = "expired"
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
func (db *DB) CreateInvitation(i *Invitation) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE team_invitations SET status='expired' WHERE status='pending' AND expires_at<=?`, time.Now().UTC()); err != nil {
		return err
	}
	if err = invitationTarget(tx, i, 0); err != nil {
		return err
	}
	if err = invitationPending(tx, i); err != nil {
		return err
	}
	p, _ := json.Marshal(i.Permissions)
	_, err = tx.Exec(`INSERT INTO team_invitations (`+invitationColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,0)`, i.ID, i.ServerID, i.NodeID, i.Username, i.Email, string(p), i.InviterID, i.TokenHash, i.Status, i.DeliveryStatus, i.ExpiresAt, i.CreatedAt, i.LastSentAt)
	if err != nil {
		return fmt.Errorf("%w: duplicate pending invitation", ErrInvitationConflict)
	}
	return tx.Commit()
}
func (db *DB) InvitationDelivery(id, hash, status string) error {
	_, err := db.conn.Exec(`UPDATE team_invitations SET delivery_status=? WHERE id=? AND token_hash=?`, status, id, hash)
	return err
}
func (db *DB) RotateInvitation(id, hash string, now time.Time) (*Invitation, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE team_invitations SET id=id WHERE id=?`, id); err != nil {
		return nil, err
	}
	i, err := scanInvitation(tx.QueryRow(`SELECT `+invitationColumns+` FROM team_invitations WHERE id=?`, id))
	if err != nil {
		return nil, err
	}
	if i.Status == "granting" {
		return nil, ErrInvitationBusy
	}
	if i.Status != "pending" && i.Status != "expired" {
		return nil, ErrInvitationGone
	}
	if err = invitationTarget(tx, i, 0); err != nil {
		return nil, err
	}
	if err = invitationPending(tx, i); err != nil {
		return nil, err
	}
	if now.Sub(i.LastSentAt) < time.Minute {
		return nil, ErrInvitationCooldown
	}
	i.TokenHash = hash
	i.ExpiresAt = now.Add(72 * time.Hour)
	i.LastSentAt = now
	i.Status = "pending"
	i.DeliveryStatus = "sending"
	_, err = tx.Exec(`UPDATE team_invitations SET token_hash=?,expires_at=?,last_sent_at=?,status='pending',delivery_status='sending' WHERE id=?`, hash, i.ExpiresAt, now, id)
	if err != nil {
		return nil, ErrInvitationConflict
	}
	return i, tx.Commit()
}

func invitationLive(i *Invitation, hash string) error {
	if i.TokenHash != hash {
		return ErrInvitationGone
	}
	if i.Status != "pending" && i.Status != "granting" {
		return ErrInvitationGone
	}
	if !i.ExpiresAt.After(time.Now()) {
		return ErrInvitationGone
	}
	return nil
}
func invitationInviter(tx *sql.Tx, i *Invitation, local bool) (*User, error) {
	u := &User{}
	err := tx.QueryRow(`SELECT id,username,role FROM users WHERE id=?`, i.InviterID).Scan(&u.ID, &u.Username, &u.Role)
	if err != nil {
		return nil, ErrInvitationForbidden
	}
	if local {
		if err = invitationPermission(tx, i.ServerID, u.ID, u.Role); err != nil {
			return nil, err
		}
	}
	return u, nil
}
func invitationPermission(tx *sql.Tx, server string, id int, role string) error {
	var owner int
	if err := tx.QueryRow(`SELECT owner_id FROM servers WHERE id=?`, server).Scan(&owner); err != nil {
		return ErrInvitationForbidden
	}
	if role == "admin" || owner == id {
		return nil
	}
	var raw string
	if err := tx.QueryRow(`SELECT permissions FROM server_subusers WHERE server_id=? AND user_id=?`, server, id).Scan(&raw); err != nil {
		return ErrInvitationForbidden
	}
	var ps []string
	if json.Unmarshal([]byte(raw), &ps) != nil {
		return ErrInvitationForbidden
	}
	for _, p := range ps {
		if p == "subusers.manage" {
			return nil
		}
	}
	return ErrInvitationForbidden
}
func invitationAccount(tx *sql.Tx, i *Invitation, sessionID int, passwordHash string) (*User, error) {
	u := &User{}
	rows, err := tx.Query(`SELECT id,username,email,password_hash,role FROM users WHERE lower(username)=lower(?) OR lower(email)=?`, i.Username, i.Email)
	if err != nil {
		return nil, err
	}
	count := 0
	for rows.Next() {
		count++
		err = rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role)
		if err != nil {
			rows.Close()
			return nil, err
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if count > 1 {
		return nil, ErrInvitationConflict
	}
	if count == 1 {
		if !strings.EqualFold(u.Email, i.Email) || u.Username != i.Username {
			return nil, ErrInvitationConflict
		}
		if u.ID != i.AccountID && u.ID != sessionID {
			return nil, ErrInvitationIdentity
		}
		return u, nil
	}
	if i.AccountID != 0 {
		return nil, ErrInvitationConflict
	}
	if passwordHash == "" {
		return nil, ErrInvitationIdentity
	}
	res, err := tx.Exec(`INSERT INTO users(username,email,password_hash,role,created_at) VALUES(?,?,?,'user',?)`, i.Username, i.Email, passwordHash, time.Now().UTC())
	if err != nil {
		return nil, ErrInvitationConflict
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	i.AccountID = int(id)
	_, err = tx.Exec(`UPDATE team_invitations SET account_id=? WHERE id=?`, id, i.ID)
	if err != nil {
		return nil, err
	}
	return &User{ID: int(id), Username: i.Username, Email: i.Email, PasswordHash: passwordHash, Role: "user"}, nil
}

// Remote account creation is durable before the network hop. Only this invite may
// recover that exact account; retries never reset its password or change its role.
func (db *DB) ReserveInvitationAccount(id, hash string, sessionID int, passwordHash string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE team_invitations SET id=id WHERE id=?`, id); err != nil {
		return err
	}
	i, err := scanInvitation(tx.QueryRow(`SELECT `+invitationColumns+` FROM team_invitations WHERE id=?`, id))
	if err != nil {
		return err
	}
	if err = invitationLive(i, hash); err != nil {
		return err
	}
	if _, err = invitationInviter(tx, i, false); err != nil {
		return err
	}
	if _, err = invitationAccount(tx, i, sessionID, passwordHash); err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE team_invitations SET status='granting' WHERE id=?`, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// The transaction serializes acceptance, rotation and revocation. The remote
// callback must not access this DB; a worker receipt makes uncertain delivery retryable.
func (db *DB) AcceptInvitation(id, hash string, sessionID int, passwordHash string, remote func(*Invitation, *User, *User) error) (*User, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE team_invitations SET id=id WHERE id=?`, id); err != nil {
		return nil, err
	}
	i, err := scanInvitation(tx.QueryRow(`SELECT `+invitationColumns+` FROM team_invitations WHERE id=?`, id))
	if err != nil {
		return nil, err
	}
	if err = invitationLive(i, hash); err != nil {
		return nil, err
	}
	inviter, err := invitationInviter(tx, i, remote == nil)
	if err != nil {
		return nil, err
	}
	u, err := invitationAccount(tx, i, sessionID, passwordHash)
	if err != nil {
		return nil, err
	}
	if err = invitationTarget(tx, i, u.ID); err != nil {
		return nil, err
	}
	if remote != nil {
		err = remote(i, u, inviter)
	} else {
		p, _ := json.Marshal(i.Permissions)
		_, err = tx.Exec(`INSERT INTO server_subusers(server_id,user_id,permissions,created_at) VALUES(?,?,?,?)`, i.ServerID, u.ID, string(p), time.Now().UTC())
	}
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`UPDATE team_invitations SET status='accepted' WHERE id=?`, id)
	if err != nil {
		return nil, err
	}
	return u, tx.Commit()
}
func (db *DB) RevokeInvitation(id string, remote func(*Invitation) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE team_invitations SET id=id WHERE id=?`, id); err != nil {
		return err
	}
	i, err := scanInvitation(tx.QueryRow(`SELECT `+invitationColumns+` FROM team_invitations WHERE id=?`, id))
	if err != nil {
		return err
	}
	if i.Status == "accepted" {
		return ErrInvitationGone
	}
	if i.Status == "granting" {
		if remote == nil {
			return ErrInvitationBusy
		}
		if err = remote(i); err != nil {
			return err
		}
	}
	_, err = tx.Exec(`UPDATE team_invitations SET status='revoked' WHERE id=?`, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// InvitationMembership is a worker-side idempotent grant/tombstone. Authorization
// and membership creation are in one transaction; revoked/deleted memberships are
// never resurrected by a late retry. Used only over the trusted node hop.
func (db *DB) InvitationMembership(id, server string, userID, inviterID int, role string, permissions []string, expires time.Time, revoke bool) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE invitation_grants SET id=id WHERE id=?`, id); err != nil {
		return err
	}
	if err = invitationPermission(tx, server, inviterID, role); err != nil {
		return err
	}
	p, _ := json.Marshal(permissions)
	raw := string(p)
	var oldServer, oldPermissions, status string
	var oldUser, membership int
	err = tx.QueryRow(`SELECT server_id,user_id,permissions,status,membership_id FROM invitation_grants WHERE id=?`, id).Scan(&oldServer, &oldUser, &oldPermissions, &status, &membership)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil {
		if oldServer != server || (!revoke && (oldUser != userID || oldPermissions != raw)) {
			return ErrInvitationConflict
		}
		if !revoke {
			if status == "revoked" {
				return ErrInvitationGone
			}
			var currentPermissions string
			if err = tx.QueryRow(`SELECT permissions FROM server_subusers WHERE id=? AND server_id=? AND user_id=?`, membership, server, userID).Scan(&currentPermissions); err != nil || currentPermissions != raw {
				return ErrInvitationGone
			}
			return tx.Commit()
		}
		if membership != 0 {
			if _, err = tx.Exec(`DELETE FROM server_subusers WHERE id=? AND server_id=? AND user_id=?`, membership, server, oldUser); err != nil {
				return err
			}
		}
		_, err = tx.Exec(`UPDATE invitation_grants SET status='revoked' WHERE id=?`, id)
	} else if revoke {
		_, err = tx.Exec(`INSERT INTO invitation_grants(id,server_id,user_id,permissions,status) VALUES(?,?,0,'[]','revoked')`, id, server)
	} else {
		if !expires.After(time.Now()) {
			return ErrInvitationGone
		}
		var n int
		if err = tx.QueryRow(`SELECT count(*) FROM users WHERE id=?`, userID).Scan(&n); err != nil {
			return err
		}
		if n != 1 {
			return ErrInvitationConflict
		}
		var username, email string
		if err = tx.QueryRow(`SELECT username,email FROM users WHERE id=?`, userID).Scan(&username, &email); err != nil {
			return err
		}
		if err = invitationTarget(tx, &Invitation{ServerID: server, NodeID: "local", Username: username, Email: email}, userID); err != nil {
			return err
		}
		res, e := tx.Exec(`INSERT INTO server_subusers(server_id,user_id,permissions,created_at) VALUES(?,?,?,?)`, server, userID, raw, time.Now().UTC())
		if e != nil {
			return ErrInvitationConflict
		}
		mid, e := res.LastInsertId()
		if e != nil {
			return e
		}
		_, err = tx.Exec(`INSERT INTO invitation_grants(id,server_id,user_id,permissions,status,membership_id) VALUES(?,?,?,?,'accepted',?)`, id, server, userID, raw, mid)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Unlike legacy cluster sync, invitation provisioning never overwrites an account.
func (db *DB) ProvisionInvitationUser(u *User) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE users SET id=id WHERE id=?`, u.ID); err != nil {
		return err
	}
	var id int
	var username, email string
	rows, err := tx.Query(`SELECT id,username,email FROM users WHERE id=? OR lower(username)=lower(?) OR lower(email)=lower(?)`, u.ID, u.Username, u.Email)
	if err != nil {
		return err
	}
	count := 0
	for rows.Next() {
		count++
		err = rows.Scan(&id, &username, &email)
		if err != nil {
			rows.Close()
			return err
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if count != 0 {
		if count != 1 || id != u.ID || username != u.Username || email != u.Email {
			return ErrInvitationConflict
		}
		return tx.Commit()
	}
	_, err = tx.Exec(`INSERT INTO users(id,username,email,password_hash,role,created_at) VALUES(?,?,?,?,?,?)`, u.ID, u.Username, u.Email, u.PasswordHash, u.Role, time.Now().UTC())
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Target checks never grant access. The same checks run under the write transaction
// at local acceptance and on the worker, after its durable receipt lookup.
func invitationTarget(tx *sql.Tx, i *Invitation, expectedID int) error {
	rows, err := tx.Query(`SELECT id,username,email,role FROM users WHERE lower(username)=lower(?) OR lower(email)=lower(?) OR (?<>0 AND id=?)`, i.Username, i.Email, expectedID, expectedID)
	if err != nil {
		return err
	}
	var u User
	count := 0
	for rows.Next() {
		count++
		if err = rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role); err != nil {
			rows.Close()
			return err
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	if count != 1 || u.Username != i.Username || !strings.EqualFold(u.Email, i.Email) || (expectedID != 0 && expectedID != u.ID) {
		return ErrInvitationConflict
	}
	if u.Role == "admin" {
		return ErrInvitationMember
	}
	if i.NodeID != "local" {
		return nil
	}
	var n int
	err = tx.QueryRow(`SELECT (SELECT count(*) FROM servers WHERE id=? AND owner_id=?) + (SELECT count(*) FROM server_subusers WHERE server_id=? AND user_id=?)`, i.ServerID, u.ID, i.ServerID, u.ID).Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrInvitationMember
	}
	return nil
}

func invitationPending(tx *sql.Tx, i *Invitation) error {
	var n int
	err := tx.QueryRow(`SELECT count(*) FROM team_invitations WHERE id<>? AND (status='granting' OR (status='pending' AND expires_at>?)) AND (lower(username)=lower(?) OR lower(email)=lower(?)) AND (node_id=? AND server_id=? OR NOT (username=? AND lower(email)=lower(?)))`, i.ID, time.Now().UTC(), i.Username, i.Email, i.NodeID, i.ServerID, i.Username, i.Email).Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrInvitationConflict
	}
	return nil
}

func (db *DB) CheckInvitationTarget(i *Invitation) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return invitationTarget(tx, i, 0)
}

// CheckInvitationTargetOnNode uses the original caller, never the node-token role.
func (db *DB) CheckInvitationTargetOnNode(server, username, email string, expectedID, callerID int, role string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = invitationPermission(tx, server, callerID, role); err != nil {
		return err
	}
	return invitationTarget(tx, &Invitation{ServerID: server, NodeID: "local", Username: username, Email: email}, expectedID)
}
