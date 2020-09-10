package types

import (
	"bytes"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

var _ exported.Header = Header{}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() exported.ClientType {
	return exported.Tendermint
}

// ConsensusState returns the updated consensus state associated with the header
func (h Header) ConsensusState() *ConsensusState {
	return &ConsensusState{
		Height:             h.GetHeight().(clienttypes.Height),
		Timestamp:          h.GetTime(),
		Root:               commitmenttypes.NewMerkleRoot(h.Header.GetAppHash()),
		NextValidatorsHash: h.Header.NextValidatorsHash,
	}
}

// GetHeight returns the current height. It returns 0 if the tendermint
// header is nil.
//
// TODO: return clienttypes.Height once interface changes
func (h Header) GetHeight() exported.Height {
	if h.Header == nil {
		return clienttypes.ZeroHeight()
	}

	// Enforce clienttypes.Height to use 0 epoch number
	// TODO: Retrieve epoch number from chain-id
	return clienttypes.NewHeight(0, uint64(h.Header.Height))
}

// GetTime returns the current block timestamp. It returns a zero time if
// the tendermint header is nil.
func (h Header) GetTime() time.Time {
	if h.Header == nil {
		return time.Time{}
	}
	return h.Header.Time
}

// ValidateBasic calls the SignedHeader ValidateBasic function and checks
// that validatorsets are not nil.
// NOTE: TrustedHeight and TrustedValidators may be empty when creating client
// with MsgCreateClient
func (h Header) ValidateBasic() error {
	if h.SignedHeader == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "tendermint signed header cannot be nil")
	}
	if h.Header == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "tendermint header cannot be nil")
	}
	tmSignedHeader, err := tmtypes.SignedHeaderFromProto(h.SignedHeader)
	if err != nil {
		return sdkerrors.Wrap(err, "header is not a tendermint header")
	}
	if err := tmSignedHeader.ValidateBasic(h.Header.GetChainID()); err != nil {
		return sdkerrors.Wrap(err, "header failed basic validation")
	}

	// TrustedHeight is less than Header for updates
	// and less than or equal to Header for misbehaviour
	if h.TrustedHeight.GT(h.GetHeight()) {
		return sdkerrors.Wrapf(ErrInvalidHeaderHeight, "TrustedHeight %d must be less than or equal to header height %d",
			h.TrustedHeight, h.GetHeight())
	}

	if h.ValidatorSet == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set is nil")
	}
	tmValset, err := tmtypes.ValidatorSetFromProto(h.ValidatorSet)
	if err != nil {
		return sdkerrors.Wrap(err, "validator set is not tendermint validator set")
	}
	if !bytes.Equal(h.Header.ValidatorsHash, tmValset.Hash()) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set does not match hash")
	}
	return nil
}
