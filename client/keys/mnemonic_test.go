package keys

import (
	"testing"

	"github.com/spf13/cobra"
)

func Test_RunMnemonicCmd(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	cmdBasic := mnemonicKeyCommand()
	cmdUser := mnemonicKeyCommand()
	cmdUser.Flags().Set(flagUserEntropy, "1")

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"mne1", args{cmdBasic, []string{}}, false},

		// TODO: Requires stdin mocking
		// {"mne2", args{cmdUser, []string{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runMnemonicCmd(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runMnemonicCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
