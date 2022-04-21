package cli

import (
	"testing"

	bankv1beta1 "github.com/cosmos/cosmos-sdk/api/cosmos/bank/v1beta1"
	"github.com/spf13/cobra"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gotest.tools/v3/assert"
)

func TestBank(t *testing.T) {
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(bankv1beta1.Query_ServiceDesc.ServiceName))
	assert.NilError(t, err)
	b := &Builder{}
	cmd := &cobra.Command{
		Use: "bank",
	}
	b.AddQueryService(cmd, desc.(protoreflect.ServiceDescriptor))
	cmd.SetArgs([]string{"balance", "--help"})
	cmd.Execute()
}
