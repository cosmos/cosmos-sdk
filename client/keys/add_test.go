package keys

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/spf13/cobra"
)

func Test_addKeyCommand(t *testing.T) {
	tests := []struct {
		name string
		want *cobra.Command
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addKeyCommand(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addKeyCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runAddCmd(t *testing.T) {
	type args struct {
		in0  *cobra.Command
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
			if err := runAddCmd(tt.args.in0, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runAddCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_printCreate(t *testing.T) {
	type args struct {
		info         keys.Info
		showMnemonic bool
		mnemonic     string
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
			if err := printCreate(tt.args.info, tt.args.showMnemonic, tt.args.mnemonic); (err != nil) != tt.wantErr {
				t.Errorf("printCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateMnemonic(t *testing.T) {
	type args struct {
		algo keys.SigningAlgo
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateMnemonic(tt.args.algo); got != tt.want {
				t.Errorf("generateMnemonic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckAndWriteErrorResponse(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		httpErr int
		err     error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckAndWriteErrorResponse(tt.args.w, tt.args.httpErr, tt.args.err); got != tt.want {
				t.Errorf("CheckAndWriteErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddNewKeyRequestHandler(t *testing.T) {
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
			if got := AddNewKeyRequestHandler(tt.args.indent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddNewKeyRequestHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeedRequestHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SeedRequestHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestRecoverRequestHandler(t *testing.T) {
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
			if got := RecoverRequestHandler(tt.args.indent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RecoverRequestHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
