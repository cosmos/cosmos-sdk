package v040

import (
	"sort"

	v039slashing "github.com/cosmos/cosmos-sdk/x/slashing/legacy/v0_39"
)

// Migrate accepts exported x/slashing genesis state from v0.39 and migrates it
// to v0.40 x/slashing genesis state. The migration includes:
//
// - Chaning SigningInfos and MissedBlocks from map to array.
func Migrate(oldGenState v039slashing.GenesisState) *GenesisState {
	// Note that the two following `for` loop over a map's keys, so are not
	// deterministic.
	var newSigningInfos = make([]SigningInfo, 0, len(oldGenState.SigningInfos))
	for address, signingInfo := range oldGenState.SigningInfos {
		newSigningInfos = append(newSigningInfos, SigningInfo{
			Address: address,
			ValidatorSigningInfo: ValidatorSigningInfo{
				Address:             signingInfo.Address,
				StartHeight:         signingInfo.StartHeight,
				IndexOffset:         signingInfo.IndexOffset,
				JailedUntil:         signingInfo.JailedUntil,
				Tombstoned:          signingInfo.Tombstoned,
				MissedBlocksCounter: signingInfo.MissedBlocksCounter,
			},
		})
	}
	var newValidatorMissedBlocks = make([]ValidatorMissedBlocks, 0, len(oldGenState.MissedBlocks))
	for address, validatorMissedBlocks := range oldGenState.MissedBlocks {
		var newMissedBlocks = make([]MissedBlock, 0, len(validatorMissedBlocks))
		for _, missedBlock := range validatorMissedBlocks {
			newMissedBlocks = append(newMissedBlocks, MissedBlock{
				Index:  missedBlock.Index,
				Missed: missedBlock.Missed,
			})
		}

		newValidatorMissedBlocks = append(newValidatorMissedBlocks, ValidatorMissedBlocks{
			Address:      address,
			MissedBlocks: newMissedBlocks,
		})
	}

	// We sort these two arrays by address, so that we get determinstic states.
	sort.Slice(newSigningInfos, func(i, j int) bool { return newSigningInfos[i].Address < newSigningInfos[j].Address })
	sort.Slice(newValidatorMissedBlocks, func(i, j int) bool { return newValidatorMissedBlocks[i].Address < newValidatorMissedBlocks[j].Address })

	return &GenesisState{
		Params: Params{
			SignedBlocksWindow:      oldGenState.Params.SignedBlocksWindow,
			MinSignedPerWindow:      oldGenState.Params.MinSignedPerWindow,
			DowntimeJailDuration:    oldGenState.Params.DowntimeJailDuration,
			SlashFractionDoubleSign: oldGenState.Params.SlashFractionDoubleSign,
			SlashFractionDowntime:   oldGenState.Params.SlashFractionDowntime,
		},
		SigningInfos: newSigningInfos,
		MissedBlocks: newValidatorMissedBlocks,
	}
}
