// +build scripts

package main

import (
	"fmt"
	"github.com/tendermint/basecoin/tests"
	"github.com/tendermint/go-wire"
)

/*
PrivKey: 019F86D081884C7D659A2FEAA0C55AD015A3BF4F1B2B0B822CD15D6C15B0F00A0867D3B5EAF0C0BF6B5A602D359DAECC86A7A74053490EC37AE08E71360587C870
PubKey: 0167D3B5EAF0C0BF6B5A602D359DAECC86A7A74053490EC37AE08E71360587C870
Address: D9B727742AA29FA638DC63D70813C976014C4CE0
*/
func main() {
	tAcc := tests.PrivAccountFromSecret("test")
	fmt.Println("PrivKey:", fmt.Sprintf("%X", tAcc.PrivKey.Bytes()))
	fmt.Println("PubKey:", fmt.Sprintf("%X", tAcc.Account.PubKey.Bytes()))
	fmt.Println("Address:", fmt.Sprintf("%X", tAcc.Account.PubKey.Address()))
	fmt.Println(string(wire.JSONBytesPretty(tAcc)))
}
