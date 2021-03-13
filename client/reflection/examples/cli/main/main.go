package main

import (
	"github.com/cosmos/cosmos-sdk/client/reflection"
	"github.com/cosmos/cosmos-sdk/client/reflection/examples/cli"
	"log"
)

func main() {
	c, err := reflection.NewClient("localhost:9090", "tcp://localhost:26657", nil)
	if err != nil {
		panic(err)
	}

	prompt := cli.NewCLI(c)
	err = prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
}
