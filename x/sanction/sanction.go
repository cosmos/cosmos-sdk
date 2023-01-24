package sanction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Define the defaults for each param field and allow consuming apps to set them.
var (
	// DefaultImmediateSanctionMinDeposit is the default to use for the MinDepositSanction.
	DefaultImmediateSanctionMinDeposit sdk.Coins = nil

	// DefaultImmediateUnsanctionMinDeposit is the default to use for the MinDepositUnsanction.
	DefaultImmediateUnsanctionMinDeposit sdk.Coins = nil
)

func DefaultParams() *Params {
	return &Params{
		ImmediateSanctionMinDeposit:   DefaultImmediateSanctionMinDeposit,
		ImmediateUnsanctionMinDeposit: DefaultImmediateUnsanctionMinDeposit,
	}
}

func (p Params) ValidateBasic() error {
	if err := p.ImmediateSanctionMinDeposit.Validate(); err != nil {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid immediate sanction min deposit: %s", err.Error())
	}
	if err := p.ImmediateUnsanctionMinDeposit.Validate(); err != nil {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid immediate unsanction min deposit: %s", err.Error())
	}
	return nil
}
