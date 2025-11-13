package main

import (
	"os"
)

func main() {
	dir, err := os.MkdirTemp("", "bankspeedtest-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	if err := NewBankSpeedTest(dir).Execute(); err != nil {
		panic(err)
	}
}
