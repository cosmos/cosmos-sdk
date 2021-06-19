package internal

import (
	io "io"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/cosmos/cosmos-sdk/container"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"go.uber.org/dig"
)

type AppName string

type Inputs struct {
	dig.In

	Name AppName
	sdk.TxDecoder
	sdk.AnteHandler
	Options []func(*baseapp.BaseApp) `group:"init"`
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

		bapp.SetCommitMultiStoreTracer(writer)

		return &theApp{
			BaseApp: bapp,
		}
	}
}

var Module = container.Provide(provide)
