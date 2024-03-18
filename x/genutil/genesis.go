package genutil

import (
	"context"

	"cosmossdk.io/core/genesis"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// InitGenesis - initialize accounts and deliver genesis transactions
func InitGenesis(
	ctx context.Context, stakingKeeper types.StakingKeeper,
	deliverTx genesis.TxHandler, genesisState types.GenesisState,
	txEncodingConfig client.TxEncodingConfig,
) (validatorUpdates []module.ValidatorUpdate, err error) {
	if len(genesisState.GenTxs) > 0 {
		moduleValidatorUpdates, err := DeliverGenTxs(ctx, genesisState.GenTxs, stakingKeeper, deliverTx, txEncodingConfig)
		if err != nil {
			return nil, err
		}
		return moduleValidatorUpdates, nil
	}
	return
}
