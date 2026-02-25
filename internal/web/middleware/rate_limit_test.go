package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type spyLimiter struct {
	allow bool
	ip    string
	token string
}

func (s *spyLimiter) Allow(ip, token string) bool {
	s.ip = ip
	s.token = token
	return s.allow
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	limiter := &spyLimiter{allow: true}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := RateLimitMiddleware(next, limiter)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("API_KEY", "Token20")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if limiter.ip != "127.0.0.1" {
		t.Fatalf("expected ip 127.0.0.1, got %s", limiter.ip)
	}

	if limiter.token != "Token20" {
		t.Fatalf("expected token Token20, got %s", limiter.token)
	}
}

func TestRateLimitMiddleware_Blocked(t *testing.T) {
	limiter := &spyLimiter{allow: false}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called when blocked")
	})

	handler := RateLimitMiddleware(next, limiter)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_InvalidRemoteAddr(t *testing.T) {
	limiter := &spyLimiter{allow: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called on invalid remote address")
	})

	handler := RateLimitMiddleware(next, limiter)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.RemoteAddr = "invalid-address"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}