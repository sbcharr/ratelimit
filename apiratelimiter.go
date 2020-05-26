package ratelimit

import "context"

type APIRateLimiter interface {
	RunContext(ctx context.Context, key string) error
}
