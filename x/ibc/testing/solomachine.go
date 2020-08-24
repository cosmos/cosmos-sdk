package testing

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

// Solomachine is a testing helper used to simulate a counterparty
// solo machine client.
type Solomachine struct {
	t *testing.T

	ClientID   string
	PrivateKey crypto.PrivKey
	PublicKey  crypto.PubKey
	Sequence   uint64
	Time       uint64
}

// NewSolomachine returns a new solomachine instance with a generated private/public
// key pair and a sequence starting at 1.
func NewSolomachine(t *testing.T, clientID string) *Solomachine {
	privKey := ed25519.GenPrivKey()

	return &Solomachine{
		t:          t,
		ClientID:   clientID,
		PrivateKey: privKey,
		PublicKey:  privKey.PubKey(),
		Sequence:   1,
		Time:       10,
	}
}

func (solo *Solomachine) ClientState() *solomachinetypes.ClientState {
	return solomachinetypes.NewClientState(solo.ConsensusState())
}

func (solo *Solomachine) ConsensusState() *solomachinetypes.ConsensusState {
	publicKey, err := std.DefaultPublicKeyCodec{}.Encode(solo.PublicKey)
	require.NoError(solo.t, err)

	return &solomachinetypes.ConsensusState{
		Sequence:  solo.Sequence,
		PublicKey: publicKey,
		Timestamp: solo.Time,
	}
}

// CreateHeader generates a new private/public key pair and creates the
// necessary signature to construct a valid solo machine header.
func (solo *Solomachine) CreateHeader() solomachinetypes.Header {
	// generate new private key and signature for header
	newPrivKey := ed25519.GenPrivKey()
	data := append(sdk.Uint64ToBigEndian(solo.Sequence), newPrivKey.PubKey().Bytes()...)
	signature, err := solo.PrivateKey.Sign(data)
	require.NoError(solo.t, err)

	publicKey, err := std.DefaultPublicKeyCodec{}.Encode(newPrivKey.PubKey())
	require.NoError(solo.t, err)

	header := solomachinetypes.Header{
		Sequence:     solo.Sequence,
		Signature:    signature,
		NewPublicKey: publicKey,
	}

	// assumes successful header update
	solo.Sequence++
	solo.PrivateKey = newPrivKey
	solo.PublicKey = newPrivKey.PubKey()

	return header
}

// CreateMisbehaviour constructs testing evidence for the solo machine client
// by signing over two different data bytes at the same sequence.
func (solo *Solomachine) CreateMisbehaviour() solomachinetypes.Misbehaviour {
	dataOne := []byte("DATA ONE")
	dataTwo := []byte("DATA TWO")

	sig, err := solo.PrivateKey.Sign(append(sdk.Uint64ToBigEndian(solo.Sequence), dataOne...))
	require.NoError(solo.t, err)

	signatureOne := solomachinetypes.SignatureAndData{
		Signature: sig,
		Data:      dataOne,
	}

	sig, err = solo.PrivateKey.Sign(append(sdk.Uint64ToBigEndian(solo.Sequence), dataTwo...))
	require.NoError(solo.t, err)

	signatureTwo := solomachinetypes.SignatureAndData{
		Signature: sig,
		Data:      dataTwo,
	}

	return solomachinetypes.Misbehaviour{
		ClientId:     solo.ClientID,
		Sequence:     solo.Sequence,
		SignatureOne: &signatureOne,
		SignatureTwo: &signatureTwo,
	}
}
