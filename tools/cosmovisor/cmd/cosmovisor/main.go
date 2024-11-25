package main

import (
	"context"
	"os"
)

func main() {
	if err := NewRootCmd().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}
