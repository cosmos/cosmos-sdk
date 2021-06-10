package v040

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// ValidateBasic does basic validation of a Plan
func (p Plan) ValidateBasic() error {
	if len(p.Name) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}
	if p.Height <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "height must be greater than 0")
	}

	return nil
}
