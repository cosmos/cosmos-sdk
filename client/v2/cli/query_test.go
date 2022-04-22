package cli

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/client/v2/internal/testpb"
)

func testExec(t *testing.T, args ...string) {
	b := &Builder{
		GetClientConn: func(ctx context.Context) grpc.ClientConnInterface {
			return testClientConn{t: t}
		},
	}
	cmd := b.AddQueryServiceCommands(&cobra.Command{Use: "test"}, protoreflect.FullName(testpb.Query_ServiceDesc.ServiceName))
	cmd.SetArgs(args)
	assert.NilError(t, cmd.Execute())
}

func TestFoo(t *testing.T) {
	testExec(t,
		"foo",
		"--a-bool",
		"--an-enum", "one",
		"--a-message", `{"bar":"abc", "baz":-3}`,
		"--duration", "4h3s",
		"--u-32", "27",
		"--u-64", "3267246890",
		"--i-32", "-253",
		"--i-64", "-234602347",
		"--str", "def",
		"--timestamp", "2019-01-02T00:01:02Z",
	)
}

func TestHelp(t *testing.T) {
	testExec(t, "foo", "-h")
}

type testClientConn struct {
	t *testing.T
}

func (t testClientConn) Invoke(_ context.Context, method string, args interface{}, _ interface{}, _ ...grpc.CallOption) error {
	in, err := protojson.Marshal(args.(proto.Message))
	if err != nil {
		return err
	}
	t.t.Logf("invoke %s: %s", method, in)
	return nil
}

func (t testClientConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	t.t.Fatal("unexpected streaming call")
	return nil, nil
}
