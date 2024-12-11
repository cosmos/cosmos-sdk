package prompt

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/internal/testpb"

	address2 "github.com/cosmos/cosmos-sdk/codec/address"
)

func getReader(inputs []string) io.ReadCloser {
	// https://github.com/manifoldco/promptui/issues/63#issuecomment-621118463
	var paddedInputs []string
	for _, input := range inputs {
		padding := strings.Repeat("a", 4096-1-len(input)%4096)
		paddedInputs = append(paddedInputs, input+"\n"+padding)
	}
	return io.NopCloser(strings.NewReader(strings.Join(paddedInputs, "")))
}

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
			// https://github.com/manifoldco/promptui/issues/63#issuecomment-621118463
			var paddedInputs []string
			for _, input := range tt.inputs {
				padding := strings.Repeat("a", 4096-1-len(input)%4096)
				paddedInputs = append(paddedInputs, input+"\n"+padding)
			}
			reader := io.NopCloser(strings.NewReader(strings.Join(paddedInputs, "")))

			got, err := promptMessage(address2.NewBech32Codec("cosmos"), address2.NewBech32Codec("cosmosvaloper"), address2.NewBech32Codec("cosmosvalcons"), "prefix", reader, tt.msg)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
