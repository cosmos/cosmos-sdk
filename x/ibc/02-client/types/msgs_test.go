package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	cmn "github.com/tendermint/tendermint/libs/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	errs "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ evidenceexported.Evidence = mockEvidence{}
var _ evidenceexported.Evidence = mockBadEvidence{}

const mockStr = "mock"

// mock GoodEvidence
type mockEvidence struct{}

// Implement Evidence interface
func (me mockEvidence) Route() string        { return mockStr }
func (me mockEvidence) Type() string         { return mockStr }
func (me mockEvidence) String() string       { return mockStr }
func (me mockEvidence) Hash() cmn.HexBytes   { return cmn.HexBytes([]byte(mockStr)) }
func (me mockEvidence) ValidateBasic() error { return nil }
func (me mockEvidence) GetHeight() int64     { return 3 }

// mock bad evidence
type mockBadEvidence struct {
	mockEvidence
}

// Override ValidateBasic
func (mbe mockBadEvidence) ValidateBasic() error {
	return errs.ErrInvalidEvidence
}

type ClientTestSuite struct {
	suite.Suite

	clientID           string
	signer             sdk.AccAddress
	cs                 tendermint.ConsensusState
	createClient       MsgCreateClient
	updateClient       MsgUpdateClient
	submitMisbehaviour MsgSubmitMisbehaviour
}

func (suite *ClientTestSuite) SetupSuite() {
	privateKey := secp256k1.GenPrivKey()
	suite.signer = sdk.AccAddress(privateKey.PubKey().Address())
	suite.cs = tendermint.ConsensusState{}
	suite.clientID = exported.ClientTypeTendermint
}

func (suite *ClientTestSuite) SetupTest() {
	suite.createClient = NewMsgCreateClient(suite.clientID, exported.ClientTypeTendermint, suite.cs, suite.signer)
	suite.updateClient = NewMsgUpdateClient(suite.clientID, tendermint.Header{}, suite.signer)
	suite.submitMisbehaviour = NewMsgSubmitMisbehaviour(suite.clientID, mockEvidence{}, suite.signer)
}

/*
	MsgCreateClient tests
*/

func (suite *ClientTestSuite) TestCreate_ValidateBasicValidMsg() {
	err := suite.createClient.ValidateBasic()
	require.Nil(suite.T(), err, "Msg failed: %v", err)
}

func (suite *ClientTestSuite) TestCreate_ValidateBasicInvalidClientID() {
	suite.createClient.ClientID = "badClient"
	err := suite.createClient.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "invalid client id")
}

func (suite *ClientTestSuite) TestCreate_ValidateBasicInvalidClientType() {
	suite.createClient.ClientType = "bad_type"
	err := suite.createClient.ValidateBasic()
	//fmt.Println(err)
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "unregistered client type")
	//fmt.Println(errors.Is(err, errs.ErrInvalidClientType))
}

func (suite *ClientTestSuite) TestCreate_ValidateBasicNilConsensusState() {
	suite.createClient.ConsensusState = nil
	err := suite.createClient.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "Nil Consensus State in msg")
	require.Truef(suite.T(), errors.Is(err, errs.ErrInvalidConsensus), "Nil Consensus State in msg with bad type of error: %v", err)
}

func (suite *ClientTestSuite) TestCreate_ValidateBasicEmptyAddress() {
	suite.createClient.Signer = sdk.AccAddress{}
	err := suite.createClient.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "Empty address")
	require.Truef(suite.T(), errors.Is(err, sdkerrors.ErrInvalidAddress), "Empty address with bad type of error: %v", err)
}

func (suite *ClientTestSuite) TestCreate_Route() {
	require.Equal(suite.T(), suite.createClient.Route(), ibctypes.RouterKey)
}

func (suite *ClientTestSuite) TestCreate_Type() {
	require.Equal(suite.T(), suite.createClient.Type(), TypeMsgCreateClient)
}

func (suite *ClientTestSuite) TestCreate_GetSigners() {
	require.Greater(suite.T(), len(suite.createClient.GetSigners()), 0)
	if len(suite.createClient.GetSigners()) > 0 {
		require.Equal(suite.T(), suite.createClient.GetSigners()[0].Bytes(), suite.signer.Bytes())
	}
}

/*
	MsgUpdateClient tests
*/

func (suite *ClientTestSuite) TestUpdate_ValidateBasicValidMsg() {
	err := suite.updateClient.ValidateBasic()
	require.Nil(suite.T(), err, "Msg failed: %v", err)
}

func (suite *ClientTestSuite) TestUpdate_ValidateBasicInvalidClientID() {
	suite.updateClient.ClientID = "badClient"
	err := suite.updateClient.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "invalid client id")
}

func (suite *ClientTestSuite) TestUpdate_ValidateBasicNilHeader() {
	suite.updateClient.Header = nil
	err := suite.updateClient.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "Nil Header")
	require.Truef(suite.T(), errors.Is(err, errs.ErrInvalidHeader), "Nil Header with bad type of error: %v", err)
}

func (suite *ClientTestSuite) TestUpdate_ValidateBasicEmptyAddress() {
	suite.updateClient.Signer = sdk.AccAddress{}
	err := suite.updateClient.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "Empty address")
	require.Truef(suite.T(), errors.Is(err, sdkerrors.ErrInvalidAddress), "Empty address with bad type of error: %v", err)
}

func (suite *ClientTestSuite) TestUpdate_Route() {
	require.Equal(suite.T(), suite.updateClient.Route(), ibctypes.RouterKey)
}

func (suite *ClientTestSuite) TestUpdate_Type() {
	require.Equal(suite.T(), suite.updateClient.Type(), TypeMsgUpdateClient)
}

func (suite *ClientTestSuite) TestUpdate_GetSigners() {
	require.Greater(suite.T(), len(suite.updateClient.GetSigners()), 0)
	if len(suite.updateClient.GetSigners()) > 0 {
		require.Equal(suite.T(), suite.updateClient.GetSigners()[0].Bytes(), suite.signer.Bytes())
	}
}

/*
	MsgSubmitMisbehaviour tests
*/

func (suite *ClientTestSuite) TestSubmitMisbehaviour_ValidateBasicValidMsg() {
	err := suite.submitMisbehaviour.ValidateBasic()
	require.Nil(suite.T(), err, "Msg failed: %v", err)
}

func (suite *ClientTestSuite) TestSubmitMisbehaviour_ValidateBasicInvalidClientID() {
	suite.submitMisbehaviour.ClientID = "badClient"
	err := suite.submitMisbehaviour.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "invalid client id")
}

func (suite *ClientTestSuite) TestSubmitMisbehaviour_ValidateBasicNilEvidence() {
	suite.submitMisbehaviour.Evidence = nil
	err := suite.submitMisbehaviour.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "Nil Evidence")
}

func (suite *ClientTestSuite) TestSubmitMisbehaviour_ValidateBasicInvalidEvidence() {
	suite.submitMisbehaviour.Evidence = mockBadEvidence{}
	err := suite.submitMisbehaviour.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "Invalid Evidence")
}

func (suite *ClientTestSuite) TestSubmitMisbehaviour_ValidateBasicEmptyAddress() {
	suite.submitMisbehaviour.Signer = sdk.AccAddress{}
	err := suite.submitMisbehaviour.ValidateBasic()
	require.NotNil(suite.T(), err, "Invalid Msg passed: %v", "Empty address")
	require.Truef(suite.T(), errors.Is(err, sdkerrors.ErrInvalidAddress), "Empty address with bad type of error: %v", err)
}

func (suite *ClientTestSuite) TestSubmitMisbehaviour_Route() {
	require.Equal(suite.T(), suite.submitMisbehaviour.Route(), ibctypes.RouterKey)
}

func (suite *ClientTestSuite) TestSubmitMisbehaviour_Type() {
	require.Equal(suite.T(), suite.submitMisbehaviour.Type(), TypeClientMisbehaviour)
}

func (suite *ClientTestSuite) TestSubmitMisbehaviour_GetSigners() {
	require.Greater(suite.T(), len(suite.submitMisbehaviour.GetSigners()), 0)
	if len(suite.submitMisbehaviour.GetSigners()) > 0 {
		require.Equal(suite.T(), suite.submitMisbehaviour.GetSigners()[0].Bytes(), suite.signer.Bytes())
	}
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
