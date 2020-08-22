package types_test

import (
	ics23 "github.com/confio/ics23/go"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	types "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *TendermintTestSuite) TestMsgCreateClientValidateBasic() {
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())
	invalidHeader := types.CreateTestHeader(suite.header.Header.GetChainID(), height, 0, suite.now, suite.valSet, nil, []tmtypes.PrivValidator{suite.privVal})
	invalidHeader.ValidatorSet = nil

	cases := []struct {
		msg     *types.MsgCreateClient
		expPass bool
		errMsg  string
	}{
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, signer), true, "success msg should pass"},
		{types.NewMsgCreateClient("(BADCHAIN)", suite.header, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, signer), false, "invalid client id passed"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, types.Fraction{Numerator: 0, Denominator: 1}, trustingPeriod, ubdPeriod, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, signer), false, "invalid trust level"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, types.DefaultTrustLevel, 0, ubdPeriod, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, signer), false, "zero trusting period passed"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, types.DefaultTrustLevel, trustingPeriod, 0, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, signer), false, "zero unbonding period passed"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, nil), false, "Empty address passed"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, types.Header{}, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, signer), false, "nil header"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, invalidHeader, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, signer), false, "invalid header"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, []*ics23.ProofSpec{nil}, false, false, signer), false, "invalid proof specs"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, nil, false, false, signer), false, "nil proof specs"},
		{types.NewMsgCreateClient(exported.ClientTypeTendermint, suite.header, types.DefaultTrustLevel, ubdPeriod, ubdPeriod, maxClockDrift, commitmenttypes.GetSDKSpecs(), false, false, signer), false, "trusting period not less than unbonding period"},
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
		msg     *types.MsgUpdateClient
		expPass bool
		errMsg  string
	}{
		{types.NewMsgUpdateClient(exported.ClientTypeTendermint, types.Header{}, signer), true, "success msg should pass"},
		{types.NewMsgUpdateClient("(badClient)", types.Header{}, signer), false, "invalid client id passed"},
		{types.NewMsgUpdateClient(exported.ClientTypeTendermint, types.Header{}, nil), false, "Empty address passed"},
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
