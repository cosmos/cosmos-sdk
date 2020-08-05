package types

import (
	"bytes"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

var _ clientexported.Header = Header{}

// Header defines the Tendermint consensus Header
// Header encapsulates all the information necessary to update from a trusted Tendermint ConsensusState
// The SignedHeader and ValidatorSet are the new untrusted update fields for the client
// The TrustedHeight is the height of a stored ConsensusState on the client that will be used to verify new untrusted header
// The Trusted ConsensusState must be within the unbonding period of current time in order to correctly verify,
// and the TrustedValidators must hash to TrustedConsensusState.NextValidatorsHash since that is the last trusted validatorset
// at the TrustedHeight
type Header struct {
	tmtypes.SignedHeader `json:"signed_header" yaml:"signed_header"` // contains the commitment root
	ValidatorSet         *tmtypes.ValidatorSet                       `json:"validator_set" yaml:"validator_set"`   // the validator set that signed Header
	TrustedHeight        uint64                                      `json:"trusted_height" yaml:"trusted_height"` // the height of a trusted header seen by client less than or equal to Header
	TrustedValidators    *tmtypes.ValidatorSet                       `json:"trusted_vals" yaml:"trusted_vals"`     // the last trusted validator set at trusted height
}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// ConsensusState returns the updated consensus state associated with the header
func (h Header) ConsensusState() ConsensusState {
	return ConsensusState{
		Height:             uint64(h.Height),
		Timestamp:          h.Time,
		Root:               commitmenttypes.NewMerkleRoot(h.AppHash),
		NextValidatorsHash: h.NextValidatorsHash,
	}
}

// GetHeight returns the current height
//
// NOTE: also referred as `sequence`
func (h Header) GetHeight() uint64 {
	return uint64(h.Height)
}

// ValidateBasic calls the SignedHeader ValidateBasic function
// and checks that validatorsets are not nil
// NOTE: TrustedHeight and TrustedValidators may be empty when creating client
// with MsgCreateClient
func (h Header) ValidateBasic(chainID string) error {
	if err := h.SignedHeader.ValidateBasic(chainID); err != nil {
		return sdkerrors.Wrap(err, "header failed basic validation")
	}
	// TrustedHeight is less than Header for updates
	// and less than or equal to Header for misbehaviour
	if h.TrustedHeight > uint64(h.Height) {
		return sdkerrors.Wrapf(ErrInvalidHeaderHeight, "TrustedHeight %d must be less than or equal to header height %d",
			h.TrustedHeight, h.Height)
	}
	if h.ValidatorSet == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set is nil")
	}
	if !bytes.Equal(h.ValidatorsHash, h.ValidatorSet.Hash()) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set does not match hash")
	}
	return nil
}

// ToABCIHeader parses the header to an ABCI header type.
// NOTE: only for testing use.
func (h Header) ToABCIHeader() abci.Header {
	return tmtypes.TM2PB.Header(h.SignedHeader.Header)
}
