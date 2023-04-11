package autocli

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"gotest.tools/v3/golden"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"

	"cosmossdk.io/client/v2/internal/testpb"
)

var buildModuleMsgCommand = func(moduleName string, b *Builder) (*cobra.Command, error) {
	cmd := topLevelCmd(moduleName, fmt.Sprintf("Transactions commands for the %s module", moduleName))

	err := b.AddMsgServiceCommands(cmd, testCmdMsgDesc)
	return cmd, err
}

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
				{
					ProtoField: "positional3_varargs",
					Varargs:    true,
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
					Usage:        "some random uint64",
					DefaultValue: "5",
				},
				"deprecated_field": {
					Deprecated: "don't use this",
				},
				"shorthand_deprecated_field": {
					Shorthand:  "d",
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
					Short:      "deprecated subcommand",
				},
			},
		},
		"skipmsg": {
			Service: testpb.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Send",
					Skip:      true,
					Short:     "skip subcommand",
				},
			},
		},
	},
}

func TestMsgOptions(t *testing.T) {
	conn := testExecCommon(t,
		buildModuleMsgCommand,
		"send", "5", "6", "1foo",
		"--uint32", "7",
		"--u64", "8",
		"--output", "json",
	)
	response := conn.out.String()
	var output testpb.MsgRequest
	err := protojson.Unmarshal([]byte(response), &output)
	assert.NilError(t, err)
	assert.Equal(t, output.GetU32(), uint32(7))
	assert.Equal(t, output.GetPositional1(), int32(5))
	assert.Equal(t, output.GetPositional2(), "6")
}

func TestMsgOutputFormat(t *testing.T) {
	conn := testExecCommon(t, buildModuleMsgCommand,
		"send", "5", "6", "1foo",
		"--output", "json",
	)
	assert.Assert(t, strings.Contains(conn.out.String(), "{"))
	conn = testExecCommon(t, buildModuleMsgCommand,
		"send", "5", "6", "1foo",
		"--output", "text",
	)

	assert.Assert(t, strings.Contains(conn.out.String(), "positional1: 5"))
}

func TestMsgOptionsError(t *testing.T) {
	conn := testExecCommon(t, buildModuleMsgCommand,
		"send", "5",
		"--uint32", "7",
		"--u64", "8",
	)

	assert.Assert(t, strings.Contains(conn.errorOut.String(), "requires at least 3 arg"))

	conn = testExecCommon(t, buildModuleMsgCommand,
		"send", "5", "6", `{"denom":"foo","amount":"1"}`,
		"--uint32", "7",
		"--u64", "abc",
	)
	assert.Assert(t, strings.Contains(conn.errorOut.String(), "invalid argument "))
}

func TestDeprecatedMsg(t *testing.T) {
	conn := testExecCommon(t, buildModuleMsgCommand, "send",
		"1", "abc", `{"denom":"foo","amount":"1"}`,
		"--deprecated-field", "foo")
	assert.Assert(t, strings.Contains(conn.out.String(), "--deprecated-field has been deprecated"))

	conn = testExecCommon(t, buildModuleMsgCommand, "send",
		"1", "abc", `{"denom":"foo","amount":"1"}`,
		"-d", "foo")
	assert.Assert(t, strings.Contains(conn.out.String(), "--shorthand-deprecated-field has been deprecated"))
}

func TestEverythingMsg(t *testing.T) {
	conn := testExecCommon(t, buildModuleMsgCommand,
		"send",
		"1",
		"abc",
		"1234foo",
		"4321foo",
		"--output", "json",
		"--a-bool",
		"--an-enum", "two",
		"--a-message", `{"bar":"abc", "baz":-3}`,
		"--duration", "4h3s",
		"--uint32", "27",
		"--u64", "3267246890",
		"--i32", "-253",
		"--i64", "-234602347",
		"--str", "def",
		"--timestamp", "2019-01-02T00:01:02Z",
		"--a-coin", "10000000foo",
		"--an-address", "cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk",
		"--bz", "c2RncXdlZndkZ3NkZw==",
		"--page-count-total",
		"--page-key", "MTIzNTQ4N3NnaGRhcw==",
		"--page-limit", "1000",
		"--page-offset", "10",
		"--page-reverse",
		"--bools", "true",
		"--bools", "false,false,true",
		"--enums", "one",
		"--enums", "five",
		"--enums", "two",
		"--strings", "abc",
		"--strings", "xyz",
		"--strings", "xyz,qrs",
		"--durations", "3s",
		"--durations", "5s",
		"--durations", "10h",
		"--some-messages", "{}",
		"--some-messages", `{"bar":"baz"}`,
		"--some-messages", `{"baz":-1}`,
		"--uints", "1,2,3",
		"--uints", "4",
	)
	response := conn.out.String()
	var output testpb.MsgRequest
	err := protojson.Unmarshal([]byte(response), &output)
	assert.NilError(t, err)
	assert.Equal(t, output.GetU32(), uint32(27))
	assert.Equal(t, output.GetU64(), uint64(3267246890))
	assert.Equal(t, output.GetPositional1(), int32(1))
	assert.Equal(t, output.GetPositional2(), "abc")
	assert.Equal(t, output.GetABool(), true)
	assert.Equal(t, output.GetAnEnum(), testpb.Enum_ENUM_TWO)
}

func TestHelpMsg(t *testing.T) {
	conn := testExecCommon(t, buildModuleMsgCommand, "-h")
	golden.Assert(t, conn.out.String(), "help-toplevel-msg.golden")

	conn = testExecCommon(t, buildModuleMsgCommand, "send", "-h")
	golden.Assert(t, conn.out.String(), "help-echo-msg.golden")

	conn = testExecCommon(t, buildModuleMsgCommand, "deprecatedmsg", "send", "-h")
	golden.Assert(t, conn.out.String(), "help-deprecated-msg.golden")
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

func TestErrorBuildMsgCommand(t *testing.T) {
	b := &Builder{}

	commandDescriptor := &autocliv1.ServiceCommandDescriptor{
		Service: testpb.Msg_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod: "Send",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{
						ProtoField: "un-existent-proto-field",
					},
				},
			},
		},
	}

	opts := map[string]*autocliv1.ModuleOptions{
		"test": {
			Tx: commandDescriptor,
		},
	}
	_, err := b.BuildMsgCommand(opts, nil)
	assert.ErrorContains(t, err, "can't find field un-existent-proto-field")

	nonExistentService := &autocliv1.ServiceCommandDescriptor{Service: "un-existent-service"}
	opts = map[string]*autocliv1.ModuleOptions{
		"test": {
			Tx: nonExistentService,
		},
	}
	_, err = b.BuildMsgCommand(opts, nil)
	assert.ErrorContains(t, err, "can't find service un-existent-service")
}

func TestNotFoundErrorsMsg(t *testing.T) {
	b := &Builder{}
	buildModuleMsgCommand := func(moduleName string, cmdDescriptor *autocliv1.ServiceCommandDescriptor) (*cobra.Command, error) {
		cmd := topLevelCmd(moduleName, fmt.Sprintf("Transactions commands for the %s module", moduleName))

		err := b.AddMsgServiceCommands(cmd, cmdDescriptor)
		return cmd, err
	}

	// Query non existent service
	_, err := buildModuleMsgCommand("test", &autocliv1.ServiceCommandDescriptor{Service: "un-existent-service"})
	assert.ErrorContains(t, err, "can't find service un-existent-service")

	_, err = buildModuleMsgCommand("test", &autocliv1.ServiceCommandDescriptor{
		Service:           testpb.Query_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{{RpcMethod: "un-existent-method"}},
	})
	assert.ErrorContains(t, err, "rpc method \"un-existent-method\" not found")

	_, err = buildModuleMsgCommand("test", &autocliv1.ServiceCommandDescriptor{
		Service: testpb.Msg_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod: "Send",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{
						ProtoField: "un-existent-proto-field",
					},
				},
			},
		},
	})
	assert.ErrorContains(t, err, "can't find field un-existent-proto-field")

	_, err = buildModuleMsgCommand("test", &autocliv1.ServiceCommandDescriptor{
		Service: testpb.Msg_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod: "Send",
				FlagOptions: map[string]*autocliv1.FlagOptions{
					"un-existent-flag": {},
				},
			},
		},
	})
	assert.ErrorContains(t, err, "can't find field un-existent-flag")
}

func TestEnhanceMessageCommand(t *testing.T) {
	b := &Builder{}
	enhanceMsg := func(cmd *cobra.Command, modOpts *autocliv1.ModuleOptions, moduleName string) error {
		txCmdDesc := modOpts.Tx
		if txCmdDesc != nil {
			subCmd := topLevelCmd(moduleName, fmt.Sprintf("Transactions commands for the %s module", moduleName))
			err := b.AddMsgServiceCommands(cmd, txCmdDesc)
			if err != nil {
				return err
			}

			cmd.AddCommand(subCmd)
		}
		return nil
	}

	// Test that the command has a subcommand
	cmd := &cobra.Command{Use: "test"}
	cmd.AddCommand(&cobra.Command{Use: "test"})
	options := map[string]*autocliv1.ModuleOptions{
		"test": {},
	}
	err := b.enhanceCommandCommon(cmd, options, map[string]*cobra.Command{}, enhanceMsg)
	assert.NilError(t, err)

	cmd = &cobra.Command{Use: "test"}
	options = map[string]*autocliv1.ModuleOptions{}
	customCommands := map[string]*cobra.Command{
		"test2": {Use: "test"},
	}
	err = b.enhanceCommandCommon(cmd, options, customCommands, enhanceMsg)
	assert.NilError(t, err)

	cmd = &cobra.Command{Use: "test"}
	options = map[string]*autocliv1.ModuleOptions{
		"test": {Tx: nil},
	}
	customCommands = map[string]*cobra.Command{}
	err = b.enhanceCommandCommon(cmd, options, customCommands, enhanceMsg)
	assert.NilError(t, err)
}
