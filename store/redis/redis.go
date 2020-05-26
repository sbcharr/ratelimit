package redis

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/xerrors"
	"time"
)

var (
	//errServerBusy      = errors.New("server is busy processing other requests, please try again later")
	errTooManyRequests   = xerrors.New("you are sending too many requests, please slow down")
	errBurstLimitReached = xerrors.New("burst limit exceeded")
)

type counter struct {
	Count       int   `redis:"count"`
	BurstCount  int   `redis:"burst_count"`
	LastUpdated int64 `redis:"last_updated"`
}

// FWRateLimiter represents fixed window rate limiting algorithm
type FWRateLimiter struct {
	limit    int
	timeUnit string
	// ttl in second
	ttl        int
	burstLimit int
	pool       *redis.Pool
}

func (fw *FWRateLimiter) RunContext(ctx context.Context, key string) error {
	// fmt.Println("limit:", fw.limit)
	conn, err := fw.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// fmt.Println("closing connection to the redis cluster")
		_ = conn.Close()
	}()
	// ping the connection
	if _, err = conn.Do("PING"); err != nil {
		return err
	}
	//time.Sleep(1*time.Millisecond)

RACE:
	// watch the key for changes
	if _, err = conn.Do("WATCH", key); err != nil {
		return err
	}

	// get the value of the key
	s, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		return err
	}
	// fmt.Printf("s = %v\n", s)
	if len(s) == 0 {
		// if the key doesn't exist then create a new value
		// fmt.Println("debug2")
		c := counter{
			Count:       1,
			BurstCount:  1,
			LastUpdated: time.Now().Truncate(time.Second).UTC().Unix(),
		}
		// fmt.Println(c.count, c.lastUpdated)
		// fmt.Printf("%s does not exist\n", fw.key)
		// fmt.Println(string(val), c.count, c.lastUpdated)
		if _, err = conn.Do("MULTI"); err != nil {
			return err
		}
		if _, err = conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(&c)...); err != nil {
			return err
		}
		if _, err = conn.Do("EXPIRE", key, fw.ttl); err != nil {
			return err
		}
		//if _, err = conn.Do("EXEC"); err != nil {
		//	return err
		//}
		a, _ := conn.Do("EXEC")
		// fmt.Println("a1 =", a)
		if a == nil {
			time.Sleep(1 * time.Nanosecond)
			goto RACE
		}

		return nil
	}
	// fmt.Println("debug1")
	c1 := counter{}
	if err = redis.ScanStruct(s, &c1); err != nil {
		return err
	}
	// fmt.Printf("%v\t%v\t%v\n", c1.Count, c1.BurstCount, c1.LastUpdated)
	clockTime := time.Now().Truncate(time.Second).UTC().Unix()
	if clockTime > c1.LastUpdated {
		// if clock has moved past the burst reset time then reset the burst counter
		c1.BurstCount = 1
		c1.LastUpdated = clockTime
	} else {
		if c1.BurstCount >= fw.burstLimit {
			return errBurstLimitReached
		} else {
			c1.BurstCount++
		}
	}
	if c1.Count >= fw.limit {
		return errTooManyRequests
	}
	c1.Count++
	if _, err = conn.Do("MULTI"); err != nil {
		return err
	}
	if _, err = conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(&c1)...); err != nil {
		return err
	}
	//if _, err = conn.Do("EXEC"); err != nil {
	//	return err
	//}
	a, _ := conn.Do("EXEC")
	// fmt.Println("a2 =", a)
	if a == nil {
		time.Sleep(1 * time.Nanosecond)
		goto RACE
	}

	return nil
}

// NewFWRateLimiter returns a new rate limiter
func NewFWRateLimiter(maxIdle int, maxActive int, idleTimeout time.Duration, host string, port string, /*key string,*/ limit int, burstLimit int, timeUnit string) (*FWRateLimiter, error) {
	if port == "" {
		port = "6379"
	}
	if idleTimeout <= 0 {
		return nil, xerrors.New("idleTimeout value should be greater than 0")
	}
	var ttl int
	addr := host + ":" + port
	pool := newPool(addr, maxIdle, maxActive, idleTimeout)
	switch timeUnit {
	case "second":
		ttl = 1
	case "minute":
		ttl = 60
	case "hour":
		ttl = 3600
	}

	return &FWRateLimiter{
		limit:      limit,
		timeUnit:   timeUnit,
		ttl:        ttl,
		burstLimit: burstLimit,
		pool:       pool,
	}, nil
}

func newPool(addr string, maxIdle int, maxActive int, idleTimeout time.Duration) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			return conn, err
		},
	}
}
