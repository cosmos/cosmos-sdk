package ibc

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientcli "github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectioncli "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelcli "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

type AppModuleBasic struct{}

var _ module.AppModuleBasic = AppModuleBasic{}

func (AppModuleBasic) Name() string {
	return "ibc"
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	commitment.RegisterCodec(cdc)
	merkle.RegisterCodec(cdc)
	client.RegisterCodec(cdc)
	tendermint.RegisterCodec(cdc)
	channel.RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return nil
}

func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	return nil
}

func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, router *mux.Router) {
	// noop
}

func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "ibc",
		Short:                      "IBC tx subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	storeKey := "ibc"

	cmd.AddCommand(cli.GetCommands(
		clientcli.GetCmdCreateClient(cdc),
		clientcli.GetCmdUpdateClient(cdc),
		connectioncli.GetCmdConnectionHandshake(storeKey, cdc),
		channelcli.GetCmdChannelHandshake(storeKey, cdc),
		channelcli.GetCmdRelay(storeKey, cdc),
	)...)

	return cmd
}

func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc",
		Short: "IBC query subcommands",
		//	DisableFlagParsing:         true,
		//		SuggestionsMinumumDistance: 2,
	}

	storeKey := "ibc"

	cmd.AddCommand(cli.GetCommands(
		clientcli.GetCmdQueryConsensusState(storeKey, cdc),
		clientcli.GetCmdQueryHeader(cdc),
		clientcli.GetCmdQueryClient(storeKey, cdc),
		connectioncli.GetCmdQueryConnection(storeKey, cdc),
		channelcli.GetCmdQueryChannel(stoerKey, cdc),
	)...)

	return cmd
}

type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

func (AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// TODO
}

func (AppModule) Route() string {
	return "ibc"
}

func (AppModule) QuerierRoute() string {
	return "ibc"
}

func (am AppModule) NewHandler() sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case client.MsgCreateClient:
			return client.HandleMsgCreateClient(ctx, msg, am.keeper.client)
		case client.MsgUpdateClient:
			return client.HandleMsgUpdateClient(ctx, msg, am.keeper.client)
		case connection.MsgOpenInit:
			return connection.HandleMsgOpenInit(ctx, msg, am.keeper.connection)
		case connection.MsgOpenTry:
			return connection.HandleMsgOpenTry(ctx, msg, am.keeper.connection)
		case connection.MsgOpenAck:
			return connection.HandleMsgOpenAck(ctx, msg, am.keeper.connection)
		case connection.MsgOpenConfirm:
			return connection.HandleMsgOpenConfirm(ctx, msg, am.keeper.connection)
		case channel.MsgOpenInit:
			return channel.HandleMsgOpenInit(ctx, msg, am.keeper.channel)
		case channel.MsgOpenTry:
			return channel.HandleMsgOpenTry(ctx, msg, am.keeper.channel)
		case channel.MsgOpenAck:
			return channel.HandleMsgOpenAck(ctx, msg, am.keeper.channel)
		case channel.MsgOpenConfirm:
			return channel.HandleMsgOpenConfirm(ctx, msg, am.keeper.channel)
		case channel.MsgReceive:
			return channel.HandleMsgReceive(ctx, msg, am.keeper.channel.Manager())
		default:
			return sdk.ErrUnknownRequest("unrecognized msg type").Result()
		}
	}
}
