package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestMsgCreateClientValidateBasic() {
	cases := []struct {
		name    string
		msg     *types.MsgCreateClient
		expPass bool
	}{
		{"valid msg", types.NewMsgCreateClient(suite.clientID, suite.ConsensusState()), true},
		{"invalid client id", types.NewMsgCreateClient("(BADCLIENTID)", suite.ConsensusState()), false},
		{"invalid consensus state with zero sequence", types.NewMsgCreateClient(suite.clientID, &types.ConsensusState{0, suite.pubKey, timestamp}), false},
		{"invalid consensus state with zero timestamp", types.NewMsgCreateClient(suite.clientID, &types.ConsensusState{1, suite.pubKey, 0}), false},
		{"invalid consensus state with nil pubkey", types.NewMsgCreateClient(suite.clientID, &types.ConsensusState{suite.sequence, nil, timestamp}), false},
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
		msg     *types.MsgUpdateClient
		expPass bool
	}{
		{"valid msg", types.NewMsgUpdateClient(suite.clientID, header), true},
		{"invalid client id", types.NewMsgUpdateClient("(BADCLIENTID)", header), false},
		{"invalid header - sequence is zero", types.NewMsgUpdateClient(suite.clientID, types.Header{0, header.Signature, header.NewPubKey}), false},
		{"invalid header - signature is empty", types.NewMsgUpdateClient(suite.clientID, types.Header{header.Sequence, []byte{}, header.NewPubKey}), false},
		{"invalid header - pubkey is empty", types.NewMsgUpdateClient(suite.clientID, types.Header{header.Sequence, header.Signature, nil}), false},
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
