package main

import (
	"cosmossdk.io/tools/hubl/internal"
)

func main() {
	cmd, err := internal.RootCommand()
	if err != nil {
		panic(err)
	}

	if err = cmd.Execute(); err != nil {
		panic(err)
	}
}
