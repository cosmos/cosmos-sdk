package safety_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/x/safety"
)

// The x/circuit CLI examples joined several type URLs into one quoted
// argument. Check every positional arg names a real request field, varargs
// come last, examples pass each pattern as its own argument, and proposal
// examples pass the title and summary gov requires.
func TestAutoCLIOptions(t *testing.T) {
	opts := safety.AppModule{}.AutoCLIOptions()

	for _, svc := range []struct {
		name string
		rpcs []string
	}{
		{opts.Query.Service, []string{"Params", "DisabledList", "IsAllowed"}},
		{opts.Tx.Service, []string{"UpdateParams", "Trip", "Reset"}},
	} {
		desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(svc.name))
		require.NoError(t, err, svc.name)
		sd := desc.(protoreflect.ServiceDescriptor)
		for _, rpc := range svc.rpcs {
			require.NotNil(t, sd.Methods().ByName(protoreflect.Name(rpc)), rpc)
		}
	}

	for _, rpc := range append(opts.Query.RpcCommandOptions, opts.Tx.RpcCommandOptions...) {
		svcName := opts.Query.Service
		if rpc.RpcMethod == "Trip" || rpc.RpcMethod == "Reset" || rpc.RpcMethod == "UpdateParams" {
			svcName = opts.Tx.Service
		}
		desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(svcName))
		require.NoError(t, err)
		method := desc.(protoreflect.ServiceDescriptor).Methods().ByName(protoreflect.Name(rpc.RpcMethod))
		require.NotNil(t, method, rpc.RpcMethod)

		for i, arg := range rpc.PositionalArgs {
			require.NotNil(t, method.Input().Fields().ByName(protoreflect.Name(arg.ProtoField)),
				"%s: unknown field %s", rpc.RpcMethod, arg.ProtoField)
			if arg.Varargs {
				require.Equal(t, len(rpc.PositionalArgs)-1, i, "%s: varargs must be last", rpc.RpcMethod)
			}
		}

		if rpc.RpcMethod == "Trip" || rpc.RpcMethod == "Reset" {
			require.NotEmpty(t, rpc.Example)
			for line := range strings.SplitSeq(rpc.Example, "\n") {
				require.NotContains(t, line, `"/`, "%s: patterns must be separate args", rpc.RpcMethod)
				require.Contains(t, line, "tx safety "+strings.ToLower(rpc.RpcMethod)+" ")
			}
		}

		// gov rejects a proposal with an empty title or summary
		if rpc.GovProposal {
			require.Contains(t, rpc.Example, "--title ", rpc.RpcMethod)
			require.Contains(t, rpc.Example, "--summary ", rpc.RpcMethod)
		}
	}
}
