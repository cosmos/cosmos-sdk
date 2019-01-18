package pow

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPowHandler(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	config := NewConfig("pow", int64(1))
	keeper := NewKeeper(input.capKey, config, input.bk, DefaultCodespace)

	handler := keeper.Handler

	addr := sdk.AccAddress([]byte("sender"))
	count := uint64(1)
	difficulty := uint64(2)

	err := InitGenesis(ctx, keeper, Genesis{uint64(1), uint64(0)})
	require.Nil(t, err)

	nonce, proof := mine(addr, count, difficulty)
	msg := NewMsgMine(addr, difficulty, count, nonce, proof)

	result := handler(ctx, msg)
	require.Equal(t, result, sdk.Result{})

	newDiff, err := keeper.GetLastDifficulty(ctx)
	require.Nil(t, err)
	require.Equal(t, newDiff, uint64(2))

	newCount, err := keeper.GetLastCount(ctx)
	require.Nil(t, err)
	require.Equal(t, newCount, uint64(1))

	// todo assert correct coin change, awaiting https://github.com/cosmos/cosmos-sdk/pull/691

	difficulty = uint64(4)
	nonce, proof = mine(addr, count, difficulty)
	msg = NewMsgMine(addr, difficulty, count, nonce, proof)

	result = handler(ctx, msg)
	require.NotEqual(t, result, sdk.Result{})
}
