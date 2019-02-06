package keys

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/spf13/cobra"
)

func Test_runListCmd(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	cmdBasic := listKeysCmd()
	keyBase1 := keys.New(dbm.NewMemDB())

	keyBase2 := keys.New(dbm.NewMemDB())
	_, err := keyBase2.CreateAccount("something", tests.TestMnemonic, "", "", 0, 0)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		kb      keys.Keybase
		args    args
		wantErr bool
	}{
		{"no keybase", nil, args{cmdBasic, []string{}}, true},
		{"with keybase1", keyBase1, args{cmdBasic, []string{}}, false},
		{"with keybase2", keyBase2, args{cmdBasic, []string{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetKeyBase(tt.kb)
			if err := runListCmd(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
