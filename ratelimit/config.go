package ratelimit

import (
	"crypto/tls"
	"time"

	"github.com/go-redis/redis"
)

// RateLimitClient encapsulates a redis client
type RateLimitClient struct {
	RedisCli *redis.Client
}

// RateLimiter holds various config parameters of the limiter
type RateLimiter struct {
	ConnectOpt *ConnectionOptions

	// Key to be watched for
	KeyPrefix string

	// Limit on requests per second
	Limit int64

	// Slice of time to be monitored for limiting requests
	TimeSlice time.Duration

	// A String representing the unit of time, valid values are
	// "hour", "minute" and "second"
	Per string
}

type ConnectionOptions struct {
	// host:port address.
	Addr string

	// Optional password. Must match the password specified in the
	// requirepass server configuration option.
	Password string
	// Database to be selected after connecting to the server.
	DB int

	// Maximum number of retries before giving up.
	// Default is to not retry failed commands.
	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is 3 seconds.
	ReadTimeout time.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes.
	IdleTimeout time.Duration
	// Frequency of idle checks.
	// Default is 1 minute. -1 disables idle check.
	IdleCheckFrequency time.Duration

	// Enables read only queries on slave nodes.
	readOnly bool

	// TLS Config to use. When set TLS will be negotiated.
	TLSConfig *tls.Config
}
