package runtime

import (
	"context"
	"errors"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
	"golang.org/x/exp/maps"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf"
)

type MMv2 struct {
	config  *runtimev2.Module
	modules map[string]appmodule.AppModule
}

func NewMMv2(config *runtimev2.Module, modules map[string]appmodule.AppModule) *MMv2 {
	modulesName := maps.Keys(modules)

	if len(config.BeginBlockers) == 0 {
		config.BeginBlockers = modulesName
	}
	if len(config.EndBlockers) == 0 {
		config.EndBlockers = modulesName
	}
	if len(config.TxValidation) == 0 {
		config.TxValidation = modulesName
	}
	if len(config.InitGenesis) == 0 {
		config.InitGenesis = modulesName
	}
	if len(config.ExportGenesis) == 0 {
		config.ExportGenesis = modulesName
	}

	return &MMv2{
		config:  config,
		modules: modules,
	}
}

// TODO refactor
func (m *MMv2) BeginBlock() func(ctx context.Context) error {
	// TODO rewrap the context into sdk.Context

	return func(ctx context.Context) error {
		for _, moduleName := range m.config.BeginBlockers {
			if module, ok := m.modules[moduleName].(appmodule.HasBeginBlocker); ok {
				if err := module.BeginBlock(ctx); err != nil {
					return fmt.Errorf("beginblocker of module %s failure: %w", module, err)
				}
			}
		}

		return nil
	}
}

// TODO refactor
func (m *MMv2) EndBlock() (endblock func(ctx context.Context) error, valupdate func(ctx context.Context) ([]appmanager.ValidatorUpdate, error)) {
	// TODO rewrap the context into sdk.Context

	validatorUpdates := []abci.ValidatorUpdate{}

	endBlock := func(ctx context.Context) error {
		for _, moduleName := range m.config.EndBlockers {
			if module, ok := m.modules[moduleName].(appmodule.HasEndBlocker); ok {
				err := module.EndBlock(ctx)
				if err != nil {
					return err
				}
			} else if module, ok := m.modules[moduleName].(sdkmodule.HasABCIEndBlock); ok { // we need to keep this for our module compatibility promise
				moduleValUpdates, err := module.EndBlock(ctx)
				if err != nil {
					return err
				}
				// use these validator updates if provided, the module manager assumes
				// only one module will update the validator set
				if len(moduleValUpdates) > 0 {
					if len(validatorUpdates) > 0 {
						return errors.New("validator end block updates already set by a previous module")
					}

					for _, updates := range moduleValUpdates {
						validatorUpdates = append(validatorUpdates, abci.ValidatorUpdate{PubKey: updates.PubKey, Power: updates.Power})
					}
				}
			} else {
				continue
			}
		}

		return nil
	}

	valUpdate := func(ctx context.Context) ([]appmanager.ValidatorUpdate, error) {
		valUpdates := make([]appmanager.ValidatorUpdate, len(validatorUpdates))
		for i, v := range validatorUpdates {
			valUpdates[i] = appmanager.ValidatorUpdate{
				PubKey: v.PubKey.GetSecp256K1(),
				Power:  v.Power,
			}
		}

		return valUpdates, nil
	}

	return endBlock, valUpdate
}

// UpgradeBlocker is PreBlocker for server v2, it supports only the upgrade module
func (m *MMv2) UpgradeBlocker() func(ctx context.Context) (bool, error) {
	// TODO rewrap the context into sdk.Context

	return func(ctx context.Context) (bool, error) {
		for _, moduleName := range m.config.BeginBlockers {
			if moduleName != "upgrade" {
				continue
			}

			if module, ok := m.modules[moduleName].(interface {
				UpgradeBlocker() func(ctx context.Context) (bool, error)
			}); ok {
				return module.UpgradeBlocker()(ctx)
			}
		}

		return false, fmt.Errorf("no upgrade module found")
	}
}

// TxValidators validates incoming transactions
func (m *MMv2) TxValidation() func(ctx context.Context, tx transaction.Tx) error {
	// TODO rewrap the context into sdk.Context

	return func(ctx context.Context, tx transaction.Tx) error {
		for _, moduleName := range m.config.TxValidation {
			if module, ok := m.modules[moduleName].(appmodule.HasTxValidation[transaction.Tx]); ok {
				return module.TxValidator(ctx, tx)
			}
		}

		return nil
	}
}

// TODO refactor
func (m *MMv2) RegisterMsgs(builder *stf.MsgRouterBuilder) error { // most important part of the PR to finish
	for _, module := range m.modules {
		_ = module
		// 	builder.RegisterHandler()
		// 	builder.RegisterPostHandler()
		// 	builder.RegisterPreHandler()
	}

	return nil
}
