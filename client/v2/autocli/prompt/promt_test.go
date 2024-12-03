package prompt

import (
	govtypes "cosmossdk.io/api/cosmos/gov/v1beta1"

	address2 "github.com/cosmos/cosmos-sdk/codec/address"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"testing"
)

func TestPrompt(t *testing.T) {
	tests := []struct {
		name    string
		want    protoreflect.Message
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Prompt(address2.NewBech32Codec("cosmos"), address2.NewBech32Codec("cosmosval"), address2.NewBech32Codec("cosmos"), "prefix", (&govtypes.MsgSubmitProposal{}).ProtoReflect())
			if (err != nil) != tt.wantErr {
				t.Errorf("Prompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Prompt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
