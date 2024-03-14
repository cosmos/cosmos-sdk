package genutil

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/module"

	"cosmossdk.io/core/genesis"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// InitGenesis - initialize accounts and deliver genesis transactions
func InitGenesis(
	ctx context.Context, stakingKeeper types.StakingKeeper,
	deliverTx genesis.TxHandler, genesisState types.GenesisState,
	txEncodingConfig client.TxEncodingConfig,
) (validatorUpdates []module.ValidatorUpdate, err error) {
	if len(genesisState.GenTxs) > 0 {
		cometValidatorUpdates, err := DeliverGenTxs(ctx, genesisState.GenTxs, stakingKeeper, deliverTx, txEncodingConfig)
		if err != nil {
			return nil, err
		}
		validatorUpdates = make([]module.ValidatorUpdate, len(cometValidatorUpdates))
		for i, v := range cometValidatorUpdates {
			if ed25519 := v.PubKey.GetEd25519(); len(ed25519) > 0 {
				validatorUpdates[i] = module.ValidatorUpdate{
					PubKey:     ed25519,
					PubKeyType: "ed25519",
					Power:      v.Power,
				}
			} else if secp256k1 := v.PubKey.GetSecp256K1(); len(secp256k1) > 0 {
				validatorUpdates[i] = module.ValidatorUpdate{
					PubKey:     secp256k1,
					PubKeyType: "secp256k1",
					Power:      v.Power,
				}
			} else {
				return nil, fmt.Errorf("unexpected validator pubkey type: %T", v.PubKey)
			}
		}
		return validatorUpdates, nil
	}
	return
}
