package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func TestMsgCreateClientValidateBasic(t *testing.T) {
	validator := tmtypes.NewValidator(tmtypes.NewMockPV().GetPubKey(), 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	now := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	cs := ibctmtypes.ConsensusState{
		Height:       height,
		Timestamp:    now,
		Root:         commitment.NewRoot([]byte("root")),
		ValidatorSet: valSet,
	}
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())

	cases := []struct {
		msg     ibctmtypes.MsgCreateClient
		expPass bool
		errMsg  string
	}{
		{ibctmtypes.NewMsgCreateClient(exported.ClientTypeTendermint, cs, trustingPeriod, ubdPeriod, signer), true, "success msg should pass"},
		{ibctmtypes.NewMsgCreateClient("BADCHAIN", cs, trustingPeriod, ubdPeriod, signer), false, "invalid client id passed"},
		{ibctmtypes.NewMsgCreateClient("goodchain", cs, trustingPeriod, ubdPeriod, signer), false, "unregistered client type passed"},
		{ibctmtypes.NewMsgCreateClient("goodchain", ibctmtypes.ConsensusState{}, trustingPeriod, ubdPeriod, signer), false, "invalid Consensus State in msg passed"},
		{ibctmtypes.NewMsgCreateClient("goodchain", cs, 0, ubdPeriod, signer), false, "zero trusting period passed"},
		{ibctmtypes.NewMsgCreateClient("goodchain", cs, trustingPeriod, 0, signer), false, "zero unbonding period passed"},
		{ibctmtypes.NewMsgCreateClient("goodchain", cs, trustingPeriod, ubdPeriod, nil), false, "Empty address passed"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.Error(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

func TestMsgUpdateClient(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	signer := sdk.AccAddress(privKey.PubKey().Address())

	cases := []struct {
		msg     ibctmtypes.MsgUpdateClient
		expPass bool
		errMsg  string
	}{
		{ibctmtypes.NewMsgUpdateClient(exported.ClientTypeTendermint, ibctmtypes.Header{}, signer), true, "success msg should pass"},
		{ibctmtypes.NewMsgUpdateClient("badClient", ibctmtypes.Header{}, signer), false, "invalid client id passed"},
		{ibctmtypes.NewMsgUpdateClient(exported.ClientTypeTendermint, ibctmtypes.Header{}, nil), false, "Empty address passed"},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "Msg %d failed: %v", i, err)
		} else {
			require.Error(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}
