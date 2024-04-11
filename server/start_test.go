package server

import (
	"reflect"
	"runtime"
	"testing"
)

func TestAbciClientType(t *testing.T) {
	for _, tt := range []struct {
		clientType  string
		creatorName string
		wantErr     bool
	}{
		{
			clientType:  "committing",
			creatorName: "github.com/tendermint/tendermint/proxy.NewCommittingClientCreator",
		},
		{
			clientType:  "local",
			creatorName: "github.com/tendermint/tendermint/proxy.NewLocalClientCreator",
		},
		{
			clientType: "cool ranch",
			wantErr:    true,
		},
	} {
		t.Run(tt.clientType, func(t *testing.T) {
			creator, err := getAbciClientCreator(tt.clientType)
			if tt.wantErr {
				if err == nil {
					t.Error("wanted error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				} else {
					creatorName := runtime.FuncForPC(reflect.ValueOf(creator).Pointer()).Name()
					if creatorName != tt.creatorName {
						t.Errorf(`want creator "%s", got "%s"`, tt.creatorName, creatorName)
					}
				}
			}
		})
	}
}
