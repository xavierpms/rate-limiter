package database

import "testing"

func TestNewTokenLimitList_ParsesValidValues(t *testing.T) {
	tokens := NewTokenLimitList("10,20")

	if tokens.GetLimit("Token10") != 10 {
		t.Fatalf("expected Token10 limit to be 10, got %d", tokens.GetLimit("Token10"))
	}

	if tokens.GetLimit("Token20") != 20 {
		t.Fatalf("expected Token20 limit to be 20, got %d", tokens.GetLimit("Token20"))
	}
}

func TestNewTokenLimitList_IgnoresInvalidValues(t *testing.T) {
	tokens := NewTokenLimitList("10,abc,,  ,30")

	if tokens.GetLimit("Token10") != 10 {
		t.Fatalf("expected Token10 limit to be 10, got %d", tokens.GetLimit("Token10"))
	}

	if tokens.GetLimit("Token30") != 30 {
		t.Fatalf("expected Token30 limit to be 30, got %d", tokens.GetLimit("Token30"))
	}

	if tokens.GetLimit("Tokenabc") != 0 {
		t.Fatalf("expected invalid token to be ignored, got %d", tokens.GetLimit("Tokenabc"))
	}
}
