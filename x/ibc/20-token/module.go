package token

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-token/client/cli"
)

type AppModuleBasic struct{}

var _ module.AppModuleBasic = AppModuleBasic{}

func (AppModuleBasic) Name() string {
	return "token"
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return nil
}

func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	return nil
}

func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	// TODO
}

func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return nil
}

type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

var _ channel.IBCModule = AppModule{}
var _ module.AppModule = AppModule{}

func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

func (AppModule) Name() string {
	return "token"
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// noop
}

func (AppModule) Route() string {
	return "token"
}

func (am AppModule) NewHandler() sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			err := am.keeper.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.ToConnection, msg.ToChannel, msg.Amount)
			if err != nil {
				return err.Result()
			}
			return sdk.Result{}
		default:
			return sdk.ErrUnknownRequest("qqq").Result()
		}
	}
}

func (am AppModule) NewIBCHandler() channel.Handler {
	return func(ctx sdk.Context, packet channel.Packet) sdk.Result {
		switch packet := packet.(type) {
		case PacketSend:
			err := am.keeper.receiveCoins(ctx, packet.ToAddress, packet.Amount)
			if err != nil {
				return err.Result()
			}
			return sdk.Result{}
		default:
			return sdk.ErrUnknownRequest("ttt").Result()
		}
	}
}

func (am AppModule) QuerierRoute() string {
	return "token"
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	panic("TODO")
}

func (am AppModule) BeginBlock(sdk.Context, abci.RequestBeginBlock) {
}

func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return nil
}
