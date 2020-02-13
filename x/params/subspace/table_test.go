package subspace_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

func TestKeyTable(t *testing.T) {
	table := subspace.NewKeyTable()

	require.Panics(t, func() { table.RegisterType(subspace.ParamSetPair{[]byte(""), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(subspace.ParamSetPair{[]byte("!@#$%"), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(subspace.ParamSetPair{[]byte("hello,"), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(subspace.ParamSetPair{[]byte("hello"), nil, nil}) })

	require.NotPanics(t, func() {
		table.RegisterType(subspace.ParamSetPair{keyBondDenom, string("stake"), validateBondDenom})
	})
	require.NotPanics(t, func() {
		table.RegisterType(subspace.ParamSetPair{keyMaxValidators, uint16(100), validateMaxValidators})
	})
	require.Panics(t, func() {
		table.RegisterType(subspace.ParamSetPair{keyUnbondingTime, time.Duration(1), nil})
	})
	require.NotPanics(t, func() {
		table.RegisterType(subspace.ParamSetPair{keyUnbondingTime, time.Duration(1), validateMaxValidators})
	})
	require.NotPanics(t, func() {
		newTable := subspace.NewKeyTable()
		newTable.RegisterParamSet(&params{})
	})

	require.Panics(t, func() { table.RegisterParamSet(&params{}) })
	require.Panics(t, func() { subspace.NewKeyTable(subspace.ParamSetPair{[]byte(""), nil, nil}) })

	require.NotPanics(t, func() {
		subspace.NewKeyTable(
			subspace.ParamSetPair{[]byte("test"), string("stake"), validateBondDenom},
			subspace.ParamSetPair{[]byte("test2"), uint16(100), validateMaxValidators},
		)
	})
}
