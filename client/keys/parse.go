package keys

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/bech32"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
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

type hexOutput struct {
	Human string `json:"human"`
	Bytes string `json:"bytes"`
}

func (ho hexOutput) String() string {
	return fmt.Sprintf("Human readable part: %v\nBytes (hex): %s", ho.Human, ho.Bytes)
}

func newHexOutput(human string, bs []byte) hexOutput {
	return hexOutput{Human: human, Bytes: fmt.Sprintf("%X", bs)}
}

type bech32Output struct {
	Formats []string `json:"formats"`
}

func newBech32Output(bs []byte) bech32Output {
	out := bech32Output{Formats: make([]string, len(bech32Prefixes))}
	for i, prefix := range bech32Prefixes {
		bech32Addr, err := bech32.ConvertAndEncode(prefix, bs)
		if err != nil {
			panic(err)
		}
		out.Formats[i] = bech32Addr
	}

	return out
}

func (bo bech32Output) String() string {
	out := make([]string, len(bo.Formats))

	for i, format := range bo.Formats {
		out[i] = fmt.Sprintf("  - %s", format)
	}

	return fmt.Sprintf("Bech32 Formats:\n%s", strings.Join(out, "\n"))
}

func parseKeyStringCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse <hex-or-bech32-address>",
		Short: "Parse address from hex to bech32 and vice versa",
		Long: `Convert and print to stdout key addresses and fingerprints from
hexadecimal into bech32 cosmos prefixed format and vice versa.
`,
		Args: cobra.ExactArgs(1),
		RunE: parseKey,
	}
	cmd.Flags().Bool(client.FlagIndentResponse, false, "Indent JSON output")

	return cmd
}

func parseKey(_ *cobra.Command, args []string) error {
	addr := strings.TrimSpace(args[0])
	if len(addr) == 0 {
		return errors.New("couldn't parse empty input")
	}
	if !(runFromBech32(addr) || runFromHex(addr)) {
		return errors.New("couldn't find valid bech32 nor hex data")
	}
	return nil
}

// print info from bech32
func runFromBech32(bech32str string) bool {
	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return false
	}
	displayParseKeyInfo(newHexOutput(hrp, bz))
	return true
}

// print info from hex
func runFromHex(hexstr string) bool {
	bz, err := hex.DecodeString(hexstr)
	if err != nil {
		return false
	}
	displayParseKeyInfo(newBech32Output(bz))
	return true
}

func displayParseKeyInfo(stringer fmt.Stringer) {
	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		fmt.Printf("%s\n", stringer)

	case OutputFormatJSON:
		var out []byte
		var err error

		if viper.GetBool(client.FlagIndentResponse) {
			out, err = cdc.MarshalJSONIndent(stringer, "", "  ")
			if err != nil {
				panic(err)
			}
		} else {
			out = cdc.MustMarshalJSON(stringer)
		}

		fmt.Println(string(out))
	}
}
