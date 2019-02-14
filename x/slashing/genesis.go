package slashing

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params       Params                          `json:"params"`
	SigningInfos map[string]ValidatorSigningInfo `json:"signing_infos"`
	MissedBlocks map[string][]MissedBlock        `json:"missed_blocks"`
}

// MissedBlock
type MissedBlock struct {
	Index  int64 `json:"index"`
	Missed bool  `json:"missed"`
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:       DefaultParams(),
		SigningInfos: make(map[string]ValidatorSigningInfo),
		MissedBlocks: make(map[string][]MissedBlock),
	}
}

// ValidateGenesis validates the slashing genesis parameters
func ValidateGenesis(data GenesisState) error {
	downtime := data.Params.SlashFractionDowntime
	if downtime.IsNegative() || downtime.GT(sdk.OneDec()) {
		return fmt.Errorf("Slashing fraction downtime should be less than or equal to one and greater than zero, is %s", downtime.String())
	}

	dblSign := data.Params.SlashFractionDoubleSign
	if dblSign.IsNegative() || dblSign.GT(sdk.OneDec()) {
		return fmt.Errorf("Slashing fraction double sign should be less than or equal to one and greater than zero, is %s", dblSign.String())
	}

	minSign := data.Params.MinSignedPerWindow
	if minSign.IsNegative() || minSign.GT(sdk.OneDec()) {
		return fmt.Errorf("Min signed per window should be less than or equal to one and greater than zero, is %s", minSign.String())
	}

	maxEvidence := data.Params.MaxEvidenceAge
	if maxEvidence < 1*time.Minute {
		return fmt.Errorf("Max evidence age must be at least 1 minute, is %s", maxEvidence.String())
	}

	downtimeJail := data.Params.DowntimeJailDuration
	if downtimeJail < 1*time.Minute {
		return fmt.Errorf("Downtime unblond duration must be at least 1 minute, is %s", downtimeJail.String())
	}

	signedWindow := data.Params.SignedBlocksWindow
	if signedWindow < 10 {
		return fmt.Errorf("Signed blocks window must be at least 10, is %d", signedWindow)
	}

	return nil
}

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState, validators []sdk.Validator) {
	for _, validator := range validators {
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
