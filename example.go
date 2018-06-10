package main

import (
	"sync"

	rl "github.com/squeakysimple/ratelimit/ratelimit"
)

var (
	address         = "127.0.0.1:6379"
	password        = ""
	database        = 0
	keyPrefix       = "squeakysimple"
	limit     int64 = 10
	per             = "second"
	mutex           = &sync.Mutex{}
)

var wg sync.WaitGroup

func main() {
	options := &rl.ConnectionOptions{
		Addr:     address,
		Password: password,
		DB:       database,
	}

	rateLimit := rl.NewRateLimiter(options, keyPrefix, limit, per)
	client, err := rateLimit.NewRateLimiterClient()
	if err != nil {
		panic(err)
	}

	defer func() {
		err = client.CloseRateLimiterClient()
		if err != nil {
			panic(err)
		}
	}()

	//u := time.Now()
	//v := time.Now().Add(time.Second)
	//for v.After(u) { // Just trying to mimic multi-threaded situation, please write in your own way
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := rateLimit.Run(client)
			if err != nil {
				panic(err)
			}
		}()
		//u = time.Now()
	}
	wg.Wait()
}
