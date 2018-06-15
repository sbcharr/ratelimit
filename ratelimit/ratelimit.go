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
func NewRateLimiter(r *redis.Client, key string, limit int64, timeSlice time.Duration, per string) RateLimiter {
	return RateLimiter{
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
func (rate RateLimiter) run() error {
	incr = func(key string) error {
		err := rate.redis.Watch(func(tx *redis.Tx) error {
			defer func() {
				tx.Close()
			}()
			n, err := tx.Get(key).Int64()
			if err != nil && err != redis.Nil {
				return err
			}

			if err == nil && n >= rate.limit {
				return errTooManyRequests
			}

			_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.Incr(key)
				switch rate.per {
				case "second":
					pipe.Expire(key, rate.timeSlice*time.Second)
				case "minute":
					pipe.Expire(key, rate.timeSlice*time.Minute)
				case "hour":
					pipe.Expire(key, rate.timeSlice*time.Hour)
				}

				_, err = pipe.Exec()
				return err
			})
			return err
		}, key)

		if err == redis.TxFailedErr {
			return incr(rate.key)
		}
		return err
	}

	// actual execution starts here
	err := incr(rate.key)
	return err
}
