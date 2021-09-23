package module

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/server"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"

	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/compat"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _, _ app.Module = &Module{}, &TxHandler{}

func (m *Module) RegisterTypes(registry types.TypeRegistry) {
	authtypes.RegisterInterfaces(registry)
}

type inputs struct {
	container.In
}

type outputs struct {
	container.Out

	Handler app.Handler
	Keeper  authkeeper.AccountKeeperI
}

func (m Module) provide(codec codec.Codec, storeKey *sdk.KVStoreKey) (
	app.Handler,
	authkeeper.AccountKeeperI,
	error,
) {
	if m.Bech32AccountPrefix == "" {
		return app.Handler{}, nil, fmt.Errorf("missing bech32_account_prefix")
	}

	keeper := authkeeper.NewAccountKeeper(codec, storeKey, panic("TODO"), nil, nil, m.Bech32AccountPrefix)
	am := auth.NewAppModule(codec, keeper, nil)
	return compat.AppModuleHandler(am), keeper, nil
}

func (m *TxHandler) RegisterTypes(interfaceRegistry types.TypeRegistry) {
	sdk.RegisterInterfaces(interfaceRegistry)
	txtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
}

type txHandlerInputs struct {
	container.In

	Codec          codec.ProtoCodecMarshaler
	Handlers       map[string]app.Handler
	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	FeeGrantKeeper feegrantkeeper.Keeper `optional:"true"`
}

func (m *TxHandler) provide(in txHandlerInputs) func(servertypes.AppOptions) func(*baseapp.BaseApp) {
	return func(appOpts servertypes.AppOptions) func(*baseapp.BaseApp) {
		return func(baseApp *baseapp.BaseApp) {
			txConfig := tx.NewTxConfig(in.Codec, tx.DefaultSignModes)
			anteHandler, err := ante.NewAnteHandler(
				ante.HandlerOptions{
					AccountKeeper:   in.AccountKeeper,
					BankKeeper:      in.BankKeeper,
					SignModeHandler: txConfig.SignModeHandler(),
					FeegrantKeeper:  in.FeeGrantKeeper,
					SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
				},
			)
			if err != nil {
				panic(err)
			}

			indexEventsStr := cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))

			indexEvents := map[string]struct{}{}
			for _, e := range indexEventsStr {
				indexEvents[e] = struct{}{}
			}

			msgSvcRouter := authmiddleware.NewMsgServiceRouter(in.Codec.InterfaceRegistry())

			for _, handler := range in.Handlers {
				for _, svc := range handler.MsgServices {
					msgSvcRouter.RegisterService(svc.Desc, svc.Impl)
				}
			}

			txHandler, err := authmiddleware.NewDefaultTxHandler(authmiddleware.TxHandlerOptions{
				Debug:             cast.ToBool(appOpts.Get(server.FlagTrace)),
				IndexEvents:       indexEvents,
				LegacyRouter:      authmiddleware.NewLegacyRouter(),
				MsgServiceRouter:  msgSvcRouter,
				LegacyAnteHandler: anteHandler,
			})
			if err != nil {
				panic(err)
			}

			baseApp.SetTxHandler(txHandler)
		}
	}
}
