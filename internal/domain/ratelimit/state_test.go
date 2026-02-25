package ratelimit

import (
	"testing"
	"time"
)

func TestNewState_DefaultValues(t *testing.T) {
	state := NewState("127.0.0.1")

	if state.Key != "127.0.0.1" {
		t.Fatalf("expected key 127.0.0.1, got %s", state.Key)
	}

	if state.Count != 0 {
		t.Fatalf("expected count 0, got %d", state.Count)
	}

	if state.BlockedAt != 0 {
		t.Fatalf("expected blockedAt 0, got %d", state.BlockedAt)
	}
}

func TestStateIsBlocked(t *testing.T) {
	now := time.Unix(1700000000, 0)

	state := State{Key: "127.0.0.1", Count: 1, BlockedAt: now.Unix()}
	if !state.IsBlocked(now.Add(1*time.Second), 5*time.Second) {
		t.Fatal("expected state to be blocked during block window")
	}

	if state.IsBlocked(now.Add(6*time.Second), 5*time.Second) {
		t.Fatal("expected state to be released after block window")
	}
}