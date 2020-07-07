package bank_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

func getAccount(t *testing.T, app *simapp.SimApp, addr sdk.AccAddress) *auth.BaseAccount {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	res := app.AccountKeeper.GetAccount(ctxCheck, addr)
	require.NotNil(t, res)
	return res.(*auth.BaseAccount)
}

func simulatedGas(t *testing.T, app *simapp.SimApp, tx sdk.Tx) uint64 {
	txBytes, err := app.Codec().MarshalBinaryLengthPrefixed(tx)
	require.NoError(t, err)

	// use the public query interface
	req := abci.RequestQuery{
		Path: "/app/simulate",
		Data: txBytes,
	}
	res := app.Query(req)
	require.Equal(t, uint32(0), res.Code)
	var simRes sdk.SimulationResponse
	err = app.Codec().UnmarshalBinaryBare(res.Value, &simRes)
	require.NoError(t, err)
	simGas := simRes.GasInfo
	fmt.Printf("(SIM) wanted: %d / used: %d\n", simGas.GasWanted, simGas.GasUsed)
	return simGas.GasUsed
}

func TestSendGasEstimates(t *testing.T) {
	// some test accounts - addr1 has tokens
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2 := secp256k1.GenPrivKey()
	addr2 := sdk.AccAddress(priv2.PubKey().Address())

	initCoins := sdk.Coins{sdk.NewInt64Coin("uatom", 12345678)}
	acc := &auth.BaseAccount{
		Address: addr1,
		Coins:   initCoins,
	}

	genAccs := []authexported.GenesisAccount{acc}
	app := simapp.SetupWithGenesisAccounts(genAccs)

	// ensure proper balance
	acct := getAccount(t, app, addr1)
	require.Equal(t, acc, acct)
	simapp.CheckBalance(t, app, addr1, initCoins)

	send := sdk.Coins{sdk.NewInt64Coin("uatom", 5678)}
	sendMsg := types.NewMsgSend(addr1, addr2, send)
	header := abci.Header{Height: app.LastBlockHeight() + 1}

	// this will build proper tx
	buildTx := func(expectedGas uint64) sdk.Tx {
		return helpers.GenTx(
			[]sdk.Msg{sendMsg},
			sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
			expectedGas,
			"",
			[]uint64{acct.GetAccountNumber()},
			[]uint64{acct.GetSequence()},
			priv1,
		)
	}

	// run simulation
	tx := buildTx(200_000)
	simGas := simulatedGas(t, app, tx)

	// deliver the tx with the gas returned from simulate (plus 10%)
	tx = buildTx(simGas + simGas/10)
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	gas, res, err := app.Deliver(tx)
	require.NoError(t, err)
	require.NotNil(t, res)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// We have wanted 1 million, used ~53k
	fmt.Printf("wanted: %d / used: %d\n", gas.GasWanted, gas.GasUsed)
	// sanity checks for reasonable gas values
	assert.Less(t, uint64(40000), gas.GasUsed)
	assert.Less(t, gas.GasUsed, uint64(80000))
}
