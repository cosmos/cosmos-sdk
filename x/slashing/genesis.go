package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params          Params                          `json:"params"`
	SigningInfos    map[string]ValidatorSigningInfo `json:"signing_infos"`
	MissedBlocks    map[string][]MissedBlock        `json:"missed_blocks"`
	SlashingPeriods []ValidatorSlashingPeriod       `json:"slashing_periods"`
}

// MissedBlock
type MissedBlock struct {
	Index  int64 `json:"index"`
	Missed bool  `json:"missed"`
}

// HubDefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:          DefaultParams(),
		SigningInfos:    make(map[string]ValidatorSigningInfo),
		MissedBlocks:    make(map[string][]MissedBlock),
		SlashingPeriods: []ValidatorSlashingPeriod{},
	}
}

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState, sdata types.GenesisState) {
	for _, validator := range sdata.Validators {
		keeper.addPubkey(ctx, validator.GetConsPubKey())
	}

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

	for _, slashingPeriod := range data.SlashingPeriods {
		keeper.addOrUpdateValidatorSlashingPeriod(ctx, slashingPeriod)
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

	slashingPeriods := []ValidatorSlashingPeriod{}
	keeper.IterateValidatorSlashingPeriods(ctx, func(slashingPeriod ValidatorSlashingPeriod) (stop bool) {
		slashingPeriods = append(slashingPeriods, slashingPeriod)
		return false
	})

	return GenesisState{
		Params:          params,
		SigningInfos:    signingInfos,
		MissedBlocks:    missedBlocks,
		SlashingPeriods: slashingPeriods,
	}
}
