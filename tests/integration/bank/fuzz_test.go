package bank_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	authtypes "cosmossdk.io/x/auth/types"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banksims "cosmossdk.io/x/bank/simulation"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simsx"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

func FuzzBankSend(f *testing.F) {
	const n = 500
	simAccs := make([]simtypes.Account, n)
	for i := 0; i < n; i++ {
		priv := secp256k1.GenPrivKey()
		simAccs[i] = simtypes.Account{PrivKey: priv, PubKey: priv.PubKey(), Address: sdk.AccAddress(priv.PubKey().Address())}
	}
	s := createTestSuiteX(f, simsx.Collect(simAccs, func(a simtypes.Account) simtestutil.GenesisAccount {
		return simtestutil.GenesisAccount{
			GenesisAccount: authtypes.NewBaseAccount(a.Address, a.PubKey, 0, 0),
			Coins:          sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000)),
		}
	}))
	bk, ak := s.BankKeeper, s.AccountKeeper
	pCtx := s.App.BaseApp.NewContext(false)
	factory := banksims.MsgSendFactory()

	f.Add([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
	f.Fuzz(func(t *testing.T, rawSeed []byte) {
		if len(rawSeed) < 8 {
			t.Skip()
			return
		}
		start := sdk.BigEndianToUint64(rawSeed[0:8])
		r := rand.New(simulation.NewByteSource(rawSeed[8:], int64(start)))
		testData := simsx.NewChainDataSource(pCtx, r, ak, bk, ak.AddressCodec(), simAccs...)
		reporter := simsx.NewBasicSimulationReporter(t).WithScope(factory.MsgType())
		ctx, _ := pCtx.CacheContext()
		_, msg := factory(ctx, testData, reporter)
		_, err := bankkeeper.NewMsgServerImpl(bk).Send(ctx, factory.Cast(msg))
		require.NoError(t, err)
	})
}
