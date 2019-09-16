package ibc

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clicli "github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	conncli "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/cli"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

const (
	ModuleName = "ibc"
	StoreKey   = ModuleName
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

var _ module.AppModuleBasic = AppModuleBasic{}

func (AppModuleBasic) Name() string {
	return ModuleName
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	commitment.RegisterCodec(cdc)
	merkle.RegisterCodec(cdc)
	client.RegisterCodec(cdc)
	tendermint.RegisterCodec(cdc)
	channel.RegisterCodec(cdc)

	client.SetMsgCodec(cdc)
	channel.SetMsgCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return nil
}

func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	return nil
}

func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	//noop
}

func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc",
		Short: "IBC transaction subcommands",
	}

	cmd.AddCommand(
		clicli.GetTxCmd(ModuleName, cdc),
		conncli.GetTxCmd(ModuleName, cdc),
		//		chancli.GetTxCmd(ModuleName, cdc),
	)

	return cmd
}

func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc",
		Short: "IBC query subcommands",
	}

	cmd.AddCommand(
		clicli.GetQueryCmd(ModuleName, cdc),
		conncli.GetQueryCmd(ModuleName, cdc),
		//		chancli.GetQueryCmd(ModuleName, cdc),
	)

	return cmd
}

type AppModule struct {
	AppModuleBasic
	Keeper
}

func NewAppModule(k Keeper) AppModule {
	return AppModule{
		Keeper: k,
	}
}

func (AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {

}

func (AppModule) Name() string {
	return ModuleName
}

func (AppModule) Route() string {
	return ModuleName
}

func (am AppModule) NewHandler() sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case client.MsgCreateClient:
			return client.HandleMsgCreateClient(ctx, msg, am.client)
		case client.MsgUpdateClient:
			return client.HandleMsgUpdateClient(ctx, msg, am.client)
		case connection.MsgOpenInit:
			return connection.HandleMsgOpenInit(ctx, msg, am.connection)
		case connection.MsgOpenTry:
			return connection.HandleMsgOpenTry(ctx, msg, am.connection)
		case connection.MsgOpenAck:
			return connection.HandleMsgOpenAck(ctx, msg, am.connection)
		case connection.MsgOpenConfirm:
			return connection.HandleMsgOpenConfirm(ctx, msg, am.connection)
		case channel.MsgOpenInit:
			return channel.HandleMsgOpenInit(ctx, msg, am.channel)
		case channel.MsgOpenTry:
			return channel.HandleMsgOpenTry(ctx, msg, am.channel)
		case channel.MsgOpenAck:
			return channel.HandleMsgOpenAck(ctx, msg, am.channel)
		case channel.MsgOpenConfirm:
			return channel.HandleMsgOpenConfirm(ctx, msg, am.channel)
		default:
			errMsg := fmt.Sprintf("unrecognized IBC message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func (am AppModule) QuerierRoute() string {
	return ModuleName
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return nil
}

func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return nil
}

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {

}

func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
