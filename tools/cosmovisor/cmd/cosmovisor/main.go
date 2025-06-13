package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	// ensure we shutdown if the process is killed and context cancellation doesn't cause an exit
	go func() {
		<-shutdownChan
		fmt.Println("Received shutdown signal, exiting gracefully...")
		// TODO configure this timeout
		time.Sleep(10 * time.Second)
		fmt.Println("Forcing process shutdown")
		os.Exit(0)
	}()
	if err := NewRootCmd().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
