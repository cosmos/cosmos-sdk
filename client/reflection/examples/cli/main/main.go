package main

import (
	"context"
	"log"

	"github.com/cosmos/cosmos-sdk/client/reflection/client"
	"github.com/cosmos/cosmos-sdk/client/reflection/examples/cli"
)

func main() {
	c, err := client.Dial(context.TODO(), "localhost:9090", "tcp://localhost:26657", nil)
	if err != nil {
		panic(err)
	}

	prompt := cli.NewCLI(c)
	err = prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
}
