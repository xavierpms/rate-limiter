package database

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/xavierpms/rate-limiter/internal/domain/ratelimit"
	"github.com/xavierpms/rate-limiter/internal/ports"
)

func TestNewRedisClientAndCRUD(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	ctx := context.Background()
	client, err := NewRedisClient(ctx, mr.Addr(), "", 0)
	if err != nil {
		t.Fatalf("expected client creation success, got %v", err)
	}

	if err := client.Set(ctx, "k1", []byte("v1")); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	value, err := client.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if value != "v1" {
		t.Fatalf("expected v1, got %s", value)
	}

	keys, err := client.Keys(ctx, "k*")
	if err != nil {
		t.Fatalf("keys failed: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}

	if err := client.Del(ctx, "k1"); err != nil {
		t.Fatalf("del failed: %v", err)
	}

	_, err = client.Get(ctx, "k1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestNewRedisClient_InvalidAddress(t *testing.T) {
	_, err := NewRedisClient(context.Background(), "127.0.0.1:0", "", 0)
	if err == nil {
		t.Fatal("expected error for invalid redis address")
	}
}

func TestRedisRateLimitRepository(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	ctx := context.Background()
	client, err := NewRedisClient(ctx, mr.Addr(), "", 0)
	if err != nil {
		t.Fatalf("expected client creation success, got %v", err)
	}

	repo := NewRedisRateLimitRepository(&client)
	state := ratelimit.State{Key: "127.0.0.1", Count: 3, BlockedAt: 1700000000}

	if err := repo.Save(ctx, state); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := repo.Get(ctx, state.Key)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if loaded != state {
		t.Fatalf("expected state %#v, got %#v", state, loaded)
	}

	keys, err := repo.ListKeys(ctx)
	if err != nil {
		t.Fatalf("list keys failed: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}

	if err := repo.Delete(ctx, state.Key); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = repo.Get(ctx, state.Key)
	if !errors.Is(err, ports.ErrStateNotFound) {
		t.Fatalf("expected ports.ErrStateNotFound, got %v", err)
	}
}
