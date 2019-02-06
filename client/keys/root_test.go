package keys

import (
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

func TestCommands(t *testing.T) {
	tests := []struct {
		name string
		want *cobra.Command
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Commands(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Commands() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterRoutes(t *testing.T) {
	type args struct {
		r      *mux.Router
		indent bool
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterRoutes(tt.args.r, tt.args.indent)
		})
	}
}
