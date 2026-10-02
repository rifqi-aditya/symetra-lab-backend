package category

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewCategory(t *testing.T) {
	userID := uuid.New()
	cat, err := NewCategory(userID, "Robotics")
	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
	if cat.Name() != "Robotics" {
		t.Errorf("Expected name Robotics, got %s", cat.Name())
	}
}
