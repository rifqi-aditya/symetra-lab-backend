package oauthstate_test

import (
	"testing"
	"time"

	"symetra-lab-backend-v2/pkg/oauthstate"
)

func TestOAuthStateManager(t *testing.T) {
	mgr := oauthstate.NewManager(100 * time.Millisecond)

	t.Run("should generate and successfully consume valid state", func(t *testing.T) {
		st, err := mgr.GenerateState()
		if err != nil {
			t.Fatalf("failed to generate state: %v", err)
		}
		if len(st) != 64 {
			t.Fatalf("expected 64 hex chars, got %d", len(st))
		}

		// First consume should succeed
		if !mgr.ValidateAndConsume(st) {
			t.Fatalf("expected state to be valid")
		}

		// Second consume should fail (one-time use)
		if mgr.ValidateAndConsume(st) {
			t.Fatalf("expected state to be consumed already")
		}
	})

	t.Run("should reject nonexistent or empty state", func(t *testing.T) {
		if mgr.ValidateAndConsume("") {
			t.Fatalf("expected empty state to be rejected")
		}
		if mgr.ValidateAndConsume("nonexistent_state_value") {
			t.Fatalf("expected invalid state to be rejected")
		}
	})

	t.Run("should reject expired state", func(t *testing.T) {
		shortMgr := oauthstate.NewManager(5 * time.Millisecond)
		st, err := shortMgr.GenerateState()
		if err != nil {
			t.Fatalf("failed to generate state: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		if shortMgr.ValidateAndConsume(st) {
			t.Fatalf("expected expired state to be rejected")
		}
	})
}

