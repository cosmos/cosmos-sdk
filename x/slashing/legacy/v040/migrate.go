package v040

import (
	"sort"

	v039slashing "github.com/cosmos/cosmos-sdk/x/slashing/legacy/v039"
	v040slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Migrate accepts exported x/slashing genesis state from v0.39 and migrates it
// to v0.40 x/slashing genesis state. The migration includes:
//
// - Chaning SigningInfos and MissedBlocks from map to array.
// - Convert addresses from bytes to bech32 strings.
// - Re-encode in v0.40 GenesisState.
func Migrate(oldGenState v039slashing.GenesisState) *v040slashing.GenesisState {
	// Note that the two following `for` loop over a map's keys, so are not
	// deterministic.
	var newSigningInfos = make([]v040slashing.SigningInfo, 0, len(oldGenState.SigningInfos))
	for address, signingInfo := range oldGenState.SigningInfos {
		newSigningInfos = append(newSigningInfos, v040slashing.SigningInfo{
			Address: address,
			ValidatorSigningInfo: v040slashing.ValidatorSigningInfo{
				Address:             signingInfo.Address.String(),
				StartHeight:         signingInfo.StartHeight,
				IndexOffset:         signingInfo.IndexOffset,
				JailedUntil:         signingInfo.JailedUntil,
				Tombstoned:          signingInfo.Tombstoned,
				MissedBlocksCounter: signingInfo.MissedBlocksCounter,
			},
		})
	}
	var newValidatorMissedBlocks = make([]v040slashing.ValidatorMissedBlocks, 0, len(oldGenState.MissedBlocks))
	for address, validatorMissedBlocks := range oldGenState.MissedBlocks {
		var newMissedBlocks = make([]v040slashing.MissedBlock, len(validatorMissedBlocks))
		for i, missedBlock := range validatorMissedBlocks {
			newMissedBlocks[i] = v040slashing.MissedBlock{
				Index:  missedBlock.Index,
				Missed: missedBlock.Missed,
			}
		}

		newValidatorMissedBlocks = append(newValidatorMissedBlocks, v040slashing.ValidatorMissedBlocks{
			Address:      address,
			MissedBlocks: newMissedBlocks,
		})
	}

	// We sort these two arrays by address, so that we get determinstic states.
	sort.Slice(newSigningInfos, func(i, j int) bool { return newSigningInfos[i].Address < newSigningInfos[j].Address })
	sort.Slice(newValidatorMissedBlocks, func(i, j int) bool { return newValidatorMissedBlocks[i].Address < newValidatorMissedBlocks[j].Address })

	return &v040slashing.GenesisState{
		Params: v040slashing.Params{
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
