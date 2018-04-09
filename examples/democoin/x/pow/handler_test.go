package pow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
)

func TestPowHandler(t *testing.T) {
	ms, capKey := setupMultiStore()

	am := auth.NewAccountMapper(capKey, &auth.BaseAccount{})
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	config := NewPowConfig("pow", int64(1))
	ck := bank.NewCoinKeeper(am)
	keeper := NewKeeper(capKey, config, ck)

	handler := keeper.Handler

	addr := sdk.Address([]byte("sender"))
	count := uint64(1)
	difficulty := uint64(2)

	err := keeper.InitGenesis(ctx, PowGenesis{uint64(1), uint64(0)})
	assert.Nil(t, err)

	nonce, proof := mine(addr, count, difficulty)
	msg := NewMineMsg(addr, difficulty, count, nonce, proof)

	result := handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	newDiff, err := keeper.GetLastDifficulty(ctx)
	assert.Nil(t, err)
	assert.Equal(t, newDiff, uint64(2))

	newCount, err := keeper.GetLastCount(ctx)
	assert.Nil(t, err)
	assert.Equal(t, newCount, uint64(1))

	// todo assert correct coin change, awaiting https://github.com/cosmos/cosmos-sdk/pull/691

	difficulty = uint64(4)
	nonce, proof = mine(addr, count, difficulty)
	msg = NewMineMsg(addr, difficulty, count, nonce, proof)

	result = handler(ctx, msg)
	assert.NotEqual(t, result, sdk.Result{})
}
