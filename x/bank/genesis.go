package bank

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is the bank state that must be provided at genesis.
type GenesisState struct {
	SendEnabled                bool    `json:"send_enabled"`
	SacrificialSendBurnPercent sdk.Dec `json:"sacficial_send_burn_percent"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(sendEnabled bool, sacrificialPercent sdk.Dec) GenesisState {
	return GenesisState{SendEnabled: sendEnabled, SacrificialSendBurnPercent: sacrificialPercent}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(true, sdk.NewDecFromIntWithPrec(sdk.NewInt(9), 1))
}

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetSendEnabled(ctx, data.SendEnabled)
	keeper.SetSacrificialSendBurnPercent(ctx, data.SacrificialSendBurnPercent)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(keeper.GetSendEnabled(ctx), keeper.GetSacrificialSendBurnPercent(ctx))
}

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if data.SacrificialSendBurnPercent.LT(sdk.ZeroDec()) || data.SacrificialSendBurnPercent.GT(sdk.OneDec()) {
		return errors.New("Sacrificial Send Burn Percent must be between 0 and 1")
	}
	return nil
}
