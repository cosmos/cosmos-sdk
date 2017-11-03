package main

import "fmt"

func main() {
	fmt.Println("vim-go")
}

type Account interface {
	Get(key interface{}) (value interface{})
	Address() []byte
	PubKey() crypto.PubKey

	// Serialize the Account to bytes.
	Bytes() []byte
}

type AccountStore interface {
	GetAccount(addr []byte) Account
	SetAccount(acc Account)
}
