package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
)

// BeginBlocker updates an existing localhost client with the latest block height.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	plan, found := k.GetUpgradePlan(ctx)
	if found {
		planHeight, found := k.GetUpgradePlanHeight(ctx)
		if found {
			// Once we are at the last block this chain will commit, set the upgraded consensus state
			// so that IBC clients can use the last NextValidatorsHash as a trusted kernel for verifying
			// headers on the next version of the chain.
			// Set the time to the last block time of the current chain.
			// In order for a client to upgrade successfully, the first block of the new chain must be committed
			// within the trusting period of the last block time on this chain.
			//
			// The upgrade plan must exist in x/upgrade, the plan must be an IBC upgrade, and the current height
			// must be one before the scheduled IBC upgrade.
			_, exists := k.GetUpgradedClient(ctx, int64(planHeight))
			if exists && ctx.BlockHeight() == int64(planHeight)-1 && ctx.BlockHeight() == plan.Height-1 {
				upgradedConsState := &ibctmtypes.ConsensusState{
					Timestamp:          ctx.BlockTime(),
					NextValidatorsHash: ctx.BlockHeader().NextValidatorsHash,
				}
				bz := k.MustMarshalConsensusState(upgradedConsState)

				k.SetUpgradedConsensusState(ctx, plan.Height, bz)
			}
		}
	}

	_, found = k.GetClientState(ctx, exported.Localhost)
	if !found {
		return
	}

	// update the localhost client with the latest block height
	if err := k.UpdateClient(ctx, exported.Localhost, nil); err != nil {
		panic(err)
	}
}
