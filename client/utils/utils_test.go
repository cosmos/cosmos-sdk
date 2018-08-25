package utils

import (
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/common"
)

func TestParseQueryResponse(t *testing.T) {
	cdc := app.MakeCodec()
	sdkResBytes := cdc.MustMarshalBinary(sdk.Result{GasUsed: 10})
	gas, err := parseQueryResponse(cdc, sdkResBytes)
	assert.Equal(t, gas, int64(10))
	assert.Nil(t, err)
	gas, err = parseQueryResponse(cdc, []byte("fuzzy"))
	assert.Equal(t, gas, int64(0))
	assert.NotNil(t, err)
}

func TestCalculateGas(t *testing.T) {
	cdc := app.MakeCodec()
	makeQueryFunc := func(gasUsed int64, wantErr bool) func(string, common.HexBytes) ([]byte, error) {
		return func(string, common.HexBytes) ([]byte, error) {
			if wantErr {
				return nil, errors.New("")
			}
			return cdc.MustMarshalBinary(sdk.Result{GasUsed: gasUsed}), nil
		}
	}
	type args struct {
		queryFuncGasUsed int64
		queryFuncWantErr bool
		adjustment       float64
	}
	tests := []struct {
		name         string
		args         args
		wantEstimate int64
		wantAdjusted int64
		wantErr      bool
	}{
		{"error", args{0, true, 1.2}, 0, 0, true},
		{"adjusted gas", args{10, false, 1.2}, 10, 12, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryFunc := makeQueryFunc(tt.args.queryFuncGasUsed, tt.args.queryFuncWantErr)
			gotEstimate, gotAdjusted, err := CalculateGas(queryFunc, cdc, []byte(""), tt.args.adjustment)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, gotEstimate, tt.wantEstimate)
			assert.Equal(t, gotAdjusted, tt.wantAdjusted)
		})
	}
}
