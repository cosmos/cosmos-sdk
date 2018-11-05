package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params          Params
	SigningInfos    map[string]ValidatorSigningInfo
	SlashingPeriods []ValidatorSlashingPeriod
}

// HubDefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:          DefaultParams(),
		SigningInfos:    make(map[string]ValidatorSigningInfo),
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
		keeper.setValidatorSigningInfo(ctx, address, info)
	}

	for _, slashingPeriod := range data.SlashingPeriods {
		keeper.addOrUpdateValidatorSlashingPeriod(ctx, slashingPeriod)
	}

	keeper.paramspace.SetParamSet(ctx, &data.Params)
}

func WriteGenesis(ctx sdk.Context, keeper Keeper) (data GenesisState) {
	var params Params
	keeper.paramspace.GetParamSet(ctx, &params)

	signingInfos := make(map[string]ValidatorSigningInfo)
	keeper.iterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info ValidatorSigningInfo) (stop bool) {
		signingInfos[address.String()] = info
		return false
	})

	slashingPeriods := []ValidatorSlashingPeriod{}
	keeper.iterateValidatorSlashingPeriods(ctx, func(slashingPeriod ValidatorSlashingPeriod) (stop bool) {
		slashingPeriods = append(slashingPeriods, slashingPeriod)
		return false
	})

	return GenesisState{
		Params:          params,
		SigningInfos:    signingInfos,
		SlashingPeriods: slashingPeriods,
	}
}
