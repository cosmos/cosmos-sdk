package main

import (
	"cosmossdk.io/tools/cosmcli/internal"
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
