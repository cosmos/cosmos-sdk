package msgservice_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	_ "cosmossdk.io/api/cosmos/bank/v1beta1"
	_ "github.com/cosmos/cosmos-sdk/testutil/testdata_pulsar"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func TestValidateServiceAnnotations(t *testing.T) {
	// We didn't add the `msg.service = true` annotation on testdata's Msg.
	err := msgservice.ValidateServiceAnnotations(nil, "testdata.Msg")
	require.Error(t, err)

	err = msgservice.ValidateServiceAnnotations(nil, "cosmos.bank.v1beta1.Msg")
	require.NoError(t, err)
}

func TestValidateMsgAnnotations(t *testing.T) {
	testcases := []struct {
		name    string
		typeURL string
		expErr  bool
	}{
		{"no signer annotation", "testdata.MsgCreateDog", true},
		{"valid signer", "cosmos.bank.v1beta1.MsgSend", false},
		{"valid signer as message", "cosmos.bank.v1beta1.MsgMultiSend", false},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := msgservice.ValidateMsgAnnotations(nil, tc.typeURL)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
