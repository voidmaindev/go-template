package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
)

// These tests cover the CSRF-protection invariants of the OAuth state flow at
// the TokenStore layer: only the exact state string minted by the server (and
// not yet expired or consumed) can retrieve the associated state data.

func TestOAuthState_StoreAndGet_RoundTrip(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	state, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	want := &OAuthStateData{
		Provider:     "google",
		RedirectURL:  "/post-login",
		PKCEVerifier: "verifier-xyz",
	}

	if err := store.StoreOAuthState(ctx, state, want, time.Minute); err != nil {
		t.Fatalf("StoreOAuthState: %v", err)
	}

	got, err := store.GetOAuthState(ctx, state)
	if err != nil {
		t.Fatalf("GetOAuthState: %v", err)
	}
	if got.Provider != want.Provider || got.RedirectURL != want.RedirectURL || got.PKCEVerifier != want.PKCEVerifier {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestOAuthState_UnknownState_ReturnsMismatch(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// An attacker who did not initiate the flow cannot produce a valid state.
	_, err := store.GetOAuthState(ctx, "attacker-forged-state-value")
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Fatalf("expected ErrOAuthStateMismatch for unknown state, got %v", err)
	}

	// Error must map to HTTP 400 so the client sees a clear CSRF rejection.
	if !commonerrors.IsBadRequest(err) {
		t.Errorf("state mismatch should have BadRequest code, got %v", err)
	}
}

func TestOAuthState_TamperedState_Rejected(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	realState, _ := GenerateToken()
	if err := store.StoreOAuthState(ctx, realState, &OAuthStateData{Provider: "google"}, time.Minute); err != nil {
		t.Fatalf("StoreOAuthState: %v", err)
	}

	// Any mutation of the state string must fail.
	tamper := func(s string) string {
		// Flip a middle hex digit so the result is guaranteed different from s.
		b := []byte(s)
		mid := len(b) / 2
		if b[mid] == 'a' {
			b[mid] = 'b'
		} else {
			b[mid] = 'a'
		}
		return string(b)
	}
	for _, tampered := range []string{
		realState + "x",
		"x" + realState,
		tamper(realState),
	} {
		if tampered == realState {
			t.Fatalf("tamper helper produced an unchanged string for %q", realState)
		}
		_, err := store.GetOAuthState(ctx, tampered)
		if !errors.Is(err, ErrOAuthStateMismatch) {
			t.Errorf("tampered state %q: want ErrOAuthStateMismatch, got %v", tampered, err)
		}
	}
}

func TestOAuthState_SingleUse_DeleteInvalidates(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	state, _ := GenerateToken()
	if err := store.StoreOAuthState(ctx, state, &OAuthStateData{Provider: "google"}, time.Minute); err != nil {
		t.Fatalf("StoreOAuthState: %v", err)
	}

	// First callback consumes state.
	if _, err := store.GetOAuthState(ctx, state); err != nil {
		t.Fatalf("first Get: %v", err)
	}
	if err := store.DeleteOAuthState(ctx, state); err != nil {
		t.Fatalf("DeleteOAuthState: %v", err)
	}

	// Replay with the same state must fail.
	_, err := store.GetOAuthState(ctx, state)
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Errorf("replayed state: want ErrOAuthStateMismatch, got %v", err)
	}
}

func TestOAuthState_Expired_Rejected(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	state, _ := GenerateToken()
	if err := store.StoreOAuthState(ctx, state, &OAuthStateData{Provider: "google"}, 30*time.Second); err != nil {
		t.Fatalf("StoreOAuthState: %v", err)
	}

	mr.FastForward(31 * time.Second)

	_, err := store.GetOAuthState(ctx, state)
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Errorf("expired state: want ErrOAuthStateMismatch, got %v", err)
	}
}
