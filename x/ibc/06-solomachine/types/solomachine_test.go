package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

type SoloMachineTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	aminoCdc  *codec.Codec
	cdc       codec.Marshaler
	store     sdk.KVStore
	privKey   crypto.PrivKey
	sequence  uint64
	clientID  string
	timestamp uint64
}

func (suite *SoloMachineTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.aminoCdc = app.Codec()
	suite.cdc = app.AppCodec()
	suite.privKey = ed25519.GenPrivKey()

	suite.sequence = 1
	suite.timestamp = 10
	suite.clientID = "solomachineclient"
	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1, Time: time.Now()})
	suite.store = app.IBCKeeper.ClientKeeper.ClientStore(suite.ctx, clientexported.ClientTypeSoloMachine)

	bz := suite.aminoCdc.MustMarshalBinaryBare(suite.ClientState())
	suite.store.Set(host.KeyClientState(), bz)
}

func TestSoloMachineTestSuite(t *testing.T) {
	suite.Run(t, new(SoloMachineTestSuite))
}

func (suite *SoloMachineTestSuite) CreateHeader() solomachinetypes.Header {
	// generate new private key and signature for header
	newPrivKey := ed25519.GenPrivKey()
	data := append(sdk.Uint64ToBigEndian(suite.sequence), newPrivKey.PubKey().Bytes()...)
	signature, err := suite.privKey.Sign(data)
	suite.Require().NoError(err)

	header := solomachinetypes.Header{
		Sequence:  suite.sequence,
		Signature: signature,
		NewPubKey: newPrivKey.PubKey().Bytes(),
	}

	// assumes successful header update
	suite.sequence++
	suite.privKey = newPrivKey

	return header
}

func (suite *SoloMachineTestSuite) ClientState() solomachinetypes.ClientState {
	return solomachinetypes.NewClientState(suite.clientID, "", suite.ConsensusState())
}

func (suite *SoloMachineTestSuite) ConsensusState() solomachinetypes.ConsensusState {
	return solomachinetypes.ConsensusState{
		Sequence:  suite.sequence,
		PubKey:    suite.privKey.PubKey().Bytes(),
		Timestamp: suite.timestamp,
	}
}

func (suite *SoloMachineTestSuite) Evidence() solomachinetypes.Evidence {
	dataOne := []byte("DATA ONE")
	dataTwo := []byte("DATA TWO")

	sig, err := suite.privKey.Sign(append(sdk.Uint64ToBigEndian(suite.sequence), dataOne...))
	suite.Require().NoError(err)

	signatureOne := solomachinetypes.SignatureAndData{
		Signature: sig,
		Data:      dataOne,
	}

	sig, err = suite.privKey.Sign(append(sdk.Uint64ToBigEndian(suite.sequence), dataTwo...))
	suite.Require().NoError(err)

	signatureTwo := solomachinetypes.SignatureAndData{
		Signature: sig,
		Data:      dataTwo,
	}

	return solomachinetypes.Evidence{
		ClientID:     suite.clientID,
		Sequence:     suite.sequence,
		SignatureOne: &signatureOne,
		SignatureTwo: &signatureTwo,
	}
}

func (suite *SoloMachineTestSuite) GetSequenceFromStore() uint64 {
	bz := suite.store.Get(host.KeyClientState())
	suite.Require().NotNil(bz)

	var clientState solomachinetypes.ClientState
	suite.aminoCdc.MustUnmarshalBinaryBare(bz, &clientState)
	return clientState.ConsensusState.Sequence
}
