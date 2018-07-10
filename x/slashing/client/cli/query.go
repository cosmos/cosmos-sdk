package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire" // XXX fix
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

// get the command to query signing info
func GetCmdQuerySigningInfo(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-info [validator-pubkey]",
		Short: "Query a validator's signing information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			pk, err := sdk.GetValPubKeyBech32(args[0])
			if err != nil {
				return err
			}
			key := slashing.GetValidatorSigningInfoKey(sdk.ValAddress(pk.Address()))
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			}
			signingInfo := new(slashing.ValidatorSigningInfo)
			cdc.MustUnmarshalBinary(res, signingInfo)

			switch viper.Get(cli.OutputFlag) {

			case "text":
				human := signingInfo.HumanReadableString()
				fmt.Println(human)

			case "json":
				// parse out the signing info
				output, err := wire.MarshalJSONIndent(cdc, signingInfo)
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
