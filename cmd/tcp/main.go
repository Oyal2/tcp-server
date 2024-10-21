package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Oyal2/tcp-server/internal/constant"
	"github.com/Oyal2/tcp-server/internal/server"
	"github.com/Oyal2/tcp-server/pkg/executor"
	"github.com/Oyal2/tcp-server/pkg/ratelimit"
)

func main() {
	// Create a signal chan to gracefully shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	// Create an Executor
	executor := executor.NewCommandExecutor()
	// Create a ratelimiter
	rateLimiter, err := ratelimit.NewIPRateLimiter(constant.DefaultRateLimit, constant.DefaultRateInterval)
	if err != nil {
		log.Fatal(err.Error())
	}
	// Create a tcp server
	params := server.TCPServerParams{
		Port:         3000,
		ReadTimeout:  constant.DefaultReadTimeout,
		WriteTimeout: constant.DefaultWriteTimeout,
		Executor:     executor,
		WaitGroup:    &sync.WaitGroup{},
		RateLimiter:  rateLimiter,
	}
	server, err := server.NewTCPServer(params)
	if err != nil {
		log.Fatalf("cannot create server: %s", err)
	}

	// Run the server
	go server.Start(ctx)

	// Watch for Ctrl+C cancel
	<-sig
	// Clean up
	cancel()
	server.Stop()
}
