package cli

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gotest.tools/v3/assert"

	bankv1beta1 "github.com/cosmos/cosmos-sdk/api/cosmos/bank/v1beta1"
)

func TestBank(t *testing.T) {
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(bankv1beta1.Query_ServiceDesc.ServiceName))
	assert.NilError(t, err)
	b := &Builder{
		GetClientConn: func(ctx context.Context) grpc.ClientConnInterface {
			return testClientConn{t: t}
		},
		JSONMarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
	}
	cmd := &cobra.Command{
		Use: "bank",
	}
	b.AddQueryService(cmd, desc.(protoreflect.ServiceDescriptor))
	cmd.SetArgs([]string{"help", "all-balances"})
	cmd.Execute()
}

type testClientConn struct {
	t *testing.T
}

func (t testClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	in, err := protojson.Marshal(args.(proto.Message))
	if err != nil {
		return err
	}
	t.t.Logf("invoke %s: %s", method, in)
	return nil
}

func (t testClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	t.t.Fatal("unexpected streaming call")
	return nil, nil
}
