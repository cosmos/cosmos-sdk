package types

import (
	"strings"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/std"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

var _ exported.ConsensusState = ConsensusState{}

// ClientType returns Solo Machine type.
func (ConsensusState) ClientType() exported.ClientType {
	return exported.SoloMachine
}

// GetHeight satisfies the ConsensusState interface
// NOTE: this function will be deprecated.
func (cs ConsensusState) GetHeight() exported.Height {
	return clienttypes.Height{}
}

// GetTimestamp returns zero.
func (cs ConsensusState) GetTimestamp() uint64 {
	return cs.Timestamp
}

// GetRoot returns nil since solo machines do not have roots.
func (cs ConsensusState) GetRoot() exported.Root {
	return nil
}

// GetPubKey unmarshals the public key into a tmcrypto.PubKey type.
func (cs ConsensusState) GetPubKey() tmcrypto.PubKey {
	publicKey, err := std.DefaultPublicKeyCodec{}.Decode(cs.PublicKey)
	if err != nil {
		panic(err)
	}

	return publicKey
}

// ValidateBasic defines basic validation for the solo machine consensus state.
func (cs ConsensusState) ValidateBasic() error {
	if cs.Timestamp == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp cannot be 0")
	}
	if cs.Diversifier != "" && strings.TrimSpace(cs.Diversifier) == "" {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "diversifier cannot contain only spaces")
	}
	if cs.PublicKey == nil || cs.GetPubKey() == nil || len(cs.GetPubKey().Bytes()) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "public key cannot be empty")
	}

	return nil
}
