package types

import (
	"github.com/tendermint/tendermint/crypto"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

var _ clientexported.ConsensusState = ConsensusState{}

// ClientType returns Solo Machine type.
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetHeight returns the sequence number.
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Sequence
}

// GetTimestamp returns zero.
func (cs ConsensusState) GetTimestamp() uint64 {
	return cs.Timestamp
}

// GetRoot returns nil since solo machines do not have roots.
func (cs ConsensusState) GetRoot() commitmentexported.Root {
	return nil
}

// GetPubKey unmarshals the public key into a tmcrypto.PubKey type.
func (cs ConsensusState) GetPubKey() tmcrypto.PubKey {
	publicKey, ok := cs.PublicKey.GetCachedValue().(crypto.PubKey)
	if !ok {
		panic("ConsensusState PublicKey is not crypto.PubKey")
	}

	return publicKey
}

// ValidateBasic defines basic validation for the solo machine consensus state.
func (cs ConsensusState) ValidateBasic() error {
	if cs.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "sequence cannot be 0")
	}
	if cs.Timestamp == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp cannot be 0")
	}
	if cs.PublicKey == nil || cs.GetPubKey() == nil || len(cs.GetPubKey().Bytes()) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "public key cannot be empty")
	}

	return nil
}
