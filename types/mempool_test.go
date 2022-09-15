package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func TestNewBTreeMempool(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000)
	require.Equal(t, 1000, len(transactions))
	mempool := sdk.NewBTreeMempool(1000)

	for _, tx := range transactions {
		ctx.WithPriority(rand.Int63())
		err := mempool.Insert(ctx, tx.(sdk.MempoolTx))
		require.NoError(t, err)
	}
}

func simulateManyTx(ctx sdk.Context, n int) []sdk.Tx {
	transactions := make([]sdk.Tx, n)
	for i := 0; i < n; i++ {
		tx := simulateTx(ctx)
		transactions[i] = tx
	}
	return transactions
}

func simulateTx(ctx sdk.Context) sdk.Tx {
	acc := authtypes.NewEmptyModuleAccount("anaccount")

	s := rand.NewSource(1)
	r := rand.New(s)
	msg := group.MsgUpdateGroupMembers{
		GroupId:       1,
		Admin:         "test",
		MemberUpdates: []group.MemberRequest{},
	}
	fees, _ := simtypes.RandomFees(r, ctx, sdk.NewCoins(sdk.NewCoin("coin", sdk.NewInt(100000000))))

	txGen := moduletestutil.MakeTestEncodingConfig().TxConfig
	accounts := simtypes.RandomAccounts(r, 2)

	tx, _ := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{&msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{acc.GetAccountNumber()},
		[]uint64{acc.GetSequence()},
		accounts[0].PrivKey,
	)
	return tx
}
