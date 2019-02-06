package keys

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func Test_updateKeyCommand(t *testing.T) {
	tests := []struct {
		name string
		want *cobra.Command
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateKeyCommand(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateKeyCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runUpdateCmd(t *testing.T) {
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
			if err := runUpdateCmd(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("runUpdateCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateKeyRequestHandler(t *testing.T) {
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
			UpdateKeyRequestHandler(tt.args.w, tt.args.r)
		})
	}
}
