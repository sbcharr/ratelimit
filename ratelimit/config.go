package ratelimit

import (
	"time"

	"github.com/go-redis/redis"
)

// RateLimit is an interface to for RateLimiter operations
type RateLimit interface {
	run() error
}

// RedisRateLimiter holds various config parameters of the limiter
type RedisRateLimiter struct {
	// A redis client based on go-redis/redis
	redis *redis.Client

	// Key to be watched for
	key string

	// Limit on requests per second
	limit int64

	// Slice of time to be monitored for limiting requests
	timeSlice time.Duration

	// A String representing the unit of time, valid values are
	// "hour", "minute" and "second"
	per string
}
