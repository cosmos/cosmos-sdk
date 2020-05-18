package types_test

import (
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine"
)

func (suite *SoloMachineTestSuite) TestMsgCreateClientValidateBasic() {
	cases := []struct {
		name    string
		msg     solomachinetypes.MsgCreateClient
		expPass bool
	}{
		{"valid msg", solomachinetypes.NewMsgCreateClient(suite.clientID, suite.ConsensusState()), true},
		{"invalid client id", solomachinetypes.NewMsgCreateClient("(BADCLIENTID)", suite.ConsensusState()), false},
		{"invalid consensus state with zero sequence", solomachinetypes.NewMsgCreateClient(suite.clientID, solomachinetypes.ConsensusState{0, suite.privKey.PubKey()}), false},
		{"invalid consensus state with nil pubkey", solomachinetypes.NewMsgCreateClient(suite.clientID, solomachinetypes.ConsensusState{suite.sequence, nil}), false},
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
		msg     solomachinetypes.MsgUpdateClient
		expPass bool
	}{
		{"valid msg", solomachinetypes.NewMsgUpdateClient(suite.clientID, header), true},
		{"invalid client id", solomachinetypes.NewMsgUpdateClient("(BADCLIENTID)", header), false},
		{"invalid header - sequence is zero", solomachinetypes.NewMsgUpdateClient(suite.clientID, solomachinetypes.Header{0, header.Signature, header.NewPubKey}), false},
		{"invalid header - signature is empty", solomachinetypes.NewMsgUpdateClient(suite.clientID, solomachinetypes.Header{header.Sequence, []byte{}, header.NewPubKey}), false},
		{"invalid header - pubkey is empty", solomachinetypes.NewMsgUpdateClient(suite.clientID, solomachinetypes.Header{header.Sequence, header.Signature, nil}), false},
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

// TODO
func (suite *SoloMachineTestSuite) TestMsgSubmitClientMisbehaviourValidateBasic() {
	cases := []struct {
		name    string
		msg     solomachinetypes.MsgSubmitClientMisbehaviour
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
