package internal

import (
	"context"
	"encoding/json"
	io "io"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/app"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/container"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AppName string

type Inputs struct {
	dig.In

	Name AppName
	sdk.TxDecoder
	sdk.AnteHandler
	Options      []func(*baseapp.BaseApp) `group:"init"`
	Handlers     []app.Handler            `group:"app"`
	InitGenesis  func(context.Context, codec.JSONCodec, json.RawMessage) abci.ResponseInitChain
	BeginBlocker func(context.Context, abci.RequestBeginBlock) abci.ResponseBeginBlock
	EndBlocker   func(context.Context, abci.RequestEndBlock) abci.ResponseEndBlock
}

func provide(inputs Inputs) servertypes.AppCreator {
	return func(logger log.Logger, db dbm.DB, traceStore io.Writer, options servertypes.AppOptions) servertypes.Application {
		bapp := baseapp.NewBaseApp(
			string(inputs.Name),
			logger,
			db,
			inputs.TxDecoder,
			inputs.Options...,
		)

		bapp.SetAnteHandler(inputs.AnteHandler)
		bapp.SetInitChainer(func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
			panic("TODO")
		})
		bapp.SetBeginBlocker(func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
			return inputs.BeginBlocker(sdk.WrapSDKContext(ctx), req)
		})
		bapp.SetEndBlocker(func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
			return inputs.EndBlocker(sdk.WrapSDKContext(ctx), req)
		})

		bapp.SetCommitMultiStoreTracer(traceStore)

		for _, handler := range inputs.Handlers {
			for _, svc := range handler.MsgServices {
				bapp.MsgServiceRouter().RegisterService(svc.Desc, svc.Impl)
			}

			for _, svc := range handler.QueryServices {
				bapp.GRPCQueryRouter().RegisterService(svc.Desc, svc.Impl)
			}
		}

		return &theApp{
			BaseApp: bapp,
		}
	}
}

var Module = container.Provide(provide)
