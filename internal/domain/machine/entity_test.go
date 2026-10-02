package machine

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewMachine_Success(t *testing.T) {
	userID := uuid.New()
	brand := "Voron Design"
	m, err := NewMachine(
		userID,
		"Voron 2.4",
		&brand,
		12000000,
		10000,
		350,
	)

	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
	if m.Name() != "Voron 2.4" {
		t.Errorf("Expected name Voron 2.4, got %s", m.Name())
	}
	if m.CurrentState() != "IDLE" {
		t.Errorf("Expected state IDLE, got %s", m.CurrentState())
	}
}
