package bank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	secp256k1_internal "gitlab.com/yawning/secp256k1-voi"
	"gitlab.com/yawning/secp256k1-voi/secec"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	_ "cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/testutil"
	"cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus"
	_ "cosmossdk.io/x/distribution"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	_ "cosmossdk.io/x/gov"
	_ "cosmossdk.io/x/protocolpool"
	_ "cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	stablePrivateKey, _ = secec.NewPrivateKeyFromScalar(secp256k1_internal.NewScalarFromUint64(100))
	priv1               = &secp256k1.PrivKey{Key: stablePrivateKey.Bytes()}
	addr1               = sdk.AccAddress(priv1.PubKey().Address())
	priv2               = secp256k1.GenPrivKey()
	addr2               = sdk.AccAddress(priv2.PubKey().Address())
	addr3               = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	coins               = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	halfCoins           = sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}
	sendMsg1            = types.NewMsgSend(addr1.String(), addr2.String(), coins)
)

type suite struct {
	BankKeeper         bankkeeper.Keeper
	AccountKeeper      types.AccountKeeper
	DistributionKeeper distrkeeper.Keeper
	App                *integration.App
	TxConfig           client.TxConfig
}

func createTestSuite(t *testing.T, genesisAccounts []authtypes.GenesisAccount) suite {
	t.Helper()
	res := suite{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.BankModule(),
		configurator.GovModule(),
		configurator.DistributionModule(),
		configurator.ProtocolPoolModule(),
	}
	var err error
	startupCfg := integration.DefaultStartUpConfig()
	var genAccounts []integration.GenesisAccount
	for _, acc := range genesisAccounts {
		genAccounts = append(genAccounts, integration.GenesisAccount{GenesisAccount: acc})
	}
	startupCfg.GenesisAccounts = genAccounts
	startupCfg.HomeDir = t.TempDir()
	res.App, err = integration.SetupWithConfiguration(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.BankKeeper, &res.AccountKeeper, &res.DistributionKeeper, &res.TxConfig)
	require.NoError(t, err)

	return res
}

func TestSendNotEnoughBalance(t *testing.T) {
	acc := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc}
	s := createTestSuite(t, genAccs)
	ctx := context.Background()

	_, state, storeErr := s.App.Store.StateLatest()
	require.NoError(t, storeErr)

	_, err := s.App.Run(ctx, state, func(ctx context.Context) error {
		err := testutil.FundAccount(
			ctx, s.BankKeeper, addr1,
			sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67)))
		require.NoError(t, err)
		res1 := s.AccountKeeper.GetAccount(ctx, addr1)
		require.NotNil(t, res1)
		require.Equal(t, acc, res1.(*authtypes.BaseAccount))

		origAccNum := res1.GetAccountNumber()
		origSeq := res1.GetSequence()
		addr1Str, err := s.AccountKeeper.AddressCodec().BytesToString(addr1)
		require.NoError(t, err)
		addr2Str, err := s.AccountKeeper.AddressCodec().BytesToString(addr2)
		require.NoError(t, err)
		sendMsg := types.NewMsgSend(addr1Str, addr2Str, sdk.Coins{sdk.NewInt64Coin("foocoin", 100)})

		// TODO how to auto-advance height with app v2 interface?
		s.App.SignCheckDeliver(
			t, ctx, 2, []sdk.Msg{sendMsg}, "", []uint64{origAccNum}, []uint64{origSeq},
			[]cryptotypes.PrivKey{priv1},
			"spendable balance 67foocoin is smaller than 100foocoin",
		)
		s.App.CheckBalance(ctx, t, addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 67)}, s.BankKeeper)
		res2 := s.AccountKeeper.GetAccount(ctx, addr1)
		require.NotNil(t, res2)

		require.Equal(t, origAccNum, res2.GetAccountNumber())
		require.Equal(t, origSeq+1, res2.GetSequence())

		return nil
	})
	require.NoError(t, err)
}
