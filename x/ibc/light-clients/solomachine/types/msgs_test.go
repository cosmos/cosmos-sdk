package types_test

import (
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
)

func (suite *SoloMachineTestSuite) TestMsgCreateClientValidateBasic() {
	cases := []struct {
		name    string
		msg     *solomachinetypes.MsgCreateClient
		expPass bool
	}{
		{"valid msg", solomachinetypes.NewMsgCreateClient(suite.clientID, suite.ConsensusState()), true},
		{"invalid client id", solomachinetypes.NewMsgCreateClient("(BADCLIENTID)", suite.ConsensusState()), false},
		{"invalid consensus state with zero sequence", solomachinetypes.NewMsgCreateClient(suite.clientID, &solomachinetypes.ConsensusState{0, suite.pubKey, timestamp}), false},
		{"invalid consensus state with zero timestamp", solomachinetypes.NewMsgCreateClient(suite.clientID, &solomachinetypes.ConsensusState{1, suite.pubKey, 0}), false},
		{"invalid consensus state with nil pubkey", solomachinetypes.NewMsgCreateClient(suite.clientID, &solomachinetypes.ConsensusState{suite.sequence, nil, timestamp}), false},
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
		msg     *solomachinetypes.MsgUpdateClient
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
