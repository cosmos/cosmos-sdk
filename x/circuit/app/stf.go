package main

import (
	"context"

	"cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/x/circuit/keeper"
)

// this installs circuit into a stf

type addrCodec struct {
}

func (a addrCodec) StringToBytes(text string) ([]byte, error) {
	return []byte(text), nil
}

func (a addrCodec) BytesToString(bz []byte) (string, error) {
	return string(bz), nil
}

func makeKeeper() keeper.Keeper {
	return keeper.NewKeeper(
		stf.NewKVStoreService([]byte("circuit")),
		"gov",
		addrCodec{},
	)
}

type Tx struct {
}

func (t Tx) Hash() [32]byte {
	// TODO implement me
	panic("implement me")
}

func (t Tx) GetMessages() ([]transaction.Type, error) {
	// TODO implement me
	panic("implement me")
}

func (t Tx) GetSenders() ([]transaction.Identity, error) {
	// TODO implement me
	panic("implement me")
}

func (t Tx) GetGasLimit() (uint64, error) {
	// TODO implement me
	panic("implement me")
}

func (t Tx) Bytes() []byte {
	// TODO implement me
	panic("implement me")
}

func NewStf() *stf.STF[Tx] {
	msgRouter := stf.NewMsgRouterBuilder()
	queryRouter := stf.NewMsgRouterBuilder()
	InstallCircuitModule(msgRouter, queryRouter)

	handleMsg, err := msgRouter.Build()
	if err != nil {
		panic(err)
	}
	handleQuery, err := queryRouter.Build()
	if err != nil {
		panic(err)
	}
	return stf.NewSTF(
		handleMsg,
		handleQuery,
		func(ctx context.Context, txs []Tx) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context, tx Tx) error {
			return nil
		},
		func(ctx context.Context) ([]appmodule.ValidatorUpdate, error) {
			return nil, nil
		},
		func(ctx context.Context, tx Tx, success bool) error {
			return nil
		},
		branch.DefaultNewWriterMap,
	)
}

func InstallCircuitModule(
	msgRouter *stf.MsgRouterBuilder,
	queryRouter *stf.MsgRouterBuilder,
) {
	k := makeKeeper()
	msgServer := keeper.NewMsgServerImpl(k)
	appmodule.RegisterHandler(msgRouter, msgServer.ResetCircuitBreaker)
	appmodule.RegisterHandler(msgRouter, msgServer.TripCircuitBreaker)

	queryServer := keeper.NewQueryServer(k)
	appmodule.RegisterHandler(queryRouter, queryServer.Account)
	appmodule.RegisterHandler(queryRouter, queryServer.Accounts)
	appmodule.RegisterHandler(queryRouter, queryServer.DisabledList)
}
