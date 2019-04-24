package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

// DONTCOVER

// nolint: errcheck
func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%x", h.Sum(nil))
}
