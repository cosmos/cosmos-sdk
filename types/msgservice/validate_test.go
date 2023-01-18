package msgservice_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	_ "cosmossdk.io/api/cosmos/bank/v1beta1"
	_ "github.com/cosmos/cosmos-sdk/testutil/testdata_pulsar"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func TestValidateServiceAnnotations(t *testing.T) {
	// We didn't add the `msg.service = true` annotation on testdata.
	err := msgservice.ValidateServiceAnnotations(nil, "testdata.Msg")
	require.Error(t, err)

	err = msgservice.ValidateServiceAnnotations(nil, "cosmos.bank.v1beta1.Msg")
	require.NoError(t, err)
}

func TestValidateMsgAnnotations(t *testing.T) {
	// We didn't add any signer on MsgCreateDog.
	err := msgservice.ValidateMsgAnnotations(nil, "testdata.MsgCreateDog")
	require.Error(t, err)

	err = msgservice.ValidateMsgAnnotations(nil, "cosmos.bank.v1beta1.MsgSend")
	require.NoError(t, err)
}
