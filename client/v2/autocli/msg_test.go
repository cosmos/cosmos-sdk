package autocli

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/client/v2/internal/testpb"
)

var buildModuleMsgCommand = func(moduleName string, b *Builder) (*cobra.Command, error) {
	cmd := topLevelCmd(moduleName, fmt.Sprintf("Transactions commands for the %s module", moduleName))
	err := b.AddMsgServiceCommands(cmd, bankAutoCLI)
	return cmd, err
}

func buildCustomModuleMsgCommand(cmdDescriptor *autocliv1.ServiceCommandDescriptor) func(moduleName string, b *Builder) (*cobra.Command, error) {
	return func(moduleName string, b *Builder) (*cobra.Command, error) {
		cmd := topLevelCmd(moduleName, fmt.Sprintf("Transactions commands for the %s module", moduleName))
		err := b.AddMsgServiceCommands(cmd, cmdDescriptor)
		return cmd, err
	}
}

var bankAutoCLI = &autocliv1.ServiceCommandDescriptor{
	Service: bankv1beta1.Msg_ServiceDesc.ServiceName,
	RpcCommandOptions: []*autocliv1.RpcCommandOptions{
		{
			RpcMethod:      "Send",
			Use:            "send [from_key_or_address] [to_address] [amount] [flags]",
			Short:          "Send coins from one account to another",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "from_address"}, {ProtoField: "to_address"}, {ProtoField: "amount"}},
		},
	},
	EnhanceCustomCommand: true,
}

func TestMsg(t *testing.T) {
	fixture := initFixture(t)
	out, err := runCmd(fixture.conn, fixture.b, buildModuleMsgCommand, "send",
		"cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk", "cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk", "1foo",
		"--generate-only",
		"--output", "json",
	)
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "msg-output.golden")

	out, err = runCmd(fixture.conn, fixture.b, buildCustomModuleMsgCommand(&autocliv1.ServiceCommandDescriptor{
		Service: bankv1beta1.Msg_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod:      "Send",
				Use:            "send [from_key_or_address] [to_address] [amount] [flags]",
				Short:          "Send coins from one account to another",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "to_address"}, {ProtoField: "amount"}},
				// from_address should be automatically added
			},
		},
		EnhanceCustomCommand: true,
	}), "send",
		"cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk", "1foo",
		"--from", "cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk",
		"--generate-only",
		"--output", "json",
	)
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "msg-output.golden")

	out, err = runCmd(fixture.conn, fixture.b, buildCustomModuleMsgCommand(&autocliv1.ServiceCommandDescriptor{
		Service: bankv1beta1.Msg_ServiceDesc.ServiceName,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod:      "Send",
				Use:            "send [from_key_or_address] [to_address] [amount] [flags]",
				Short:          "Send coins from one account to another",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "to_address"}, {ProtoField: "amount"}},
				FlagOptions: map[string]*autocliv1.FlagOptions{
					"from_address": {Name: "sender"}, // use a custom flag for signer
				},
			},
		},
		EnhanceCustomCommand: true,
	}), "send",
		"cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk", "1foo",
		"--sender", "cosmos1y74p8wyy4enfhfn342njve6cjmj5c8dtl6emdk",
		"--generate-only",
		"--output", "json",
	)
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "msg-output.golden")
}

func TestMsgOptionsError(t *testing.T) {
	fixture := initFixture(t)

	_, err := runCmd(fixture.conn, fixture.b, buildModuleMsgCommand,
		"send", "5",
	)
	assert.ErrorContains(t, err, "accepts 3 arg(s)")

	_, err = runCmd(fixture.conn, fixture.b, buildModuleMsgCommand,
		"send", "foo", "bar", "invalid",
	)
	assert.ErrorContains(t, err, "invalid argument")
}

func TestHelpMsg(t *testing.T) {
	fixture := initFixture(t)

	out, err := runCmd(fixture.conn, fixture.b, buildModuleMsgCommand, "-h")
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "help-toplevel-msg.golden")

	out, err = runCmd(fixture.conn, fixture.b, buildModuleMsgCommand, "send", "-h")
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "help-echo-msg.golden")
}

func TestBuildCustomMsgCommand(t *testing.T) {
	b := &Builder{}
	customCommandCalled := false
	appOptions := AppOptions{
		ModuleOptions: map[string]*autocliv1.ModuleOptions{
			"test": {
				Tx: &autocliv1.ServiceCommandDescriptor{
					Service:           testpb.Msg_ServiceDesc.ServiceName,
					RpcCommandOptions: []*autocliv1.RpcCommandOptions{},
				},
			},
		},
	}

	cmd, err := b.BuildMsgCommand(appOptions, map[string]*cobra.Command{
		"test": {Use: "test", Run: func(cmd *cobra.Command, args []string) {
			customCommandCalled = true
		}},
	})
	assert.NilError(t, err)
	cmd.SetArgs([]string{"test", "tx"})
	assert.NilError(t, cmd.Execute())
	assert.Assert(t, customCommandCalled)
}

<<<<<<< HEAD
func TestErrorBuildMsgCommand(t *testing.T) {
	fixture := initFixture(t)
	b := fixture.b
	b.AddQueryConnFlags = nil
	b.AddTxConnFlags = nil

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

	appOptions := AppOptions{
		ModuleOptions: map[string]*autocliv1.ModuleOptions{
			"test": {
				Tx: commandDescriptor,
			},
		},
		AddressCodec:          b.AddressCodec,
		ValidatorAddressCodec: b.ValidatorAddressCodec,
		ClientCtx:             b.ClientCtx,
	}

	_, err := b.BuildMsgCommand(appOptions, nil)
	assert.ErrorContains(t, err, "can't find field un-existent-proto-field")

	nonExistentService := &autocliv1.ServiceCommandDescriptor{Service: "un-existent-service"}
	appOptions.ModuleOptions["test"].Tx = nonExistentService
	_, err = b.BuildMsgCommand(appOptions, nil)
	assert.ErrorContains(t, err, "can't find service un-existent-service")
}

=======
>>>>>>> b62301d9d (feat(client/v2): signing (#17913))
func TestNotFoundErrorsMsg(t *testing.T) {
	fixture := initFixture(t)
	b := fixture.b
	b.AddQueryConnFlags = nil
	b.AddTxConnFlags = nil

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
