package autocli

import (
	"bytes"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/internal/testpb"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gotest.tools/v3/assert"
	"net"
	"testing"
)

var testCmdMsgDesc = &autocliv1.ServiceCommandDescriptor{
	Service: testpb.Msg_ServiceDesc.ServiceName,
	RpcCommandOptions: []*autocliv1.RpcCommandOptions{
		{
			RpcMethod:  "Send",
			Use:        "send [pos1] [pos2] [pos3...]",
			Version:    "1.0",
			Alias:      []string{"s"},
			SuggestFor: []string{"send"},
			Example:    "send 1 abc {}",
			Short:      "send msg the value provided by the user",
			Long:       "send msg the value provided by the user as a proto JSON object with populated with the provided fields and positional arguments",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{
					ProtoField: "positional1",
				},
				{
					ProtoField: "positional2",
				},
			},
			FlagOptions: map[string]*autocliv1.FlagOptions{
				"u32": {
					Name:      "uint32",
					Shorthand: "u",
					Usage:     "some random uint32",
				},
				"i32": {
					Usage:        "some random int32",
					DefaultValue: "3",
				},
				"u64": {
					Usage:             "some random uint64",
					NoOptDefaultValue: "5",
				},
				"deprecated_field": {
					Deprecated: "don't use this",
				},
				"shorthand_deprecated_field": {
					Shorthand:  "s",
					Deprecated: "bad idea",
				},
				"hidden_bool": {
					Hidden: true,
				},
			},
		},
	},
	SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
		// we test the sub-command functionality using the same service with different options
		"deprecatedmsg": {
			Service: testpb.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:  "Send",
					Deprecated: "dont use this",
				},
			},
		},
	},
}

func testMsgExec(t *testing.T, args ...string) *testClientConn {
	server := grpc.NewServer()
	testpb.RegisterMsgServer(server, &testMessageServer{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NilError(t, err)
	go func() {
		err := server.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()
	defer server.GracefulStop()
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
	}
	b := &Builder{
		GetClientConn: func(*cobra.Command) (grpc.ClientConnInterface, error) {
			return conn, nil
		},
	}
	cmd, err := b.BuildModuleMsgCommand("test", testCmdMsgDesc)
	assert.NilError(t, err)
	cmd.SetArgs(args)
	cmd.SetOut(conn.out)
	assert.NilError(t, cmd.Execute())
	return conn
}

func TestMsgOptions(t *testing.T) {
	conn := testMsgExec(t,
		"send", "5", "6",
		"--uint32", "7",
		"--u64", // no opt default value 5

	)
	fmt.Println(conn.out.String())
	//lastReq := conn.lastRequest.(*testpb.MsgRequest)
	//assert.Equal(t, uint32(27), lastReq.U32) // shorthand got set
	//assert.Equal(t, int32(3), lastReq.I32)   // default value got set
	//assert.Equal(t, uint64(5), lastReq.U64)  // no opt default value got set
}

func TestBuildCustomMsgCommand(t *testing.T) {
	b := &Builder{}
	customCommandCalled := false
	cmd, err := b.BuildMsgCommand(map[string]*autocliv1.ModuleOptions{
		"test": {
			Tx: testCmdMsgDesc,
		},
	}, map[string]*cobra.Command{
		"test": {Use: "test", Run: func(cmd *cobra.Command, args []string) {
			customCommandCalled = true
		}},
	})
	assert.NilError(t, err)
	cmd.SetArgs([]string{"test", "tx"})
	assert.NilError(t, cmd.Execute())
	assert.Assert(t, customCommandCalled)
}

func TestBuilder_BuildMsgMethodCommand(t *testing.T) {

}

type testMessageServer struct {
	testpb.UnimplementedMsgServer
}
