package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec" // XXX fix
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

// GetCmdQuerySigningInfo implements the command to query signing info.
func GetCmdQuerySigningInfo(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-info [validator-pubkey]",
		Short: "Query a validator's signing information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pk, err := sdk.GetConsPubKeyBech32(args[0])
			if err != nil {
				return err
			}

			key := slashing.GetValidatorSigningInfoKey(sdk.ConsAddress(pk.Address()))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			signingInfo := new(slashing.ValidatorSigningInfo)
			cdc.MustUnmarshalBinaryLengthPrefixed(res, signingInfo)

			switch viper.Get(cli.OutputFlag) {

			case "text":
				human := signingInfo.HumanReadableString()
				fmt.Println(human)

			case "json":
				// parse out the signing info
				output, err := codec.MarshalJSONIndent(cdc, signingInfo)
				if err != nil {
					return err
				}
				fmt.Println(string(output))
			}

			return nil
		},
	}

	return cmd
}
