package offchain

import (
	"github.com/cosmos/cosmos-sdk/client"
	"testing"
)

func TestSign(t *testing.T) {
	type args struct {
		ctx      client.Context
		rawBytes []byte
		fromName string
		encoding string
		output   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sign(tt.args.ctx, tt.args.rawBytes, tt.args.fromName, tt.args.encoding, tt.args.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Sign() got = %v, want %v", got, tt.want)
			}
		})
	}
}
