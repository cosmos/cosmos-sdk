package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper Keeper, stakingKeeper types.StakingKeeper, data types.GenesisState) {
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

		voteArray := keeper.GetVoteArray(ctx, address)
		if voteArray == nil {
			voteArray = types.NewVoteArray(data.Params.SignedBlocksWindow)
		}

		for _, missed := range array {
			if missed.Index >= data.Params.SignedBlocksWindow {
				// this could happen only if params got changed and SignedBlocksWindow is now lower
				// then it was during ExportGenesis
				continue
			}
			if missed.Missed {
				voteArray.Get(missed.Index).Miss()
			}
		}
		keeper.SetVoteArray(ctx, address, voteArray)
	}

	keeper.SetParams(ctx, data.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper Keeper) (data types.GenesisState) {
	params := keeper.GetParams(ctx)
	signingInfos := make(map[string]types.ValidatorSigningInfo)
	missedBlocks := make(map[string][]types.MissedBlock)
	keeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		bechAddr := address.String()
		signingInfos[bechAddr] = info
		localMissedBlocks := []types.MissedBlock{}

		voteArray := keeper.GetVoteArray(ctx, address)
		// no missed block - terminate early
		if voteArray == nil {
			return false
		}
		for i := int64(0); i < params.SignedBlocksWindow; i++ {
			vote := voteArray.Get(i)
			// not missed blocks are skipped in InitGenesis, so we skip them here as well
			if vote.Missed() {
				localMissedBlocks = append(localMissedBlocks, types.MissedBlock{Index: i, Missed: true})
			}
		}
		missedBlocks[bechAddr] = localMissedBlocks

		return false
	})

	return types.NewGenesisState(params, signingInfos, missedBlocks)
}
