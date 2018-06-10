package ratelimit

import (
	"errors"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

var (
	mutex = &sync.Mutex{}
	incr  func(string) error // incr is a thread-safe function to get-and-set value for the key
)

// Custom error messages
var (
	errServerBusy      = errors.New("server is busy processing other requests, please try again later")
	errTooManyRequests = errors.New("you are sending too many requests, please slow down")
	errPer             = errors.New("valid value of 'Per' are 'second', 'minute' and 'hour'")
)

// NewRateLimiter returns a new rate limiter
func NewRateLimiter(c *ConnectionOptions, keyPrefix string, limit int64, per string) *RateLimiter {
	return &RateLimiter{
		ConnectOpt: c,
		KeyPrefix:  keyPrefix,
		Limit:      limit,
		//TimeSlice:  timeSlice,
		Per: per,
	}
}

// NewRateLimiterClient creates a new redis client, use this client across threads and dont close it for every call
func (r *RateLimiter) NewRateLimiterClient() (*RateLimitClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     r.ConnectOpt.Addr,
		Password: r.ConnectOpt.Password,
		DB:       r.ConnectOpt.DB,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	cli := RateLimitClient{RedisCli: client}

	return &cli, nil
}

// CloseRateLimiterClient closes the client, releasing any open resources.
func (client *RateLimitClient) CloseRateLimiterClient() error {
	err := client.RedisCli.Close()
	if err != nil {
		return err
	}

	return nil
}

// Run runs the ratelimiter
func (r *RateLimiter) Run(client *RateLimitClient) error {
	var ts int
	var dur float64

	switch r.Per {
	case "second":
		ts = time.Now().Second()
	case "minute":
		ts = time.Now().Minute()
	case "hour":
		ts = time.Now().Hour()
	default:
		return errPer
	}

	// Key is a combination of actual key and second/minute/hour number
	key := r.KeyPrefix + ":" + strconv.Itoa(ts)

	incr = func(key string) error {
		switch r.Per {
		case "second":
			dur = math.Abs(float64(time.Now().Second() - ts))
		case "minute":
			dur = math.Abs(float64(time.Now().Minute() - ts))
		case "hour":
			dur = math.Abs(float64(time.Now().Hour() - ts))
		}
		if dur > 0 {
			return errServerBusy
		}
		err := client.RedisCli.Watch(func(tx *redis.Tx) error {
			defer func() {
				tx.Close()
			}()
			n, err := tx.Get(key).Int64()
			if err != nil && err != redis.Nil {
				return err
			}

			if err == nil && n >= r.Limit {
				return errTooManyRequests
			}

			_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.Incr(key)
				switch r.Per {
				case "second":
					//pipe.Expire(key, r.TimeSlice*time.Second)
					pipe.Expire(key, time.Second)
				case "minute":
					pipe.Expire(key, time.Minute)
				case "hour":
					pipe.Expire(key, time.Hour)
				}

				_, err = pipe.Exec()
				return err
			})
			return err
		}, key)

		if err == redis.TxFailedErr {
			return incr(key)
		}
		return err
	}
	// actual execution starts here
	err := incr(key)
	return err
}
