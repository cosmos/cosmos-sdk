package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec" // XXX fix
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

// GetCmdQuerySigningInfo implements the command to query signing info.
func GetCmdQuerySigningInfo(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "signing-info [validator-pubkey]",
		Short: "Query a validator's signing information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			pk, err := sdk.GetConsPubKeyBech32(args[0])
			if err != nil {
				return err
			}

			key := slashing.GetValidatorSigningInfoKey(sdk.ConsAddress(pk.Address()))

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			var signingInfo slashing.ValidatorSigningInfo
			cdc.MustUnmarshalBinaryLengthPrefixed(res, signingInfo)
			return cliCtx.PrintOutput(signingInfo)
		},
	}
}

// GetCmdQueryParams implements a command to fetch slashing parameters.
func GetCmdQueryParams(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
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
			return cliCtx.PrintOutput(params)
		},
	}
}
