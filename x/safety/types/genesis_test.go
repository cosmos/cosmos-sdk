package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

var (
	ac        = addresscodec.NewBech32Codec("cosmos")
	authority = sdk.AccAddress("authority___________").String()
	operator  = sdk.AccAddress("operator____________").String()
)

func TestParamsValidate(t *testing.T) {
	require.NoError(t, types.DefaultParams(authority).Validate(ac))

	cases := map[string]struct {
		params types.Params
		err    error
	}{
		"operator set":        {types.Params{Authority: authority, Operator: operator, Scope: []string{"/cosmos.bank.*"}}, nil},
		"empty scope":         {types.Params{Authority: authority, Operator: operator}, nil},
		"empty authority":     {types.Params{Scope: []string{"*"}}, types.ErrInvalidParams},
		"bad authority":       {types.Params{Authority: "cosmos1bad", Scope: []string{"*"}}, sdkerrors.ErrInvalidAddress},
		"bad operator":        {types.Params{Authority: authority, Operator: "nope"}, sdkerrors.ErrInvalidAddress},
		"bad scope pattern":   {types.Params{Authority: authority, Scope: []string{"cosmos.bank.*"}}, types.ErrInvalidPattern},
		"dup scope patterns":  {types.Params{Authority: authority, Scope: []string{"*", "*"}}, types.ErrDuplicatePattern},
		"uppercase authority": {types.Params{Authority: strings.ToUpper(authority)}, sdkerrors.ErrInvalidAddress},
		"uppercase operator":  {types.Params{Authority: authority, Operator: strings.ToUpper(operator)}, sdkerrors.ErrInvalidAddress},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.params.Validate(ac)
			if tc.err == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func TestGenesisValidate(t *testing.T) {
	require.NoError(t, types.DefaultGenesisState(authority).Validate(ac))

	gs := types.NewGenesisState(types.DefaultParams(authority), []string{"/cosmos.bank.*", "/cosmos.gov.v1.MsgVote", "*"})
	require.NoError(t, gs.Validate(ac))

	gs = types.NewGenesisState(types.DefaultParams(authority), []string{"/cosmos.bank.*", "/cosmos.bank.*"})
	require.ErrorIs(t, gs.Validate(ac), types.ErrDuplicatePattern)

	gs = types.NewGenesisState(types.DefaultParams(authority), []string{"cosmos.bank"})
	require.ErrorIs(t, gs.Validate(ac), types.ErrInvalidPattern)

	gs = types.NewGenesisState(types.Params{}, nil)
	require.ErrorIs(t, gs.Validate(ac), types.ErrInvalidParams)
}

func TestMsgValidate(t *testing.T) {
	require.NoError(t, types.NewMsgTrip(operator, "/cosmos.bank.*").Validate(ac))
	require.ErrorIs(t, types.NewMsgTrip(operator).Validate(ac), types.ErrEmptyPatterns)
	require.ErrorIs(t, types.NewMsgTrip(operator, "*", "*").Validate(ac), types.ErrDuplicatePattern)
	require.ErrorIs(t, types.NewMsgTrip(operator, "bank").Validate(ac), types.ErrInvalidPattern)
	require.ErrorIs(t, types.NewMsgTrip("", "*").Validate(ac), sdkerrors.ErrInvalidAddress)
	require.ErrorIs(t, types.NewMsgTrip("cosmos1xyz", "*").Validate(ac), sdkerrors.ErrInvalidAddress)

	require.NoError(t, types.NewMsgReset(authority, "/cosmos.bank.*").Validate(ac))
	require.ErrorIs(t, types.NewMsgReset(authority).Validate(ac), types.ErrEmptyPatterns)
	require.ErrorIs(t, types.NewMsgReset(authority, "/a.*", "/a.*").Validate(ac), types.ErrDuplicatePattern)
	require.ErrorIs(t, types.NewMsgReset("", "*").Validate(ac), sdkerrors.ErrInvalidAddress)
	require.ErrorIs(t, types.NewMsgReset("cosmos1xyz", "*").Validate(ac), sdkerrors.ErrInvalidAddress)

	up := &types.MsgUpdateParams{Authority: authority, Params: types.DefaultParams(operator)}
	require.NoError(t, up.Validate(ac))
	up.Authority = ""
	require.ErrorIs(t, up.Validate(ac), sdkerrors.ErrInvalidAddress)
	up.Authority = "cosmos1xyz"
	require.ErrorIs(t, up.Validate(ac), sdkerrors.ErrInvalidAddress)
	up = &types.MsgUpdateParams{Authority: authority, Params: types.Params{Authority: authority, Scope: []string{"x"}}}
	require.ErrorIs(t, up.Validate(ac), types.ErrInvalidPattern)

	// the operator must be canonical, as params.operator is, so EventTrip
	// names it exactly as params does
	require.ErrorIs(t, types.NewMsgTrip(strings.ToUpper(operator), "*").Validate(ac), sdkerrors.ErrInvalidAddress)
}
