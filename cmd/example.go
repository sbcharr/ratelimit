package main

import (
	"fmt"
	"github.com/sbcharr/ratelimit/apiserver"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sbcharr/ratelimit/store/redis"
)

var (
	pidFile     string
	rateLimiter *redis.FWRateLimiter
	limitErr    error
)

const (
	rateLimit   int           = 180
	timeUnit    string        = "minute"
	burstLimit  int           = 20
	maxIdle     int           = 100
	maxActive   int           = 2500
	idleTimeout time.Duration = 60 * time.Second
	host        string        = "127.0.0.1"
	port        string        = ""
)

func init() {
	pidFile = os.Getenv("RATELIMIT_PIDFILE") // "/var/run/user/1000/ratelimit.pid"
}

func init() {
	rateLimiter, limitErr = redis.NewFWRateLimiter(maxIdle, maxActive, idleTimeout, host, port, rateLimit, burstLimit, timeUnit)
	if limitErr != nil {
		panic(limitErr)
	}
}

func writePidFile() {
	if pidFile != "" {
		pid := os.Getpid()
		fmt.Println(fmt.Sprintf(`main.writePidFile() Writing pid %d to %s`, pid, pidFile))
		err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
		if err != nil {
			fmt.Println(fmt.Sprintf(`main.writePidFile() Error while writing pid '%d' to '%s' :: %s`, pid, pidFile, err))
			os.Exit(1)
		}
	} else {
		log.Fatal("no pid file has been supplied, exiting...")
	}
}

func signalHandler() {
	var err error
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	for sig := range ch {
		fmt.Println(fmt.Sprintf("main.signalHandler() Received signal %v, shutting down gracefully...", sig))
		if _, err = os.Stat(pidFile); err == nil {
			if err = os.Remove(pidFile); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
		os.Exit(0)
	}
}

func main() {
	go writePidFile()
	go apiserver.WebAppAPIServer(rateLimiter)
	signalHandler()
	// fmt.Println("process completed")
}
