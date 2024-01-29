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
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/stf"
)

type MMv2 struct {
	config  *runtimev2.Module
	modules map[string]appmodule.AppModule
}

func NewMMv2(config *runtimev2.Module, modules map[string]appmodule.AppModule) *MMv2 {
	modulesName := maps.Keys(modules)

	if len(config.PreBlockers) == 0 {
		config.PreBlockers = modulesName
	}
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

// BeginBlock runs the begin-block logic of all modules
func (m *MMv2) BeginBlock() func(ctx context.Context) error {
	// TODO rewrap the context into sdk.Context

	return func(ctx context.Context) error {
		for _, moduleName := range m.config.BeginBlockers {
			if module, ok := m.modules[moduleName].(appmodule.HasBeginBlocker); ok {
				if err := module.BeginBlock(ctx); err != nil {
					return fmt.Errorf("failed to run beginblocker for %s: %w", moduleName, err)
				}
			}
		}

		return nil
	}
}

// EndBlock runs the end-block logic of all modules and tx validator updates
func (m *MMv2) EndBlock() (endblock func(ctx context.Context) error, valupdate func(ctx context.Context) ([]appmodule.ValidatorUpdate, error)) {
	// TODO rewrap the context into sdk.Context

	validatorUpdates := []abci.ValidatorUpdate{}

	endBlock := func(ctx context.Context) error {
		for _, moduleName := range m.config.EndBlockers {
			if module, ok := m.modules[moduleName].(appmodule.HasEndBlocker); ok {
				err := module.EndBlock(ctx)
				if err != nil {
					return fmt.Errorf("failed to run endblock for %s: %w", moduleName, err)
				}
			} else if module, ok := m.modules[moduleName].(sdkmodule.HasABCIEndBlock); ok { // we need to keep this for our module compatibility promise
				moduleValUpdates, err := module.EndBlock(ctx)
				if err != nil {
					return fmt.Errorf("failed to run enblock for %s: %w", moduleName, err)
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

	valUpdate := func(ctx context.Context) ([]appmodule.ValidatorUpdate, error) {
		valUpdates := []appmodule.ValidatorUpdate{}

		// get validator updates of legacy modules using HasABCIEndBlock
		for i, v := range validatorUpdates {
			valUpdates[i] = appmodule.ValidatorUpdate{
				PubKey: v.PubKey.GetEd25519(),
				Power:  v.Power,
			}
		}

		// get validator updates of modules implementing directly the new HasUpdateValidators interface
		for _, v := range m.modules {
			if module, ok := v.(appmodule.HasUpdateValidators); ok {
				up, err := module.UpdateValidators(ctx)
				if err != nil {
					return nil, err
				}

				valUpdates = append(valUpdates, up...)
			}
		}

		return valUpdates, nil
	}

	return endBlock, valUpdate
}

// PreBlocker runs the pre-block logic of all modules
func (m *MMv2) PreBlocker() func(ctx context.Context, txs []transaction.Tx) error {
	// TODO rewrap the context into sdk.Context

	return func(ctx context.Context, txs []transaction.Tx) error {
		for _, moduleName := range m.config.PreBlockers {
			if module, ok := m.modules[moduleName].(appmodule.HasPreBlocker); ok {
				if _, err := module.PreBlock(ctx); err != nil {
					return fmt.Errorf("failed to run preblock for %s: %w", moduleName, err)
				}
			}
		}

		return nil
	}
}

// TxValidators validates incoming transactions
func (m *MMv2) TxValidation() func(ctx context.Context, tx transaction.Tx) error {
	// TODO rewrap the context into sdk.Context

	return func(ctx context.Context, tx transaction.Tx) error {
		for _, moduleName := range m.config.TxValidation {
			if module, ok := m.modules[moduleName].(appmodule.HasTxValidation[transaction.Tx]); ok {
				if err := module.TxValidator(ctx, tx); err != nil {
					return fmt.Errorf("failed to run txvalidator for %s: %w", moduleName, err)
				}
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
