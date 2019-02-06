package keys

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto"
)

func Test_multiSigKey_GetName(t *testing.T) {
	tests := []struct {
		name string
		m    multiSigKey
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.GetName(); got != tt.want {
				t.Errorf("multiSigKey.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_multiSigKey_GetType(t *testing.T) {
	tests := []struct {
		name string
		m    multiSigKey
		want keys.KeyType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.GetType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("multiSigKey.GetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_multiSigKey_GetPubKey(t *testing.T) {
	tests := []struct {
		name string
		m    multiSigKey
		want crypto.PubKey
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.GetPubKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("multiSigKey.GetPubKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_multiSigKey_GetAddress(t *testing.T) {
	tests := []struct {
		name string
		m    multiSigKey
		want sdk.AccAddress
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.GetAddress(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("multiSigKey.GetAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_showKeysCmd(t *testing.T) {
	tests := []struct {
		name string
		want *cobra.Command
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := showKeysCmd(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("showKeysCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runShowCmd(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runShowCmd(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runShowCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateMultisigThreshold(t *testing.T) {
	type args struct {
		k     int
		nKeys int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateMultisigThreshold(tt.args.k, tt.args.nKeys); (err != nil) != tt.wantErr {
				t.Errorf("validateMultisigThreshold() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getBechKeyOut(t *testing.T) {
	type args struct {
		bechPrefix string
	}
	tests := []struct {
		name    string
		args    args
		want    bechKeyOutFn
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBechKeyOut(tt.args.bechPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBechKeyOut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBechKeyOut() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKeyRequestHandler(t *testing.T) {
	type args struct {
		indent bool
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetKeyRequestHandler(tt.args.indent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKeyRequestHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
