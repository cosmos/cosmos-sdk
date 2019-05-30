package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper Keeper, stakingKeeper StakingKeeper, data GenesisState) {
	stakingKeeper.IterateValidators(ctx,
		func(index int64, validator sdk.Validator) bool {
			keeper.addPubkey(ctx, validator.GetConsPubKey())
			return false
		},
	)

	for addr, info := range data.SigningInfos {
		address, err := sdk.ConsAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}
		keeper.SetValidatorSigningInfo(ctx, address, info)
	}

	for addr, array := range data.MissedBlocks {
		address, err := sdk.ConsAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}
		for _, missed := range array {
			keeper.setValidatorMissedBlockBitArray(ctx, address, missed.Index, missed.Missed)
		}
	}

	keeper.paramspace.SetParamSet(ctx, &data.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper Keeper) (data GenesisState) {
	var params Params
	keeper.paramspace.GetParamSet(ctx, &params)

	signingInfos := make(map[string]ValidatorSigningInfo)
	missedBlocks := make(map[string][]MissedBlock)
	keeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info ValidatorSigningInfo) (stop bool) {
		bechAddr := address.String()
		signingInfos[bechAddr] = info
		localMissedBlocks := []MissedBlock{}

		keeper.IterateValidatorMissedBlockBitArray(ctx, address, func(index int64, missed bool) (stop bool) {
			localMissedBlocks = append(localMissedBlocks, MissedBlock{index, missed})
			return false
		})
		missedBlocks[bechAddr] = localMissedBlocks

		return false
	})

	return GenesisState{
		Params:       params,
		SigningInfos: signingInfos,
		MissedBlocks: missedBlocks,
	}
}
