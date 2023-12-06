package appmanager

import (
	"context"

	"cosmossdk.io/core/appmanager"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	queryRouter      *GRPCQueryRouter
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

		// pretransaction hook
		// TODO: move to execTx

		// exec the transaction
		txr, err := ExecTx(ctx, am.logger, tx, false)
		if err != nil {
			return appmanager.ResponseDeliverBlock{}, err
		}
		txResult[i] = txr

		// posttransaction hook
		// TODO: move to ExecTx
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

/*
Things app manager needs to do:

Transaction:
- txdecoder
	- the transaction is already decoded in consensus there should not be a need to decode it again here
	- we should register the interface registstry to the txCodec
- txvalidator
	- needs to register the antehandlers

Execution: (DeliverBlock)
- execution of a transaction
	- Preblock call
	- BeginBlock call
		- PremessageHook
	- DeliverTx call
		- PostmessageHook
	- EndBlock call
- ability to register hooks
- ability to register messages and queries

Genesis:
- read genesis
- execute genesis txs

States:
- ExecuteTx
- SimulateTx

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
