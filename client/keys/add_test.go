package keys

import (
	"bufio"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

func Test_runAddCmdBasic(t *testing.T) {
	cmd := addKeyCommand()
	assert.NotNil(t, cmd)

	// Empty
	err := runAddCmd(cmd, []string{})
	assert.EqualError(t, err, "not enough arguments")

	// Missing input (enter password)
	err = runAddCmd(cmd, []string{"keyname"})
	assert.EqualError(t, err, "EOF")

	// Prepare a keybase
	kbHome, kbCleanUp, err := tests.GetTempDir("Test_runDeleteCmd")
	assert.NoError(t, err)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(cli.HomeFlag, kbHome)
	viper.Set(cli.OutputFlag, OutputFormatText)

	{
		// Now enter password
		cleanUp := client.OverrideStdin(bufio.NewReader(strings.NewReader("test1234\ntest1234\n")))
		defer cleanUp()
		err = runAddCmd(cmd, []string{"keyname"})
		assert.NoError(t, err)
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
