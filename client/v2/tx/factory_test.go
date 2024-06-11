package tx

import (
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"
	"github.com/cosmos/gogoproto/grpc"
	"reflect"
	"testing"
)

func TestNewFactory(t *testing.T) {
	type args struct {
		keybase    keyring.Keyring
		txConfig   TxConfig
		ac         address.Codec
		conn       grpc.ClientConn
		parameters TxParameters
	}
	tests := []struct {
		name    string
		args    args
		want    Factory
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFactory(tt.args.keybase, tt.args.txConfig, tt.args.ac, tt.args.conn, tt.args.parameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFactory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFactory() got = %v, want %v", got, tt.want)
			}
		})
	}
}
