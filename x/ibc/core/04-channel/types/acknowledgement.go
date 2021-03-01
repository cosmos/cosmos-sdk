package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewResultAcknowledgement returns a new instance of Acknowledgement using an Acknowledgement_Result
// type in the Response field.
func NewResultAcknowledgement(result []byte) Acknowledgement {
	return Acknowledgement{
		Response: &Acknowledgement_Result{
			Result: result,
		},
	}
}

// NewErrorAcknowledgement returns a new instance of Acknowledgement using an Acknowledgement_Error
// type in the Response field.
func NewErrorAcknowledgement(err string) Acknowledgement {
	return Acknowledgement{
		Response: &Acknowledgement_Error{
			Error: err,
		},
	}
}

// GetBytes is a helper for serialising acknowledgements
func (ack Acknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&ack))
}

// ValidateBasic performs a basic validation of the acknowledgement
func (ack Acknowledgement) ValidateBasic() error {
	switch resp := ack.Response.(type) {
	case *Acknowledgement_Result:
		if len(resp.Result) == 0 {
			return sdkerrors.Wrap(ErrInvalidAcknowledgement, "acknowledgement result cannot be empty")
		}
	case *Acknowledgement_Error:
		if strings.TrimSpace(resp.Error) == "" {
			return sdkerrors.Wrap(ErrInvalidAcknowledgement, "acknowledgement error cannot be empty")
		}
	default:
		return sdkerrors.Wrapf(ErrInvalidAcknowledgement, "unsupported acknowledgement response field type %T", resp)
	}
	return nil
}
