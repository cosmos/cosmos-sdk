package msgservice

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
)

func TestValidateServiceAnnotations(t *testing.T) {
	// Find an arbitrary query service that hasn't the service=true annotation.
	sd, err := protoregistry.GlobalFiles.FindDescriptorByName("cosmos.bank.v1beta1.Query")
	require.NoError(t, err)
	err = validateMsgServiceAnnotations(nil, sd.(protoreflect.ServiceDescriptor))
	require.Error(t, err)

	sd, err = protoregistry.GlobalFiles.FindDescriptorByName("cosmos.bank.v1beta1.Msg")
	require.NoError(t, err)
	err = validateMsgServiceAnnotations(nil, sd.(protoreflect.ServiceDescriptor))
	require.NoError(t, err)
}

func TestValidateMsgAnnotations(t *testing.T) {
	testcases := []struct {
		name    string
		message proto.Message
		expErr  bool
	}{
		{"no signer annotation", &testpb.Dog{}, true},
		{"valid signer", &bankv1beta1.MsgSend{}, false},
		{"valid signer as message", &bankv1beta1.MsgMultiSend{}, false},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateMsgAnnotations(nil, tc.message.ProtoReflect().Descriptor())
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
