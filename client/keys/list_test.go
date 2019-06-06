package keys

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
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
	viper.Set(flags.FlagHome, kbHome2)

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
			viper.Set(flags.FlagHome, tt.kbDir)
			if err := runListCmd(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runListCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
