package keys

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/bech32"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var bech32Prefixes = []string{
	sdk.Bech32PrefixAccAddr,
	sdk.Bech32PrefixAccPub,
	sdk.Bech32PrefixValAddr,
	sdk.Bech32PrefixValPub,
	sdk.Bech32PrefixConsAddr,
	sdk.Bech32PrefixConsPub,
}

func parseKeyStringCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "parse <hex-or-bech32-address>",
		Short: "Parse key hex or bech32 address and show relevant information",
		Long: "Convert and print to stdout key addresses and fingerprints from hexadecimal into" +
			" bech32 cosmos prefixed format and vice versa.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if !(runFromBech32(args[0]) || runFromHex(args[0])) {
				return errors.New("couldn't find valid bech32 nor hex data")
			}
			return nil
		},
	}
}

// print info from bech32
func runFromBech32(bech32str string) bool {
	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return false
	}
	fmt.Printf("Human readible part: %v\nBytes (hex): %X\n", hrp, bz)
	return true
}

// print info from hex
func runFromHex(hexstr string) bool {
	bz, err := hex.DecodeString(hexstr)
	if err != nil {
		return false
	}
	fmt.Println("Bech32 formats:")
	for _, prefix := range bech32Prefixes {
		bech32Addr, err := bech32.ConvertAndEncode(prefix, bz)
		if err != nil {
			panic(err)
		}
		fmt.Println("  - " + bech32Addr)
	}
	return true
}
