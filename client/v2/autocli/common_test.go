package autocli

import (
	"bytes"
	"cosmossdk.io/client/v2/internal/testpb"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"gotest.tools/v3/assert"

	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

func testExecCommon(t *testing.T, buildModuleCommand func(string, *Builder) (*cobra.Command, error), args ...string) *testClientConn {
	server := grpc.NewServer()
	testpb.RegisterQueryServer(server, &testEchoServer{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NilError(t, err)
	go func() {
		err := server.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	clientConn, err := grpc.Dial(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	defer func() {
		err := clientConn.Close()
		if err != nil {
			panic(err)
		}
	}()

	conn := &testClientConn{
		ClientConn: clientConn,
		t:          t,
		out:        &bytes.Buffer{},
		errorOut:   &bytes.Buffer{},
	}
	b := &Builder{
		GetClientConn: func(*cobra.Command) (grpc.ClientConnInterface, error) {
			return conn, nil
		},
		AddQueryConnFlags: flags.AddQueryFlagsToCmd,
		AddTxConnFlags:    flags.AddTxFlagsToCmd,
	}

	cmd, err := buildModuleCommand("test", b)
	assert.NilError(t, err)
	assert.NilError(t, err)
	cmd.SetArgs(args)
	cmd.SetOut(conn.out)
	cmd.SetErr(conn.errorOut)
	cmd.Execute()
	return conn
}

func TestBinaryFile(t *testing.T) {
	conn := testExecCommon(t, buildModuleQueryCommand,
		"echo",
		"1", "abc", `{"denom":"foo","amount":"1"}`,
		"-u", "27", // shorthand
		"--u64", "5", // no opt default value
	)
	lastReq := conn.lastRequest.(*testpb.EchoRequest)
	assert.Equal(t, uint32(27), lastReq.U32) // shorthand got set
	assert.Equal(t, int32(3), lastReq.I32)   // default value got set
	assert.Equal(t, uint64(5), lastReq.U64)  // no opt default value got set
}
