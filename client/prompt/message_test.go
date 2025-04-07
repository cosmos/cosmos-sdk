package prompt

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/client/prompt/internal/testpb"
	address2 "github.com/cosmos/cosmos-sdk/codec/address"
)

// getReader creates an io.ReadCloser for testing input simulation.
// it provides inputs as newline-separated strings.
func getReader(inputs []string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(strings.Join(inputs, "\n")))
}

// TestPromptMessage tests the standard library implementation of the message prompting system.
// It verifies that various input types are correctly handled when populating protobuf messages.
func TestPromptMessage(t *testing.T) {
	tests := []struct {
		name   string
		msg    protoreflect.Message
		inputs []string
	}{
		{
			name: "testPb",
			inputs: []string{
				"1", "2", "string", "bytes", "10101010", "0", "234234", "3", "4", "5", "true", "ENUM_ONE",
				"bar", "6", "10000", "stake", "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				"bytes", "6", "7", "false", "false", "true,false,true", "1,2,3", "hello,hola,ciao", "ENUM_ONE,ENUM_TWO",
				"10239", "0", "No", "bar", "343", "No", "134", "positional2", "23455", "stake", "No", "deprecate",
				"shorthand", "false", "cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z",
			},
			msg: (&testpb.MsgRequest{}).ProtoReflect(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := getReader(tt.inputs)

			// test with the standard library-based implementation
			got, err := promptMessage(
				address2.NewBech32Codec("cosmos"),
				address2.NewBech32Codec("cosmosvaloper"),
				address2.NewBech32Codec("cosmosvalcons"),
				"prefix",
				reader,
				tt.msg,
			)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
