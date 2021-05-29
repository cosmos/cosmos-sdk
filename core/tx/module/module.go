package module

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/core/app_config"

	tx2 "github.com/cosmos/cosmos-sdk/types/tx"

	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/core/module/app"
	"github.com/cosmos/cosmos-sdk/core/tx"
)

//func init() {
//	app.RegisterAppModule(appModule{})
//}

type appModule struct {
	Config *tx.Module
	Codec  codec.Codec
}

func (a appModule) New() app.Handler {
	return appModule{}
}

var _ app_config.HasTxHandler = appModule{}

func (a appModule) RegisterTypes(registry types.InterfaceRegistry) {
	panic("implement me")
}

func (a appModule) InitGenesis(ctx context.Context, codec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	panic("implement me")
}

func (a appModule) ExportGenesis(ctx context.Context, codec codec.JSONCodec) json.RawMessage {
	panic("implement me")
}

func (a appModule) RegisterMsgServices(registrar grpc.ServiceRegistrar) {}

func (a appModule) RegisterQueryServices(registrar grpc.ServiceRegistrar) {}

func (a appModule) TxHandler(params app_config.TxHandlerParams) app_config.TxHandler {
	return txHandler{
		Module:    a.Config,
		msgRouter: params.MsgRouter,
	}
}

type HasMiddleware interface {
	RegisterTxMiddleware(registrar MiddlewareRegistrar)
}

type MiddlewareRegistrar interface {
	RegisterTxMiddlewareFactory(configType interface{}, factory MiddlewareFactory)
}

type MiddlewareFactory func(config interface{}) Middleware

//type Middleware interface {
//	OnCheckTx(ctx context.Context, tx tx2.Tx, req abci.RequestCheckTx, next TxHandler) (abci.ResponseCheckTx, error)
//	OnDeliverTx(ctx context.Context, tx tx2.Tx, req abci.RequestDeliverTx, next TxHandler) (abci.ResponseDeliverTx, error)
//}

type TxHandler interface {
	CheckTx(ctx context.Context, tx tx2.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error)
	DeliverTx(ctx context.Context, tx tx2.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error)
}

type txHandler struct {
	*tx.Module
	msgRouter app_config.MsgRouter
}

func (t txHandler) CheckTx(ctx context.Context, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	panic("implement me")
}

func (t txHandler) DeliverTx(ctx context.Context, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	panic("implement me")
}

type Middleware func(TxHandler) TxHandler
