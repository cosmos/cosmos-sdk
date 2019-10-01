package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	storestate "github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

const (
	FlagProve = "prove"
)

func state(cdc *codec.Codec, storeKey string, prefix []byte, connid, clientid string) connection.State {
	base := storestate.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	man := connection.NewManager(base, climan)
	return man.CLIState(connid, clientid)
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

func QueryConnection(ctx context.CLIContext, obj connection.State, prove bool) (res utils.JSONState, err error) {
	q := storestate.NewCLIQuerier(ctx)

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
		return utils.NewJSONState(
			conn, connp,
			avail, availp,
			kind, kindp,
		), nil
	}

	return utils.NewJSONState(
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
			state := state(cdc, storeKey, version.Prefix(version.Version), args[0], "")
			jsonObj, err := QueryConnection(ctx, state, viper.GetBool(FlagProve))
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, jsonObj))

			return nil
		},
	}

	cmd.Flags().Bool(FlagProve, false, "(optional) show proofs for the query results")

	return cmd
}
