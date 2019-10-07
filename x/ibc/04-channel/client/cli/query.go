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

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

const (
	FlagProve = "prove"
)

func object(ctx context.CLIContext, cdc *codec.Codec, storeKey string, prefix []byte, portid, chanid string) (channel.State, error) {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewManager(base, connman)
	return man.CLIQuery(state.NewCLIQuerier(ctx), portid, chanid)
}

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcQueryCmd := &cobra.Command{
		Use:                "channel",
		Short:              "Channel query subcommands",
		DisableFlagParsing: true,
	}

	ibcQueryCmd.AddCommand(cli.GetCommands(
		GetCmdQueryChannel(storeKey, cdc),
	)...)

	return ibcQueryCmd
}

func QueryChannel(ctx context.CLIContext, obj channel.State, prove bool) (res utils.JSONState, err error) {
	q := state.NewCLIQuerier(ctx)

	conn, connp, err := obj.ChannelCLI(q)
	if err != nil {
		return
	}
	avail, availp, err := obj.AvailableCLI(q)
	if err != nil {
		return
	}

	seqsend, seqsendp, err := obj.SeqSendCLI(q)
	if err != nil {
		return
	}

	seqrecv, seqrecvp, err := obj.SeqRecvCLI(q)
	if err != nil {
		return
	}

	if prove {
		return utils.NewJSONState(
			conn, connp,
			avail, availp,
			//			kind, kindp,
			seqsend, seqsendp,
			seqrecv, seqrecvp,
		), nil
	}

	return utils.NewJSONState(
		conn, nil,
		avail, nil,
		seqsend, nil,
		seqrecv, nil,
	), nil
}

func GetCmdQueryChannel(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel [port-id] [chan-id]",
		Short: "Query stored connection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			obj, err := object(ctx, cdc, storeKey, version.DefaultPrefix(), args[0], args[1])
			if err != nil {
				return err
			}
			jsonobj, err := QueryChannel(ctx, obj, viper.GetBool(FlagProve))
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

/*
func object(cdc *codec.Codec, storeKey string, prefix []byte, portid, chanid string, connids []string) channel.Stage {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewManager(base, connman)
	return man.CLIState(portid, chanid, connids)
}
*/
/*
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcQueryCmd := &cobra.Command{
		Use:                        "connection",
		Short:                      "Channel query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcQueryCmd.AddCommand(cli.GetCommands(
	//		GetCmdQueryChannel(storeKey, cdc),
	)...)
	return ibcQueryCmd
}
*/
/*


 */
