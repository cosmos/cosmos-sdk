package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
)

const (
	FlagProve = "prove"
)

func object(cdc *codec.Codec, storeKey string, prefix []byte, connid, clientid string) connection.Object {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	man := connection.NewManager(base, climan)
	return man.CLIObject(connid, clientid)
}

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcQueryCmd := &cobra.Command{
		Use:                        "connection",
		Short:                      "Connection query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcQueryCmd.AddCommand(cli.GetCommands(
		GetCmdQueryConnection(storeKey, cdc),
	)...)
	return ibcQueryCmd
}

func QueryConnection(ctx context.CLIContext, obj connection.Object, prove bool) (res utils.JSONObject, err error) {
	q := state.NewCLIQuerier(ctx)

	conn, connp, err := obj.ConnectionCLI(q)
	if err != nil {
		return
	}
	avail, availp, err := obj.AvailableCLI(q)
	if err != nil {
		return
	}
	kind, kindp, err := obj.KindCLI(q)
	if err != nil {
		return
	}

	if prove {
		return utils.NewJSONObject(
			conn, connp,
			avail, availp,
			kind, kindp,
		), nil
	}

	return utils.NewJSONObject(
		conn, nil,
		avail, nil,
		kind, nil,
	), nil
}

func GetCmdQueryConnection(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Query stored connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			obj := object(cdc, storeKey, ibc.VersionPrefix(ibc.Version), args[0], "")
			jsonobj, err := QueryConnection(ctx, obj, viper.GetBool(FlagProve))
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, jsonobj))

			return nil
		},
	}

	cmd.Flags().Bool(FlagProve, false, "(optional) show proofs for the query results")

	return cmd
}
