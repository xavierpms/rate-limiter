package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeLimiter struct {
	allow      bool
	receivedIP string
	receivedTK string
}

func (f *fakeLimiter) Allow(ip, token string) bool {
	f.receivedIP = ip
	f.receivedTK = token
	return f.allow
}

func TestNewHTTPHandler_Returns200WhenLimiterAllows(t *testing.T) {
	limiter := &fakeLimiter{allow: true}
	handler := NewHTTPHandler(limiter)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("API_KEY", "Token20")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if !strings.Contains(string(body), "Hello World") {
		t.Fatalf("expected response body to contain Hello World, got %q", string(body))
	}

	if limiter.receivedIP != "127.0.0.1" {
		t.Fatalf("expected limiter to receive ip 127.0.0.1, got %q", limiter.receivedIP)
	}

	if limiter.receivedTK != "Token20" {
		t.Fatalf("expected limiter to receive token Token20, got %q", limiter.receivedTK)
	}
}

func TestNewHTTPHandler_Returns429WhenLimiterBlocks(t *testing.T) {
	limiter := &fakeLimiter{allow: false}
	handler := NewHTTPHandler(limiter)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}
}
