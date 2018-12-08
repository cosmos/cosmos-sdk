package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/tendermint/tendermint/libs/bech32"
)

var bech32Prefixes = []string{"cosmos", "cosmospub", "cosmosvaloper", "cosmosvaloperpub", "cosmosvalcons", "cosmosvalconspub"}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Must specify an input string")
	}
	arg := os.Args[1]
	runFromBech32(arg)
	runFromHex(arg)
}

// Print info from bech32.
func runFromBech32(bech32str string) {
	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		fmt.Println("Not a valid bech32 string")
		return
	}
	fmt.Println("Bech32 parse:")
	fmt.Printf("Human readible part: %v\nBytes (hex): %X\n",
		hrp,
		bz,
	)
}

func runFromHex(hexaddr string) {
	bz, err := hex.DecodeString(hexaddr)
	if err != nil {
		fmt.Println("Not a valid hex string")
		return
	}
	fmt.Println("Hex parse:")
	fmt.Println("Bech32 formats:")
	for _, prefix := range bech32Prefixes {
		bech32Addr, err := bech32.ConvertAndEncode(prefix, bz)
		if err != nil {
			panic(err)
		}
		fmt.Println("  - " + bech32Addr)
	}
}
