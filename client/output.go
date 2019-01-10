package client

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

// Printable defines which structs can be printed by
// CLI output functions
type Printable interface {
	HumanReadableString() string
}

// PrintOutput prints output while respecting output and indent flags
// NOTE: pass in marshalled structs that have been unmarshalled
func PrintOutput(cdc *codec.Codec, toPrint Printable) {
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Println(toPrint.HumanReadableString())
	case "json":
		if viper.GetBool(FlagIndentResponse) {
			out, err := codec.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(out))
		} else {
			fmt.Println(string(cdc.MustMarshalJSON(toPrint)))
		}
	}
}
