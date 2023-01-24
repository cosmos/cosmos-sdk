package main

import (
	"cosmossdk.io/tools/hubl/internal"
)

func main() {
	cmd, err := internal.RootCommand()
	if err != nil {
		panic(err)
	}

	err = cmd.Execute()
	if err != nil {
		panic(err)
	}
}
