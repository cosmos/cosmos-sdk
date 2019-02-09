package keys

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/stretchr/testify/assert"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/spf13/cobra"
)

func Test_runListCmd(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	cmdBasic := listKeysCmd()

	// Prepare some keybases
	kbHome1, cleanUp1 := tests.NewTestCaseDir(t)
	defer cleanUp1()
	// Do nothing, leave home1 empty

	kbHome2, cleanUp2 := tests.NewTestCaseDir(t)
	defer cleanUp2()
	viper.Set(cli.HomeFlag, kbHome2)

	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	_, err = kb.CreateAccount("something", tests.TestMnemonic, "", "", 0, 0)
	assert.NoError(t, err)

	testData := []struct {
		name    string
		kbDir   string
		args    args
		wantErr bool
	}{
		{"invalid keybase", "/dev/null", args{cmdBasic, []string{}}, true},
		{"keybase: empty", kbHome1, args{cmdBasic, []string{}}, false},
		{"keybase: w/key", kbHome2, args{cmdBasic, []string{}}, false},
	}
	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(cli.HomeFlag, tt.kbDir)
			if err := runListCmd(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
