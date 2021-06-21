package internal

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/query"
	genutilprovider "github.com/cosmos/cosmos-sdk/x/genutil/provider"
)

func AppConfigProvider(config *app.Config) container.Option {
	moduleConfigMap := map[string]*codecTypes.Any{}

	if config.Abci.TxHandler == nil {
		return container.Error(fmt.Errorf("missing tx handler"))
	}
	moduleConfigMap["tx"] = config.Abci.TxHandler

	for _, modConfig := range config.Modules {
		moduleConfigMap[modConfig.Name] = modConfig.Config
	}

	provideAbciMethods := container.Provide(func(in inputs) (outputs, error) {
		handlerMap := map[string]app.Handler{}
		for _, h := range in.Handlers {
			handlerMap[h.ID.Name()] = h
		}

		// InitGenesis
		var initGenesis []func(context.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate
		got := map[string]bool{}
		for _, name := range config.Abci.InitGenesis {
			h, ok := handlerMap[name]
			if !ok || h.InitGenesis == nil {
				return outputs{}, fmt.Errorf("missing InitGenesis for module %s", name)
			}
			got[name] = true
			initGenesis = append(initGenesis, h.InitGenesis)
		}

		for name, h := range handlerMap {
			if !got[name] && h.InitGenesis != nil {
				return outputs{}, fmt.Errorf("module %s is missing from init_genesis order", name)
			}
		}

		// BeginBlock
		var beginBlock []func(context.Context, abci.RequestBeginBlock)
		got = map[string]bool{}
		for _, name := range config.Abci.BeginBlock {
			h, ok := handlerMap[name]
			if !ok || h.BeginBlocker == nil {
				return outputs{}, fmt.Errorf("missing BeginBlocker for module %s", name)
			}
			got[name] = true
			beginBlock = append(beginBlock, h.BeginBlocker)
		}

		for name, h := range handlerMap {
			if !got[name] && h.BeginBlocker != nil {
				return outputs{}, fmt.Errorf("module %s is missing from begin_block order", name)
			}
		}

		// EndBlock
		var endBlock []func(context.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
		got = map[string]bool{}
		for _, name := range config.Abci.EndBlock {
			h, ok := handlerMap[name]
			if !ok || h.EndBlocker == nil {
				return outputs{}, fmt.Errorf("missing EndBlocker for module %s", name)
			}
			got[name] = true
			endBlock = append(endBlock, h.EndBlocker)
		}

		for name, h := range handlerMap {
			if !got[name] && h.EndBlocker != nil {
				return outputs{}, fmt.Errorf("module %s is missing from end_block order", name)
			}
		}

		return outputs{
			InitGenesis: func(goCtx context.Context, jsonCodec codec.JSONCodec, message json.RawMessage) abci.ResponseInitChain {
				panic("TODO")
			},
			BeginBlocker: func(goCtx context.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
				ctx := sdk.UnwrapSDKContext(goCtx)
				ctx = ctx.WithEventManager(sdk.NewEventManager())

				for _, f := range beginBlock {
					f(sdk.WrapSDKContext(ctx), req)
				}

				return abci.ResponseBeginBlock{
					Events: ctx.EventManager().ABCIEvents(),
				}
			},
			EndBlocker: func(ctx context.Context, block abci.RequestEndBlock) abci.ResponseEndBlock {
				panic("TODO")
			},
		}, nil
	})

	return container.Options(
		provideAbciMethods,
		app.ComposeModules(moduleConfigMap),
		// TODO should these be here:
		container.Provide(genutilprovider.Provider),
		query.Module,
	)
}

type inputs struct {
	Handlers []app.Handler `group:"app"`
}

type outputs struct {
	InitGenesis  func(context.Context, codec.JSONCodec, json.RawMessage) abci.ResponseInitChain
	BeginBlocker func(context.Context, abci.RequestBeginBlock) abci.ResponseBeginBlock
	EndBlocker   func(context.Context, abci.RequestEndBlock) abci.ResponseEndBlock
}
