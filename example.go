package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	rl "github.com/squeakysimple/ratelimit/ratelimit"
)

var (
	address         = "127.0.0.1:6379"
	password        = ""
	database        = 0
	keyPrefix       = "squeakysimple"
	limit     int64 = 10
	timeSlice       = 1 * time.Second
	per             = "second"
	mutex           = &sync.Mutex{}
)

var wg sync.WaitGroup

// NewRedisClient returns a redis client with supplied option
func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       database,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func main() {
	client, redisErr := NewRedisClient()
	if redisErr != nil {
		panic(redisErr)
	}
	defer func() {
		err := client.Close()
		if err != nil {
			panic(err)
		}
	}()

	rateLimiter := rl.NewRateLimiter(client, keyPrefix, limit, timeSlice, per)

	//u := time.Now()
	//v := time.Now().Add(time.Second)
	//for v.After(u) { // Just trying to mimic multi-threaded situation, please write in your own way
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := rl.ApplyRateLimit(rateLimiter)
			if err != nil {
				fmt.Println(err)
			}
		}()
		//u = time.Now()
	}
	wg.Wait()
}
