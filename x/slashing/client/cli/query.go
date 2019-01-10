package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
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
				fmt.Println(signingInfo.HumanReadableString())
			case "json":
				if viper.GetBool(client.FlagIndentResponse) {
					out, _ := codec.MarshalJSONIndent(cdc, signingInfo)
					fmt.Println(string(out))
				} else {
					fmt.Println(string(cdc.MustMarshalJSON(signingInfo)))
				}
			}
			return nil
		},
	}

	return cmd
}

// GetCmdQueryParams implements a command to fetch slashing parameters.
func GetCmdQueryParams(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current slashing parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			route := fmt.Sprintf("custom/%s/parameters", slashing.QuerierRoute)

			res, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params slashing.Params
			cdc.MustUnmarshalJSON(res, &params)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				fmt.Println(params.HumanReadableString())
			case "json":
				if viper.GetBool(client.FlagIndentResponse) {
					out, _ := codec.MarshalJSONIndent(cdc, params)
					fmt.Println(string(out))
				} else {
					fmt.Println(string(cdc.MustMarshalJSON(params)))
				}
			}

			return nil
		},
	}

	return cmd
}
