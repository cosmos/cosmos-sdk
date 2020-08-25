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
		{"valid msg", types.NewMsgCreateClient(suite.solomachine.ClientID, suite.solomachine.ConsensusState()), true},
		{"invalid client id", types.NewMsgCreateClient("(BADCLIENTID)", suite.solomachine.ConsensusState()), false},
		{"invalid consensus state with zero sequence", types.NewMsgCreateClient(suite.solomachine.ClientID, &types.ConsensusState{0, suite.solomachine.ConsensusState().PublicKey, suite.solomachine.Time}), false},
		{"invalid consensus state with zero timestamp", types.NewMsgCreateClient(suite.solomachine.ClientID, &types.ConsensusState{1, suite.solomachine.ConsensusState().PublicKey, 0}), false},
		{"invalid consensus state with nil pubkey", types.NewMsgCreateClient(suite.solomachine.ClientID, &types.ConsensusState{suite.solomachine.Sequence, nil, suite.solomachine.Time}), false},
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
	header := suite.solomachine.CreateHeader()

	cases := []struct {
		name    string
		msg     *types.MsgUpdateClient
		expPass bool
	}{
		{"valid msg", types.NewMsgUpdateClient(suite.solomachine.ClientID, header), true},
		{"invalid client id", types.NewMsgUpdateClient("(BADCLIENTID)", header), false},
		{"invalid header - sequence is zero", types.NewMsgUpdateClient(suite.solomachine.ClientID, &types.Header{0, header.Signature, header.NewPublicKey}), false},
		{"invalid header - signature is empty", types.NewMsgUpdateClient(suite.solomachine.ClientID, &types.Header{header.Sequence, []byte{}, header.NewPublicKey}), false},
		{"invalid header - pubkey is empty", types.NewMsgUpdateClient(suite.solomachine.ClientID, &types.Header{header.Sequence, header.Signature, nil}), false},
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
