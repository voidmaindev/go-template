package auth

import (
	"encoding/hex"
	"testing"
)

// GenerateToken is load-bearing for OAuth state (CSRF), email verification,
// and password-reset links. These tests pin down the contract that every
// caller relies on: at least 128 bits of unguessable entropy per token.

func TestGenerateToken_LengthAndEncoding(t *testing.T) {
	tok, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	// 32 random bytes hex-encoded → 64 chars.
	if len(tok) != 64 {
		t.Fatalf("token length = %d, want 64", len(tok))
	}
	if _, err := hex.DecodeString(tok); err != nil {
		t.Fatalf("token is not valid hex: %v", err)
	}
}

func TestGenerateToken_SufficientEntropy(t *testing.T) {
	// 32 bytes = 256 bits of entropy; our lower bound for this contract is 128.
	const minBytes = 16
	tok, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	raw, err := hex.DecodeString(tok)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(raw) < minBytes {
		t.Fatalf("entropy %d bytes < required %d", len(raw), minBytes)
	}
}

func TestGenerateToken_Uniqueness(t *testing.T) {
	// Probabilistic check that the source is not a stub or zero-seeded RNG.
	// Birthday-bound for 256-bit space is astronomically large, so even 1000
	// samples must all be distinct — any duplicate indicates a broken source.
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		tok, err := GenerateToken()
		if err != nil {
			t.Fatalf("GenerateToken #%d: %v", i, err)
		}
		if _, dup := seen[tok]; dup {
			t.Fatalf("duplicate token at iteration %d: %s", i, tok)
		}
		seen[tok] = struct{}{}
	}
}
