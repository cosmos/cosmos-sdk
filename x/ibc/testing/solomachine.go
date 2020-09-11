package ibctesting

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/std"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

// Solomachine is a testing helper used to simulate a counterparty
// solo machine client.
type Solomachine struct {
	t *testing.T

	cdc         codec.BinaryMarshaler
	ClientID    string
	PrivateKey  crypto.PrivKey
	PublicKey   crypto.PubKey
	Sequence    uint64
	Time        uint64
	Diversifier string
}

// NewSolomachine returns a new solomachine instance with a generated private/public
// key pair and a sequence starting at 1.
func NewSolomachine(t *testing.T, cdc codec.BinaryMarshaler, clientID, diversifier string) *Solomachine {
	privKey := ed25519.GenPrivKey()

	return &Solomachine{
		t:           t,
		cdc:         cdc,
		ClientID:    clientID,
		PrivateKey:  privKey,
		PublicKey:   privKey.PubKey(),
		Sequence:    1,
		Time:        10,
		Diversifier: diversifier,
	}
}

// ClientState returns a new solo machine ClientState instance. Default usage does not allow update
// after governance proposal
func (solo *Solomachine) ClientState() *solomachinetypes.ClientState {
	return solomachinetypes.NewClientState(solo.Sequence, solo.ConsensusState(), false)
}

// ConsensusState returns a new solo machine ConsensusState instance
func (solo *Solomachine) ConsensusState() *solomachinetypes.ConsensusState {
	publicKey, err := std.DefaultPublicKeyCodec{}.Encode(solo.PublicKey)
	require.NoError(solo.t, err)

	return &solomachinetypes.ConsensusState{
		PublicKey:   publicKey,
		Diversifier: solo.Diversifier,
		Timestamp:   solo.Time,
	}
}

// GetHeight returns an exported.Height with Sequence as EpochHeight
func (solo *Solomachine) GetHeight() exported.Height {
	return clienttypes.NewHeight(0, solo.Sequence)
}

// CreateHeader generates a new private/public key pair and creates the
// necessary signature to construct a valid solo machine header.
func (solo *Solomachine) CreateHeader() *solomachinetypes.Header {
	// generate new private key and signature for header
	newPrivKey := ed25519.GenPrivKey()

	publicKey, err := std.DefaultPublicKeyCodec{}.Encode(newPrivKey.PubKey())
	require.NoError(solo.t, err)

	data := &solomachinetypes.HeaderData{
		NewPubKey:      publicKey,
		NewDiversifier: solo.Diversifier,
	}

	dataBz, err := solo.cdc.MarshalBinaryBare(data)
	require.NoError(solo.t, err)

	signBytes := &solomachinetypes.SignBytes{
		Sequence:    solo.Sequence,
		Timestamp:   solo.Time,
		Diversifier: solo.Diversifier,
		Data:        dataBz,
	}

	signBz, err := solo.cdc.MarshalBinaryBare(signBytes)
	require.NoError(solo.t, err)

	signature, err := solo.PrivateKey.Sign(signBz)
	require.NoError(solo.t, err)

	header := &solomachinetypes.Header{
		Sequence:       solo.Sequence,
		Timestamp:      solo.Time,
		Signature:      signature,
		NewPublicKey:   publicKey,
		NewDiversifier: solo.Diversifier,
	}

	// assumes successful header update
	solo.Sequence++
	solo.PrivateKey = newPrivKey
	solo.PublicKey = newPrivKey.PubKey()

	return header
}

// CreateMisbehaviour constructs testing misbehaviour for the solo machine client
// by signing over two different data bytes at the same sequence.
func (solo *Solomachine) CreateMisbehaviour() *solomachinetypes.Misbehaviour {
	dataOne := []byte("DATA ONE")
	dataTwo := []byte("DATA TWO")

	signBytes := &solomachinetypes.SignBytes{
		Sequence:    solo.Sequence,
		Timestamp:   solo.Time,
		Diversifier: solo.Diversifier,
		Data:        dataOne,
	}

	signBz, err := solo.cdc.MarshalBinaryBare(signBytes)
	require.NoError(solo.t, err)

	sig, err := solo.PrivateKey.Sign(signBz)
	require.NoError(solo.t, err)

	signatureOne := solomachinetypes.SignatureAndData{
		Signature: sig,
		Data:      dataOne,
	}

	signBytes = &solomachinetypes.SignBytes{
		Sequence:    solo.Sequence,
		Timestamp:   solo.Time,
		Diversifier: solo.Diversifier,
		Data:        dataTwo,
	}

	signBz, err = solo.cdc.MarshalBinaryBare(signBytes)
	require.NoError(solo.t, err)

	sig, err = solo.PrivateKey.Sign(signBz)
	require.NoError(solo.t, err)

	signatureTwo := solomachinetypes.SignatureAndData{
		Signature: sig,
		Data:      dataTwo,
	}

	return &solomachinetypes.Misbehaviour{
		ClientId:     solo.ClientID,
		Sequence:     solo.Sequence,
		SignatureOne: &signatureOne,
		SignatureTwo: &signatureTwo,
	}
}
