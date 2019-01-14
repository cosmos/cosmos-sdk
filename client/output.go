package client

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

// PrintOutput prints output while respecting output and indent flags
// NOTE: pass in marshalled structs that have been unmarshaled
// because this function doesn't return errors for marshaling
func PrintOutput(cdc *codec.Codec, toPrint fmt.Stringer) {
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Println(toPrint.String())
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
