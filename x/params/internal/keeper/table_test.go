package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/params/internal/keeper"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

func TestKeyTable(t *testing.T) {
	table := NewKeyTable()

	require.Panics(t, func() { table.RegisterType(keeper.ParamSetPair{[]byte(""), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(keeper.ParamSetPair{[]byte("!@#$%"), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(keeper.ParamSetPair{[]byte("hello,"), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(keeper.ParamSetPair{[]byte("hello"), nil, nil}) })

	require.NotPanics(t, func() {
		table.RegisterType(keeper.ParamSetPair{subspace.keyBondDenom, string("stake"), subspace.validateBondDenom})
	})
	require.NotPanics(t, func() {
		table.RegisterType(keeper.ParamSetPair{subspace.keyMaxValidators, uint16(100), subspace.validateMaxValidators})
	})
	require.Panics(t, func() {
		table.RegisterType(keeper.ParamSetPair{subspace.keyUnbondingTime, time.Duration(1), nil})
	})
	require.NotPanics(t, func() {
		table.RegisterType(keeper.ParamSetPair{subspace.keyUnbondingTime, time.Duration(1), subspace.validateMaxValidators})
	})
	require.NotPanics(t, func() {
		newTable := NewKeyTable()
		newTable.RegisterParamSet(&subspace.params{})
	})

	require.Panics(t, func() { table.RegisterParamSet(&subspace.params{}) })
	require.Panics(t, func() { NewKeyTable(keeper.ParamSetPair{[]byte(""), nil, nil}) })

	require.NotPanics(t, func() {
		NewKeyTable(
			keeper.ParamSetPair{[]byte("test"), string("stake"), subspace.validateBondDenom},
			keeper.ParamSetPair{[]byte("test2"), uint16(100), subspace.validateMaxValidators},
		)
	})
}
