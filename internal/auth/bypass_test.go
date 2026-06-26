package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mgmt-glasses/moe-manager/internal/auth"
)

func TestBypassVerifierUsesTokenAsUID(t *testing.T) {
	user, err := auth.BypassVerifier{}.VerifyIDToken(context.Background(), "u_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UID != "u_001" {
		t.Fatalf("expected UID u_001, got %q", user.UID)
	}
}

func TestBypassVerifierTrimsToken(t *testing.T) {
	user, err := auth.BypassVerifier{}.VerifyIDToken(context.Background(), "  Bearer-less-raw  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UID != "Bearer-less-raw" {
		t.Fatalf("expected trimmed UID, got %q", user.UID)
	}
}

func TestBypassVerifierRejectsEmptyToken(t *testing.T) {
	for _, raw := range []string{"", "   "} {
		_, err := auth.BypassVerifier{}.VerifyIDToken(context.Background(), raw)
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken for %q, got %v", raw, err)
		}
	}
}
