package usecase

import (
	"context"
	"log"
	"time"

	"github.com/xavierpms/rate-limiter/internal/domain/ratelimit"
	"github.com/xavierpms/rate-limiter/internal/ports"
)

type RateLimiter struct {
	ctx             context.Context
	defaultLimit    int
	cleanupInterval time.Duration
	blockDuration   time.Duration
	tokenLimits     ports.TokenLimitProvider
	store           ports.RateLimitStateRepository
	now             func() time.Time
}

func NewIpRateLimiter(ctx context.Context, limit int, interval time.Duration, blockInterval time.Duration, listTokens ports.TokenLimitProvider, store ports.RateLimitStateRepository) *RateLimiter {
	rl := &RateLimiter{
		ctx:             ctx,
		defaultLimit:    limit,
		cleanupInterval: interval,
		blockDuration:   blockInterval,
		tokenLimits:     listTokens,
		store:           store,
		now:             time.Now,
	}
	if interval > 0 {
		go rl.cleanupLoop()
	}
	return rl
}

func (rl *RateLimiter) Allow(ip, token string) bool {
	req, err := rl.loadRequest(ip)
	if err != nil {
		log.Println("failed to load rate-limit state:", err)
		return false
	}

	if rl.isBlocked(req) {
		return false
	}

	limit := rl.defaultLimit
	if tokenLimit := rl.tokenLimits.LimitFor(token); tokenLimit > 0 {
		limit = tokenLimit
	}

	if req.Count >= limit {
		if err := rl.saveRequest(ratelimit.State{Key: ip, Count: 0, BlockedAt: rl.now().Unix()}); err != nil {
			log.Println("failed to persist blocked state:", err)
		}
		return false
	}

	if err := rl.saveRequest(ratelimit.State{Key: ip, Count: req.Count + 1, BlockedAt: 0}); err != nil {
		log.Println("failed to persist request state:", err)
		return false
	}

	return true
}

func (rl *RateLimiter) loadRequest(ip string) (ratelimit.State, error) {
	state, err := rl.store.Get(rl.ctx, ip)
	if err == ports.ErrStateNotFound {
		initial := ratelimit.NewState(ip)
		if saveErr := rl.saveRequest(initial); saveErr != nil {
			return ratelimit.State{}, saveErr
		}
		return initial, nil
	}
	if err != nil {
		return ratelimit.State{}, err
	}

	return state, nil
}

func (rl *RateLimiter) saveRequest(state ratelimit.State) error {
	return rl.store.Save(rl.ctx, state)
}

func (rl *RateLimiter) isBlocked(state ratelimit.State) bool {
	return state.IsBlocked(rl.now(), rl.blockDuration)
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			keys, err := rl.store.ListKeys(rl.ctx)
			if err != nil {
				log.Println("cleanup failed to list keys:", err)
				continue
			}
			for _, key := range keys {
				if err := rl.store.Delete(rl.ctx, key); err != nil {
					log.Println("cleanup failed to delete key:", err)
				}
			}
		}
	}
}
