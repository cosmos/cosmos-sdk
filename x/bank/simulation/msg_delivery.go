package simulation

import (
	"math/rand"

	"cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	types3 "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	types2 "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// sendMsgSend sends a transaction with a MsgSend from a provided random account.
func sendMsgSend(
	r *rand.Rand, app *baseapp.BaseApp,
	txGen client.TxConfig,
	bk keeper.Keeper, ak types.AccountKeeper,
	msg *types.MsgSend, ctx types2.Context, chainID string, privkeys []types3.PrivKey,
) error {
	var (
		fees types2.Coins
		err  error
	)

	from, err := ak.AddressCodec().StringToBytes(msg.FromAddress)
	if err != nil {
		return err
	}

	account := ak.GetAccount(ctx, from)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	coins, hasNeg := spendable.SafeSub(msg.Amount...)
	if !hasNeg {
		fees, err = simulation.RandomFees(r, coins)
		if err != nil {
			return err
		}
	}
	tx, err := sims.GenSignedMockTx(
		r,
		txGen,
		[]types2.Msg{msg},
		fees,
		sims.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		privkeys...,
	)
	if err != nil {
		return err
	}

	_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
	if err != nil {
		return err
	}

	return nil
}
