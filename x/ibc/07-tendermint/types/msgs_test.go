package types_test

import (
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

func (suite *TendermintTestSuite) TestMsgCreateClientValidateBasic() {
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())

	cases := []struct {
		msg     ibctmtypes.MsgCreateClient
		expPass bool
		errMsg  string
	}{
		{ibctmtypes.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, trustingPeriod, ubdPeriod, maxClockDrift, signer), true, "success msg should pass"},
		{ibctmtypes.NewMsgCreateClient("(BADCHAIN)", suite.header, trustingPeriod, ubdPeriod, maxClockDrift, signer), false, "invalid client id passed"},
		{ibctmtypes.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, 0, ubdPeriod, maxClockDrift, signer), false, "zero trusting period passed"},
		{ibctmtypes.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, trustingPeriod, 0, maxClockDrift, signer), false, "zero unbonding period passed"},
		{ibctmtypes.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, trustingPeriod, ubdPeriod, maxClockDrift, nil), false, "Empty address passed"},
		{ibctmtypes.NewMsgCreateClient(exported.ClientTypeTendermint, ibctmtypes.Header{}, trustingPeriod, ubdPeriod, maxClockDrift, signer), false, "nil header"},
		{ibctmtypes.NewMsgCreateClient(exported.ClientTypeTendermint, ibctmtypes.Header{}, trustingPeriod, ubdPeriod, maxClockDrift, signer), false, "invalid header"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %v", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

func (suite *TendermintTestSuite) TestMsgUpdateClient() {
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())

	cases := []struct {
		msg     ibctmtypes.MsgUpdateClient
		expPass bool
		errMsg  string
	}{
		{ibctmtypes.NewMsgUpdateClient(exported.ClientTypeTendermint, ibctmtypes.Header{}, signer), true, "success msg should pass"},
		{ibctmtypes.NewMsgUpdateClient("(badClient)", ibctmtypes.Header{}, signer), false, "invalid client id passed"},
		{ibctmtypes.NewMsgUpdateClient(exported.ClientTypeTendermint, ibctmtypes.Header{}, nil), false, "Empty address passed"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %v", i, err)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}
