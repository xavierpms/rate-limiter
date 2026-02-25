package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadFromEnv_SuccessWithDefaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("expected default http addr :8080, got %s", cfg.HTTPAddr)
	}

	if cfg.RedisDB != 0 {
		t.Fatalf("expected default redis db 0, got %d", cfg.RedisDB)
	}

	if cfg.DefaultLimit != 10 {
		t.Fatalf("expected default limit 10, got %d", cfg.DefaultLimit)
	}

	if cfg.CleanupInterval != 100*time.Millisecond {
		t.Fatalf("expected cleanup interval 100ms, got %s", cfg.CleanupInterval)
	}

	if cfg.BlockDuration != 200*time.Millisecond {
		t.Fatalf("expected block duration 200ms, got %s", cfg.BlockDuration)
	}
}

func TestLoadFromEnv_SuccessWithOptionalValues(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("RATELIMIT_HTTP_ADDR", ":9090")
	t.Setenv("RATELIMIT_REDIS_DB", "2")
	t.Setenv("RATELIMIT_REDIS_PASSWORD", "secret")
	t.Setenv("RATELIMIT_TOKEN_LIST", "20,50")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected http addr :9090, got %s", cfg.HTTPAddr)
	}

	if cfg.RedisDB != 2 {
		t.Fatalf("expected redis db 2, got %d", cfg.RedisDB)
	}

	if cfg.RedisPassword != "secret" {
		t.Fatalf("expected redis password secret, got %s", cfg.RedisPassword)
	}

	if cfg.TokenLimits != "20,50" {
		t.Fatalf("expected token limits 20,50, got %s", cfg.TokenLimits)
	}
}

func TestLoadFromEnv_MissingRequiredVariable(t *testing.T) {
	t.Setenv("RATELIMIT", "10")
	t.Setenv("RATELIMIT_CLEANUP_INTERVAL", "100")
	t.Setenv("RATELIMIT_BLOCK_TIME", "200")
	t.Setenv("RATELIMIT_REDIS_URL", "")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for missing required environment variable")
	}

	if !strings.Contains(err.Error(), "missing required environment variable: RATELIMIT_REDIS_URL") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadFromEnv_InvalidRequiredInt(t *testing.T) {
	t.Setenv("RATELIMIT", "invalid")
	t.Setenv("RATELIMIT_CLEANUP_INTERVAL", "100")
	t.Setenv("RATELIMIT_BLOCK_TIME", "200")
	t.Setenv("RATELIMIT_REDIS_URL", "redis:6379")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid RATELIMIT")
	}

	if !strings.Contains(err.Error(), "invalid RATELIMIT") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadFromEnv_InvalidRedisDB(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("RATELIMIT_REDIS_DB", "abc")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid RATELIMIT_REDIS_DB")
	}

	if !strings.Contains(err.Error(), "invalid RATELIMIT_REDIS_DB") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("RATELIMIT", "10")
	t.Setenv("RATELIMIT_CLEANUP_INTERVAL", "100")
	t.Setenv("RATELIMIT_BLOCK_TIME", "200")
	t.Setenv("RATELIMIT_REDIS_URL", "redis:6379")
	t.Setenv("RATELIMIT_HTTP_ADDR", "")
	t.Setenv("RATELIMIT_REDIS_DB", "")
	t.Setenv("RATELIMIT_REDIS_PASSWORD", "")
	t.Setenv("RATELIMIT_TOKEN_LIST", "")
}
