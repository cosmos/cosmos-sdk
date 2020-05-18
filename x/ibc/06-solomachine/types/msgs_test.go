package types_test

import (
	ibcsmtypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine"
)

const (
	clientID = "testclient"
)

func (suite *SoloMachineTestSuite) TestMsgCreateClientValidateBasic() {
	cases := []struct {
		name    string
		msg     ibcsmtypes.MsgCreateClient
		expPass bool
	}{
		{"valid msg", ibcsmtypes.NewMsgCreateClient(clientID, suite.ConsensusState()), true},
		{"invalid client id", ibcsmtypes.NewMsgCreateClient("(BADCLIENTID)", suite.ConsensusState()), false},
		{"invalid consensus state with zero sequence", ibcsmtypes.NewMsgCreateClient(clientID, ibcsmtypes.ConsensusState{0, suite.privKey.PubKey()}), false},
		{"invalid consensus state with nil pubkey", ibcsmtypes.NewMsgCreateClient(clientID, ibcsmtypes.ConsensusState{suite.sequence, nil}), false},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %v", i, tc.name)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.name)
		}
	}
}

func (suite *SoloMachineTestSuite) TestMsgUpdateClientValidateBasic() {
	header := suite.CreateHeader()

	cases := []struct {
		name    string
		msg     ibcsmtypes.MsgUpdateClient
		expPass bool
	}{
		{"valid msg", ibcsmtypes.NewMsgUpdateClient(clientID, header), true},
		{"invalid client id", ibcsmtypes.NewMsgUpdateClient("(BADCLIENTID)", header), false},
		{"invalid header - sequence is zero", ibcsmtypes.NewMsgUpdateClient(clientID, ibcsmtypes.Header{0, header.Signature, header.NewPubKey}), false},
		{"invalid header - signature is empty", ibcsmtypes.NewMsgUpdateClient(clientID, ibcsmtypes.Header{header.Sequence, []byte{}, header.NewPubKey}), false},
		{"invalid header - pubkey is empty", ibcsmtypes.NewMsgUpdateClient(clientID, ibcsmtypes.Header{header.Sequence, header.Signature, nil}), false},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %v", i, tc.name)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.name)
		}
	}

}

func (suite *SoloMachineTestSuite) TestMsgSubmitClientMisbehaviourValidateBasic() {
	cases := []struct {
		name    string
		msg     ibcsmtypes.MsgSubmitClientMisbehaviour
		expPass bool
	}{}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %v", i, tc.name)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.name)
		}
	}

}
