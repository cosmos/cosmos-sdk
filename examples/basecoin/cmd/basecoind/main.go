package main

import "github.com/cosmos/cosmos-sdk/examples/basecoin/app"

func main() {
	// TODO CREATE CLI

	bapp := app.NewBasecoinApp("")
	bapp.RunForever()
}
