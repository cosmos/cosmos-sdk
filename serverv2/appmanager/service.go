package appmanager

import (
	"context"
	"errors"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/serverv2/core/appmanager"
	"github.com/cosmos/cosmos-sdk/serverv2/core/event"
	"github.com/cosmos/cosmos-sdk/serverv2/core/transaction"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

type Store interface {
	Branch() (any, error)
	WorkingHash() ([]byte, error)
}

type AppManager[T transaction.Tx] struct {
	// ModuleManager     *module.Manager
	// configurator      module.Configurator
	// config            *runtimev1alpha1.Module
	storeKeys         []storetypes.StoreKey
	txCodec           transaction.Codec[T]
	txValidator       transaction.Validator[T]
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               codec.Codec
	// amino             *codec.LegacyAmino
	// basicManager      module.BasicManager
	// baseAppOptions    []BaseAppOption
	msgServiceRouter *MsgServiceRouter
	queryRouter      *QueryRouter
	// appConfig         *appv1alpha1.Config
	logger log.Logger
}

func NewAppManager[T transaction.Tx](txv transaction.Validator[T], txc transaction.Codec[T], logger log.Logger) appmanager.App[T] {
	return AppManager[T]{}
}

func (am AppManager[T]) ChainID() string {
	panic("implement me")
}

func (am AppManager[T]) AppVersion() (uint64, error) {
	panic("implement me")
}

func (am AppManager[T]) InitChain(context.Context, appmanager.RequestInitChain) (appmanager.ResponseInitChain, error) {
	panic("implement me")
}

func (am AppManager[T]) DeliverBlock(ctx context.Context, req appmanager.RequestDeliverBlock[T]) (appmanager.ResponseDeliverBlock, error) {
	txResult := make([]appmanager.TxResult, len(req.Txs))
	events := make([]event.Event, 0)

	beginEvents, err := BeginBlock()
	if err != nil {
		return appmanager.ResponseDeliverBlock{}, err
	}

	events = append(events, beginEvents...)

	for i, tx := range req.Txs {
		// // decode the transaction
		// tx, err := am.txCodec.Decode(bz)
		// if err != nil {
		// 	return appmanager.ResponseDeliverBlock{}, err
		// }

		// validate the transaction
		ctx, txerr := am.txValidator.Validate(ctx, []T{tx}, false)
		if txerr != nil {
			return appmanager.ResponseDeliverBlock{}, txerr[tx.Hash()] // TODO: dont return to execute other txs
		}

		// exec the transaction
		txr, err := ExecTx(ctx, am.logger, tx, false)
		if err != nil {
			return appmanager.ResponseDeliverBlock{}, err
		}
		txResult[i] = txr

	}

	endEvents, err := EndBlock()
	if err != nil {
		return appmanager.ResponseDeliverBlock{}, err
	}
	events = append(events, endEvents...)

	return appmanager.ResponseDeliverBlock{
		TxResults: txResult,
		Events:    events,
	}, nil
}

// Query implements the Query method for application based queries
func (am AppManager[T]) Query(ctx context.Context, qr *QueryRequest) (*QueryResponse, error) {
	telemetry.IncrCounter(1, "query", "count")
	telemetry.IncrCounter(1, "query", qr.Path)
	defer telemetry.MeasureSince(time.Now(), qr.Path)

	// handle gRPC routes first rather than calling splitPath because '/' characters
	// are used as part of gRPC paths
	if grpcHandler := am.queryRouter.Route(qr.Path); grpcHandler != nil {
		return grpcHandler(ctx, qr)
	}

	return nil, errors.New("unknown query path")
}

// func handleQueryGRPC(handler GRPCQueryHandler) {
// 	ctx, err := app.CreateQueryContext(req.Height, req.Prove)
// 	if err != nil {
// 		return sdkerrors.QueryResult(err, app.trace)
// 	}

// 	resp, err := handler(ctx, req)
// 	if err != nil {
// 		resp = sdkerrors.QueryResult(gRPCErrorToSDKError(err), app.trace)
// 		resp.Height = req.Height
// 		return resp
// 	}

// 	return resp
// }

/*
Things app manager needs to do:


Genesis:
- read genesis
- execute genesis txs

Queries:
- Query Router points to modules

Messages:
- Message Router points to modules

Config:
- QueryGasLimit
- HaltTime
- HaltBlock

Recovery:
- Panic Recovery for the app manager


*/
