package msgservice_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
)

func TestValidateServiceAnnotations(t *testing.T) {
	// Find an arbitrary query service that hasn't the service=true annotation.
	sd, err := protoregistry.GlobalFiles.FindDescriptorByName("cosmos.bank.v1beta1.Query")
	require.NoError(t, err)
	err = msgservice.ValidateMsgServiceAnnotations(sd.(protoreflect.ServiceDescriptor))
	require.Error(t, err)

	sd, err = protoregistry.GlobalFiles.FindDescriptorByName("cosmos.bank.v1beta1.Msg")
	require.NoError(t, err)
	err = msgservice.ValidateMsgServiceAnnotations(sd.(protoreflect.ServiceDescriptor))
	require.NoError(t, err)
}
