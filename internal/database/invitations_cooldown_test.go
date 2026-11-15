package database

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestInvitationResendCooldownBoundary(t *testing.T) {
	db, owner := invitationDB(t)
	i := seedInvitation(t, db, owner, "cooldown")
	deadline := i.LastSentAt.Add(2 * time.Hour)
	for _, now := range []time.Time{i.LastSentAt, i.LastSentAt.Add(time.Minute), deadline.Add(-time.Nanosecond)} {
		_, err := db.RotateInvitation(i.ID, "blocked", now)
		var cooldown *InvitationCooldownError
		if !errors.Is(err, ErrInvitationCooldown) || !errors.As(err, &cooldown) || !cooldown.AvailableAt.Equal(deadline) {
			t.Fatalf("cooldown at %v: %v", now, err)
		}
		current, err := db.GetInvitation(i.ID)
		if err != nil || current.TokenHash != i.TokenHash || !current.LastSentAt.Equal(i.LastSentAt) || !current.ExpiresAt.Equal(i.ExpiresAt) {
			t.Fatalf("blocked resend mutated invitation: %+v, %v", current, err)
		}
	}
	rotated, err := db.RotateInvitation(i.ID, "rotated", deadline)
	if err != nil || rotated.TokenHash != "rotated" || !rotated.LastSentAt.Equal(deadline) || !rotated.ResendAvailableAt().Equal(deadline.Add(2*time.Hour)) {
		t.Fatalf("exact boundary: %+v, %v", rotated, err)
	}
	if _, err = db.InvitationByToken(i.TokenHash); err == nil {
		t.Fatal("old token still works")
	}
	if _, err = db.RotateInvitation(i.ID, "blocked-again", deadline.Add(time.Hour)); !errors.Is(err, ErrInvitationCooldown) {
		t.Fatalf("rotated cooldown: %v", err)
	}
}

func TestInvitationResendCooldownConcurrent(t *testing.T) {
	db, owner := invitationDB(t)
	i := seedInvitation(t, db, owner, "concurrentresend")
	now := i.ResendAvailableAt()
	var successes atomic.Int32
	var wg sync.WaitGroup
	for n := 0; n < 12; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := db.RotateInvitation(i.ID, fmt.Sprintf("rotated-%d", n), now)
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, ErrInvitationCooldown) {
				t.Errorf("unexpected resend error: %v", err)
			}
		}(n)
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful resends: %d", successes.Load())
	}
}

func TestInvitationResendCooldownPersistsAcrossReopen(t *testing.T) {
	db, owner := invitationDB(t)
	i := seedInvitation(t, db, owner, "persisted")
	var sequence int
	var name, path string
	if err := db.conn.QueryRow("PRAGMA database_list").Scan(&sequence, &name, &path); err != nil {
		t.Fatal(err)
	}
	now := i.ResendAvailableAt()
	if _, err := db.RotateInvitation(i.ID, "persisted-rotation", now); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	current, err := reopened.GetInvitation(i.ID)
	if err != nil || !current.LastSentAt.Equal(now) || current.TokenHash != "persisted-rotation" {
		t.Fatalf("reopened invitation: %+v, %v", current, err)
	}
	if _, err = reopened.RotateInvitation(i.ID, "blocked", now.Add(time.Hour)); !errors.Is(err, ErrInvitationCooldown) {
		t.Fatalf("reopened cooldown: %v", err)
	}
	if _, err = reopened.RotateInvitation(i.ID, "eligible", now.Add(InvitationResendCooldown)); err != nil {
		t.Fatalf("reopened boundary: %v", err)
	}
}

func TestInvitationResendFailedRotationDoesNotChangeToken(t *testing.T) {
	db, owner := invitationDB(t)
	i := seedInvitation(t, db, owner, "collision")
	other := seedInvitation(t, db, owner, "other")
	if _, err := db.RotateInvitation(i.ID, other.TokenHash, i.ResendAvailableAt()); !errors.Is(err, ErrInvitationConflict) {
		t.Fatalf("expected conflicting hash: %v", err)
	}
	current, err := db.GetInvitation(i.ID)
	if err != nil || current.TokenHash != i.TokenHash || !current.LastSentAt.Equal(i.LastSentAt) {
		t.Fatalf("failed rotation mutated invitation: %+v, %v", current, err)
	}
}
