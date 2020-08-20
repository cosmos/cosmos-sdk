package types_test

import (
	pt "github.com/gogo/protobuf/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/params/types"
)

func TestKeyTable(t *testing.T) {
	table := types.NewKeyTable()

	require.Panics(t, func() { table.RegisterType(types.ParamSetPair{[]byte(""), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(types.ParamSetPair{[]byte("!@#$%"), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(types.ParamSetPair{[]byte("hello,"), nil, nil}) })
	require.Panics(t, func() { table.RegisterType(types.ParamSetPair{[]byte("hello"), nil, nil}) })

	require.NotPanics(t, func() {
		table.RegisterType(types.ParamSetPair{keyBondDenom, &pt.StringValue{Value: "stake"}, validateBondDenom})
	})
	require.NotPanics(t, func() {
		table.RegisterType(types.ParamSetPair{keyMaxValidators, &pt.UInt64Value{Value: 100}, validateMaxValidators})
	})
	require.Panics(t, func() {
		table.RegisterType(types.ParamSetPair{keyUnbondingTime, pt.DurationProto(time.Hour * 1), nil})
	})
	require.NotPanics(t, func() {
		table.RegisterType(types.ParamSetPair{keyUnbondingTime, pt.DurationProto(time.Hour * 1), validateMaxValidators})
	})
	require.NotPanics(t, func() {
		newTable := types.NewKeyTable()
		newTable.RegisterParamSet(&params{})
	})

	require.Panics(t, func() { table.RegisterParamSet(&params{}) })
	require.Panics(t, func() { types.NewKeyTable(types.ParamSetPair{[]byte(""), nil, nil}) })

	require.NotPanics(t, func() {
		types.NewKeyTable(
			types.ParamSetPair{[]byte("test"), &pt.StringValue{Value: "stake"}, validateBondDenom},
			types.ParamSetPair{[]byte("test2"), &pt.UInt64Value{Value: 100}, validateMaxValidators},
		)
	})
}
