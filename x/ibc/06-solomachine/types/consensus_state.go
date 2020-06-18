package types

import (
	tmcrypto "github.com/tendermint/tendermint/crypto"

	//	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

var _ clientexported.ConsensusState = ConsensusState{}

// ClientType returns Solo Machine type
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetHeight returns the sequence number
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Sequence
}

// GetTimestamp returns zero.
func (cs ConsensusState) GetTimestamp() uint64 {
	return cs.Timestamp
}

// GetRoot returns nil as solo machines do not have roots
func (cs ConsensusState) GetRoot() commitmentexported.Root {
	return nil
}

// GetPubKey unmarshals the public key into a crypto.PubKey type.
func (cs ConsensusState) GetPubKey() tmcrypto.PubKey {
	//	if len(cs.PubKey) == 0 {
	//		return nil
	//	}

	pubKey, err := std.DefaultPublicKeyCodec{}.Decode(cs.PubKey)
	if err != nil {
		panic(err)
	}

	return pubKey
}

// ValidateBasic defines basic validation for the solo machine consensus state.
func (cs ConsensusState) ValidateBasic() error {
	if cs.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "sequence cannot be 0")
	}
	if cs.Timestamp == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp cannot be 0")
	}
	if cs.PubKey == nil || len(cs.GetPubKey().Bytes()) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "public key cannot be empty")
	}

	return nil
}
