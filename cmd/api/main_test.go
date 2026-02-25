package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xavierpms/rate-limiter/internal/config"
)

func TestRun_ReturnsLoaderError(t *testing.T) {
	originalLoader := loadConfig
	originalRunner := runApplication
	t.Cleanup(func() {
		loadConfig = originalLoader
		runApplication = originalRunner
	})

	expectedErr := errors.New("load failed")
	loadConfig = func() (config.Config, error) {
		return config.Config{}, expectedErr
	}

	runApplication = func(ctx context.Context, cfg config.Config) error {
		t.Fatal("runApplication should not be called when loadConfig fails")
		return nil
	}

	err := run()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestRun_ReturnsRunnerError(t *testing.T) {
	originalLoader := loadConfig
	originalRunner := runApplication
	t.Cleanup(func() {
		loadConfig = originalLoader
		runApplication = originalRunner
	})

	loadedConfig := config.Config{HTTPAddr: ":8080", DefaultLimit: 10}
	loadConfig = func() (config.Config, error) {
		return loadedConfig, nil
	}

	expectedErr := errors.New("runner failed")
	runApplication = func(ctx context.Context, cfg config.Config) error {
		if cfg != loadedConfig {
			t.Fatalf("expected config %#v, got %#v", loadedConfig, cfg)
		}
		if _, ok := ctx.Deadline(); ok {
			t.Fatal("did not expect deadline on background context")
		}
		return expectedErr
	}

	err := run()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestRun_Success(t *testing.T) {
	originalLoader := loadConfig
	originalRunner := runApplication
	t.Cleanup(func() {
		loadConfig = originalLoader
		runApplication = originalRunner
	})

	loadedConfig := config.Config{
		HTTPAddr:        ":9090",
		DefaultLimit:    20,
		CleanupInterval: 100 * time.Millisecond,
		BlockDuration:   2 * time.Second,
	}

	loadConfig = func() (config.Config, error) {
		return loadedConfig, nil
	}

	called := false
	runApplication = func(ctx context.Context, cfg config.Config) error {
		called = true
		if cfg != loadedConfig {
			t.Fatalf("expected config %#v, got %#v", loadedConfig, cfg)
		}
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
		return nil
	}

	err := run()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !called {
		t.Fatal("expected runApplication to be called")
	}
}