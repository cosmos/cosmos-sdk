package main

import (
	"fmt"

	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
)

func main() {
	// This must be compiled beforehand and given to dredd as parameter, in the meantime the server should be running
	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))
	h.BeforeAll(func(t []*trans.Transaction) {
		fmt.Println("before all modification")
	})
	h.BeforeEach(func(t *trans.Transaction) {
		fmt.Println("before each modification")
	})
	h.Before("/node_version > GET", func(t *trans.Transaction) {
		fmt.Println("before modification")
	})
	h.BeforeEachValidation(func(t *trans.Transaction) {
		fmt.Println("before each validation modification")
	})
	h.BeforeValidation("/node_version > GET", func(t *trans.Transaction) {
		fmt.Println("before validation modification")
	})
	h.After("/node_version > GET", func(t *trans.Transaction) {
		fmt.Println("after modification")
	})
	h.AfterEach(func(t *trans.Transaction) {
		fmt.Println("after each modification")
	})
	h.AfterAll(func(t []*trans.Transaction) {
		fmt.Println("after all modification")
	})
	server.Serve()
	defer server.Listener.Close()
}
