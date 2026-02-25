package ratelimit

import "time"

type State struct {
	Key       string `json:"key"`
	Count     int    `json:"count"`
	BlockedAt int64  `json:"blocked_at"`
}

func NewState(key string) State {
	return State{Key: key, Count: 0, BlockedAt: 0}
}

func (s State) IsBlocked(now time.Time, blockDuration time.Duration) bool {
	if s.BlockedAt <= 0 {
		return false
	}
	releaseAt := time.Unix(s.BlockedAt, 0).Add(blockDuration)
	return now.Before(releaseAt)
}
