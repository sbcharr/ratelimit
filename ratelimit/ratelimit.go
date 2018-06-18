package ratelimit

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
)

var (
	incr func(string) error // incr is a thread-safe function to get-and-set value for the key
)

// Custom error messages
var (
	//errServerBusy      = errors.New("server is busy processing other requests, please try again later")
	errTooManyRequests = errors.New("you are sending too many requests, please slow down")
	//errPer             = errors.New("valid value of 'Per' are 'second', 'minute' and 'hour'")
)

// NewRateLimiter returns a new rate limiter
func NewRateLimiter(r *redis.Client, key string, limit int64, timeSlice time.Duration, per string) *RedisRateLimiter {
	return &RedisRateLimiter{
		redis:     r,
		key:       key,
		limit:     limit,
		timeSlice: timeSlice,
		per:       per,
	}
}

// ApplyRateLimit executes run function for the type
func ApplyRateLimit(r RateLimit) error {
	return r.run()
}

// run runs the ratelimiter
func (l *RedisRateLimiter) run() error {
	incr = func(key string) error {
		err := l.redis.Watch(func(tx *redis.Tx) error {
			defer func() {
				tx.Close()
			}()
			n, err := tx.Get(key).Int64()
			if err != nil && err != redis.Nil {
				return err
			}

			if err == nil && n >= l.limit {
				return errTooManyRequests
			}

			_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.Incr(key)
				switch l.per {
				case "second":
					pipe.Expire(key, l.timeSlice*time.Second)
				case "minute":
					pipe.Expire(key, l.timeSlice*time.Minute)
				case "hour":
					pipe.Expire(key, l.timeSlice*time.Hour)
				}

				_, err = pipe.Exec()
				return err
			})
			return err
		}, key)

		if err == redis.TxFailedErr {
			return incr(l.key)
		}
		return err
	}

	// actual execution starts here
	err := incr(l.key)
	return err
}
