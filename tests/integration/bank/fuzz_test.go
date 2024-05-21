package bank_test

import (
	"math/rand"
	"testing"

	authtypes "cosmossdk.io/x/auth/types"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banksims "cosmossdk.io/x/bank/simulation"
	"cosmossdk.io/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/stretchr/testify/require"
)

func FuzzBankSend(f *testing.F) {
	const n = 1000
	simAccs := make([]simtypes.Account, n)
	for i := 0; i < n; i++ {
		priv := secp256k1.GenPrivKey()
		simAccs[i] = simtypes.Account{PrivKey: priv, PubKey: priv.PubKey(), Address: sdk.AccAddress(priv.PubKey().Address())}
	}
	s := createTestSuite(f, simsx.Collect(simAccs, func(a simtypes.Account) authtypes.GenesisAccount {
		return &authtypes.BaseAccount{Address: a.Address.String()}
	}))
	bk, ak := s.BankKeeper, s.AccountKeeper
	pCtx := s.App.BaseApp.NewContext(false)
	for i := 0; i < n; i++ {
		require.NoError(f, testutil.FundAccount(pCtx, bk, simAccs[i].Address, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000))))
	}
	bankWithContext := simsx.SpendableCoinserFn(func(addr sdk.AccAddress) sdk.Coins {
		return bk.SpendableCoins(pCtx, addr)
	})
	pReporter := &simsx.BasicSimulationReporter{}
	factory := banksims.MsgSendFactory(bk)
	f.Fuzz(func(t *testing.T, rawSeed []byte) {
		if len(rawSeed) < 8 {
			t.Skip()
			return
		}
		testData := simsx.NewChainDataSource(rand.New(simulation.NewByteSource(rawSeed, 1)), ak, bankWithContext, ak.AddressCodec(), simAccs...)
		reporter := pReporter.WithScope(factory.MsgType())
		ctx, _ := pCtx.CacheContext()
		_, msg := factory(ctx, testData, reporter)
		_, err := bankkeeper.NewMsgServerImpl(bk).Send(ctx, factory.Cast(msg))
		require.NoError(t, err)
		if reporter.IsSkipped() {
			t.Skip(reporter.Comment())
			return
		}
	})
}
