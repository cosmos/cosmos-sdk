package cli

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc"
)

const (
	FlagNode1 = "node1"
	FlagNode2 = "node2"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcTxCmd := &cobra.Command{
		Use:                        "ibc",
		Short:                      "IBC transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcTxCmd.AddCommand(client.PostCommands(
		GetCmdEstablish(cdc),
		GetCmdRelay(cdc),
	)...)

	return ibcTxCmd
}

// gaiad tx ibc establish --node1 tcp://() --node2 tcp://() clientid connectionid channelid
func GetCmdEstablish(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "establish",
		Short: "create new client with a consensus state",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx1 := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc).
				WithNode(viper.GetString(FlagNode1))

			cliCtx2 := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc).
				WithNoe(viper.GetString(FlagNode2))

			keeper := ibc.DummyKeeper()

		
		},
	}

	cmd.Flags().AddFlagSet(FlagIP)

	return cmd
}

func BuildCreateClientMsg(cliCtx context.CLIContext, statePath string, txBldr auth.TxBuilder) (auth.TxBuilder, sdk.Msg, error) {
	contents, err := ioutil.ReadFile(statePath)
	if err != nil {
		return txBldr, nil, err
	}

	var state client.ConsensusState
	if cliCtxFrom.Cdc.UnmarshalJSON(contents, &state); err != nil {
		return txBldr, nil, err
	}
	
	msg := ibc.MsgCreateClient {
		ClientID: viper.GetString(FlagClientID),
		ConsensusState: state,
		Signer: cliCtx.GetFromAddress().GetFromAddress()
	}

	return txBldr, msg, nil
}

// gaiad tx ibc relay
func GetCmdRelay(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay",
		Short: "relay packets between two chains",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx1 := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc).
				WithNode(viper.GetString(FlagNode1))

			cliCtx2 := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc).
				WithNoe(viper.GetString(FlagNode2))

			keeper := ibc.DummyKeeper()
		
			
		
		},
	}

	cmd.Flags().AddFlagSet(FlagIP)

	return cmd
}

// Copied from client/context/query.go
func query(fromCtx context.CLIContext, key []byte) ([]byte, merkle.Proof, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, nil, err
	}

	opts := rpcclient.ABCIQueryOptions {
		Height: ctx.Height,
		Prove: true,
	}

	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return nil, nil, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return nil, nil, errors.New(resp.Log)
	}

	return resp.Value, resp.Proof, nil
}

func relay(fromCtx, toCtx context.CLIContext, connid, chanid string) {
	keeper := ibc.DummyKeeper()

	for {
		time.Sleep(5 * time.Second)
	
		obj := keeper.channel.Object(connid, chanid)

		value, proof, err := query(fromCtx, obj.seqsend.Key())
		if err != nil {
			// XXX
		}

		msg := ibc.MsgReceive{
			ConnectionID: connid,
			ChannelID: chanid,
			Packet: toCtx.Cdc.MustUnmarshalBinary(value),
			Signer: toCtx.GetFromAddress(),
		}

		utils.GenerateOrBroadcastMsgs(toCtx, txBldr, []sdk.Msg{msg})
	}
}
