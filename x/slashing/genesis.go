package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, stakingKeeper types.StakingKeeper, data types.GenesisState) {
	stakingKeeper.IterateValidators(ctx,
		func(index int64, validator exported.ValidatorI) bool {
			keeper.AddPubkey(ctx, validator.GetConsPubKey())
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
			keeper.SetValidatorMissedBlockBitArray(ctx, address, missed.Index, missed.Missed)
		}
	}

	keeper.SetParams(ctx, data.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) (data types.GenesisState) {
	params := keeper.GetParams(ctx)
	signingInfos := make(map[string]types.ValidatorSigningInfo)
	missedBlocks := make(map[string][]types.MissedBlock)
	keeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		bechAddr := address.String()
		signingInfos[bechAddr] = info
		localMissedBlocks := []types.MissedBlock{}

		keeper.IterateValidatorMissedBlockBitArray(ctx, address, func(index int64, missed bool) (stop bool) {
			localMissedBlocks = append(localMissedBlocks, types.NewMissedBlock(index, missed))
			return false
		})
		missedBlocks[bechAddr] = localMissedBlocks

		return false
	})

	return types.NewGenesisState(params, signingInfos, missedBlocks)
}
