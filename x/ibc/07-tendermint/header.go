package tendermint

import (
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienterrors "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
)

var _ exported.Header = Header{}

// Header defines the Tendermint consensus Header
type Header struct {
	tmtypes.SignedHeader                       // contains the commitment root
	ValidatorSet         *tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() exported.ClientType {
	return exported.Tendermint
}

// GetHeight returns the current height
//
// NOTE: also referred as `sequence`
func (h Header) GetHeight() uint64 {
	return uint64(h.Height)
}

// ValidateBasic calls the SignedHeader ValidateBasic function
// and checks that validatorsets are not nil
func (h Header) ValidateBasic(chainID string) error {
	if err := h.SignedHeader.ValidateBasic(chainID); err != nil {
		return sdkerrors.Wrap(clienterrors.ErrInvalidHeader, err.Error())
	}
	if h.ValidatorSet == nil {
		return sdkerrors.Wrap(clienterrors.ErrInvalidHeader, "validator set is nil")
	}
	return nil
}
