package app

import (
	"fmt"
	io "io"
	"reflect"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Name string

var BaseAppProvider = container.Options(
	container.AutoGroupTypes(reflect.TypeOf(func(*baseapp.BaseApp) {})),
	container.Provide(provideBaseApp),
)

type baseAppInput struct {
	container.StructArgs

	Name      Name          `optional:"true"`
	TxDecoder sdk.TxDecoder `optional:"true"`
	Options   []func(*baseapp.BaseApp)
}

type app struct {
	*baseapp.BaseApp
}

var _ types.Application

func provideBaseApp(inputs baseAppInput) types.AppCreator {
	return func(logger log.Logger, db dbm.DB, tracer io.Writer, appOptions types.AppOptions) types.Application {
		name := inputs.Name
		if name == "" {
			name = "simapp"
		}

		txDecoder := inputs.TxDecoder
		if txDecoder == nil {
			txDecoder = func(txBytes []byte) (sdk.Tx, error) {
				return nil, fmt.Errorf("no TxDecoder, can't decode transactions")
			}
		}

		baseApp := baseapp.NewBaseApp(string(name), logger, db, txDecoder, inputs.Options...)

		if tracer != nil {
			baseApp.SetCommitMultiStoreTracer(tracer)
		}

		return &app{baseApp}
	}
}

func (a app) RegisterAPIRoutes(server *api.Server, config config.APIConfig) {
	panic("implement me")
}

func (a app) RegisterTxService(clientCtx client.Context) {
	panic("implement me")
}

func (a app) RegisterTendermintService(clientCtx client.Context) {
	panic("implement me")
}
