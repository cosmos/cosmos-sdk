package compat

import (
	"context"
	"encoding/json"

	grpc1 "github.com/gogo/protobuf/grpc"
	"google.golang.org/grpc"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func AppModuleHandler(id app.ModuleID, module module.AppModule) app.Handler {
	cfg := &configurator{}
	module.RegisterServices(cfg)
	return app.Handler{
		ID: id,
		InitGenesis: func(ctx context.Context, jsonCodec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
			return module.InitGenesis(types.UnwrapSDKContext(ctx), jsonCodec, message)
		},
		BeginBlocker: func(ctx context.Context, req abci.RequestBeginBlock) {
			module.BeginBlock(types.UnwrapSDKContext(ctx), req)
		},
		EndBlocker: func(ctx context.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
			return module.EndBlock(types.UnwrapSDKContext(ctx), req)
		},
		MsgServices:   cfg.msgServices,
		QueryServices: cfg.queryServices,
	}
}

type configurator struct {
	msgServices   []app.ServiceImpl
	queryServices []app.ServiceImpl
}

func (c *configurator) MsgServer() grpc1.Server {
	return &serviceRegistrar{impls: c.msgServices}
}

func (c *configurator) QueryServer() grpc1.Server {
	return &serviceRegistrar{impls: c.queryServices}
}

func (c *configurator) RegisterMigration(moduleName string, forVersion uint64, handler module.MigrationHandler) error {
	return nil
}

var _ module.Configurator = &configurator{}

type serviceRegistrar struct {
	impls []app.ServiceImpl
}

func (s *serviceRegistrar) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.impls = append(s.impls, app.ServiceImpl{
		Desc: desc,
		Impl: impl,
	})
}

var _ grpc.ServiceRegistrar = &serviceRegistrar{}
