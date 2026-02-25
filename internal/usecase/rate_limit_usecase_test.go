package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/xavierpms/rate-limiter/internal/domain/ratelimit"
	"github.com/xavierpms/rate-limiter/internal/ports"
)

type fakeRepository struct {
	mu   sync.Mutex
	data map[string]ratelimit.State
}

type fakeTokenLimits struct {
	limits map[string]int
}

func (f fakeTokenLimits) LimitFor(token string) int {
	return f.limits[token]
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{data: make(map[string]ratelimit.State)}
}

func (f *fakeRepository) Get(ctx context.Context, key string) (ratelimit.State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	value, ok := f.data[key]
	if !ok {
		return ratelimit.State{}, ports.ErrStateNotFound
	}

	return value, nil
}

func (f *fakeRepository) Save(ctx context.Context, state ratelimit.State) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[state.Key] = state
	return nil
}

func (f *fakeRepository) ListKeys(ctx context.Context) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	keys := make([]string, 0, len(f.data))
	for key := range f.data {
		keys = append(keys, key)
	}

	return keys, nil
}

func (f *fakeRepository) Delete(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, key)
	return nil
}

func (f *fakeRepository) stateByKey(t *testing.T, key string) ratelimit.State {
	t.Helper()

	state, ok := f.data[key]
	if !ok {
		t.Fatalf("expected key %s to exist", key)
	}
	return state
}

func TestAllow_FirstRequestIsAllowed(t *testing.T) {
	repository := newFakeRepository()
	tokens := fakeTokenLimits{limits: map[string]int{}}
	ratelimiter := NewIpRateLimiter(context.Background(), 2, 0, time.Second, tokens, repository)

	allowed := ratelimiter.Allow("127.0.0.1", "")
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}

	state := repository.stateByKey(t, "127.0.0.1")
	if state.Count != 1 {
		t.Fatalf("expected count 1, got %d", state.Count)
	}
}

func TestAllow_BlocksWhenDefaultLimitExceeded(t *testing.T) {
	repository := newFakeRepository()
	tokens := fakeTokenLimits{limits: map[string]int{}}
	ratelimiter := NewIpRateLimiter(context.Background(), 1, 0, time.Second, tokens, repository)

	if !ratelimiter.Allow("127.0.0.1", "") {
		t.Fatal("expected first request to be allowed")
	}

	if ratelimiter.Allow("127.0.0.1", "") {
		t.Fatal("expected second request to be blocked")
	}

	state := repository.stateByKey(t, "127.0.0.1")
	if state.BlockedAt == 0 {
		t.Fatal("expected block timestamp to be set")
	}
}

func TestAllow_UsesTokenLimitWhenPresent(t *testing.T) {
	repository := newFakeRepository()
	tokens := fakeTokenLimits{limits: map[string]int{"Token2": 2}}
	ratelimiter := NewIpRateLimiter(context.Background(), 1, 0, time.Second, tokens, repository)

	if !ratelimiter.Allow("127.0.0.1", "Token2") {
		t.Fatal("expected first token request to be allowed")
	}

	if !ratelimiter.Allow("127.0.0.1", "Token2") {
		t.Fatal("expected second token request to be allowed")
	}

	if ratelimiter.Allow("127.0.0.1", "Token2") {
		t.Fatal("expected third token request to be blocked")
	}
}

func TestAllow_ReleasesAfterBlockDuration(t *testing.T) {
	repository := newFakeRepository()
	tokens := fakeTokenLimits{limits: map[string]int{}}
	ratelimiter := NewIpRateLimiter(context.Background(), 1, 0, 2*time.Second, tokens, repository)

	now := time.Unix(1700000000, 0)
	ratelimiter.now = func() time.Time { return now }

	if !ratelimiter.Allow("127.0.0.1", "") {
		t.Fatal("expected first request to be allowed")
	}
	if ratelimiter.Allow("127.0.0.1", "") {
		t.Fatal("expected second request to be blocked")
	}

	now = now.Add(3 * time.Second)
	if !ratelimiter.Allow("127.0.0.1", "") {
		t.Fatal("expected request after block duration to be allowed")
	}
}

func TestAllow_ReturnsFalseOnRepositoryFailure(t *testing.T) {
	repository := &failingRepository{}
	tokens := fakeTokenLimits{limits: map[string]int{}}
	ratelimiter := NewIpRateLimiter(context.Background(), 1, 0, time.Second, tokens, repository)

	if ratelimiter.Allow("127.0.0.1", "") {
		t.Fatal("expected request to fail when repository is unavailable")
	}
}

type failingRepository struct{}

func (f *failingRepository) Get(ctx context.Context, key string) (ratelimit.State, error) {
	return ratelimit.State{}, errors.New("repository error")
}

func (f *failingRepository) Save(ctx context.Context, state ratelimit.State) error {
	return errors.New("repository error")
}

func (f *failingRepository) ListKeys(ctx context.Context) ([]string, error) {
	return nil, errors.New("repository error")
}

func (f *failingRepository) Delete(ctx context.Context, key string) error {
	return errors.New("repository error")
}
