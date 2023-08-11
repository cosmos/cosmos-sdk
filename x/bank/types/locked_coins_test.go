package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestVestingLockedContextFuncs(t *testing.T) {
	tests := []struct {
		name string
		ctx  sdk.Context
		exp  bool
	}{
		{
			name: "brand new mostly empty context",
			ctx:  sdk.NewContext(nil, tmproto.Header{}, false, nil),
			exp:  false,
		},
		{
			name: "context with bypass",
			ctx:  WithVestingLockedBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil)),
			exp:  true,
		},
		{
			name: "context with bypass on one that originally was without it",
			ctx:  WithVestingLockedBypass(WithoutVestingLockedBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil))),
			exp:  true,
		},
		{
			name: "context with bypass twice",
			ctx:  WithVestingLockedBypass(WithVestingLockedBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil))),
			exp:  true,
		},
		{
			name: "context without bypass",
			ctx:  WithoutVestingLockedBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil)),
			exp:  false,
		},
		{
			name: "context without bypass on one that originally had it",
			ctx:  WithoutVestingLockedBypass(WithVestingLockedBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil))),
			exp:  false,
		},
		{
			name: "context without bypass twice",
			ctx:  WithoutVestingLockedBypass(WithoutVestingLockedBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil))),
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := HasVestingLockedBypass(tc.ctx)
			assert.Equal(t, tc.exp, actual, "HasVestingLockedBypass")
		})
	}
}

func TestContextFuncsDoNotModifyProvided(t *testing.T) {
	origCtx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	assert.False(t, HasVestingLockedBypass(origCtx), "HasVestingLockedBypass(origCtx)")
	afterWith := WithVestingLockedBypass(origCtx)
	assert.True(t, HasVestingLockedBypass(afterWith), "HasVestingLockedBypass(afterWith)")
	assert.False(t, HasVestingLockedBypass(origCtx), "HasVestingLockedBypass(origCtx) after giving it to WithVestingLockedBypass")
	afterWithout := WithoutVestingLockedBypass(afterWith)
	assert.False(t, HasVestingLockedBypass(afterWithout), "HasVestingLockedBypass(afterWithout)")
	assert.True(t, HasVestingLockedBypass(afterWith), "HasVestingLockedBypass(afterWith) after giving it to WithoutVestingLockedBypass")
	assert.False(t, HasVestingLockedBypass(origCtx), "HasVestingLockedBypass(origCtx) after giving afterWith to WithoutVestingLockedBypass")
}

func TestKeyContainsSpecificName(t *testing.T) {
	assert.Contains(t, bypassKey, "vesting", "bypassKey")
	assert.Contains(t, bypassKey, "bypass", "bypassKey")
}
