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
	tmtypes.SignedHeader
	ValidatorSet     *tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
	NextValidatorSet *tmtypes.ValidatorSet `json:"next_validator_set" yaml:"next_validator_set"`
}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() exported.ClientType {
	return exported.Tendermint
}

// GetCommitter returns the ValidatorSet that committed header
func (h Header) GetCommitter() exported.Committer {
	return Committer{h.ValidatorSet}
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
		return err
	}
	if h.ValidatorSet == nil {
		return sdkerrors.Wrap(clienterrors.ErrInvalidHeader(clienterrors.DefaultCodespace), "ValidatorSet is nil")
	}
	if h.NextValidatorSet == nil {
		return sdkerrors.Wrap(clienterrors.ErrInvalidHeader(clienterrors.DefaultCodespace), "NextValidatorSet is nil")
	}
	return nil
}
