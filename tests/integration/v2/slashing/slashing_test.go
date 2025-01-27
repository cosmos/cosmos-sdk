package slashing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	_ "cosmossdk.io/x/accounts" // import as blank for app wiring
	_ "cosmossdk.io/x/bank"     // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	_ "cosmossdk.io/x/consensus"    // import as blank for app wiring
	_ "cosmossdk.io/x/distribution" // import as blank for app wiring
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	_ "cosmossdk.io/x/mint"         // import as blank for app wiring
	_ "cosmossdk.io/x/protocolpool" // import as blank for app wiring
	_ "cosmossdk.io/x/slashing"     // import as blank for app wiring
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	"cosmossdk.io/x/slashing/testutil"
	slashingtypes "cosmossdk.io/x/slashing/types"
	_ "cosmossdk.io/x/staking" // import as blank for app wiring
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/genutil" // import as blank for app wiring
)

var (
	priv1        = secp256k1.GenPrivKey()
	addr1        = sdk.AccAddress(priv1.PubKey().Address())
	valaddrCodec = codecaddress.NewBech32Codec("cosmosvaloper")

	valKey  = ed25519.GenPrivKey()
	valAddr = sdk.AccAddress(valKey.PubKey().Address())
)

type fixture struct {
	app *integration.App

	ctx context.Context

	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	slashingKeeper slashingkeeper.Keeper
	stakingKeeper  *stakingkeeper.Keeper
	distrKeeper    distrkeeper.Keeper

	slashingMsgServer slashingtypes.MsgServer
	txConfig          client.TxConfig

	valAddrs []sdk.ValAddress
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	res := fixture{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.StakingModule(),
		configurator.SlashingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
		configurator.MintModule(),
		configurator.DistributionModule(),
		configurator.ProtocolPoolModule(),
	}

	acc := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	var err error
	startupCfg := integration.DefaultStartUpConfig(t)
	startupCfg.GenesisAccounts = []integration.GenesisAccount{{GenesisAccount: acc}}

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.HeaderService = &integration.HeaderService{}

	res.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.bankKeeper, &res.accountKeeper, &res.stakingKeeper, &res.slashingKeeper, &res.distrKeeper, &res.txConfig)
	require.NoError(t, err)

	res.ctx = res.app.StateLatestContext(t)

	// set default staking params
	// TestParams set the SignedBlocksWindow to 1000 and MaxMissedBlocksPerWindow to 500
	err = res.slashingKeeper.Params.Set(res.ctx, testutil.TestParams())
	assert.NilError(t, err)

	addrDels := simtestutil.AddTestAddrsIncremental(res.bankKeeper, res.stakingKeeper, res.ctx, 6, res.stakingKeeper.TokensFromConsensusPower(res.ctx, 200))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)

	consaddr0, err := res.stakingKeeper.ConsensusAddressCodec().BytesToString(addrDels[0])
	require.NoError(t, err)
	consaddr1, err := res.stakingKeeper.ConsensusAddressCodec().BytesToString(addrDels[1])
	require.NoError(t, err)

	info1 := slashingtypes.NewValidatorSigningInfo(consaddr0, int64(4), time.Unix(2, 0), false, int64(10))
	info2 := slashingtypes.NewValidatorSigningInfo(consaddr1, int64(5), time.Unix(2, 0), false, int64(10))

	err = res.slashingKeeper.ValidatorSigningInfo.Set(res.ctx, sdk.ConsAddress(addrDels[0]), info1)
	require.NoError(t, err)
	err = res.slashingKeeper.ValidatorSigningInfo.Set(res.ctx, sdk.ConsAddress(addrDels[1]), info2)
	require.NoError(t, err)

	res.valAddrs = valAddrs
	res.slashingMsgServer = slashingkeeper.NewMsgServerImpl(res.slashingKeeper)

	return &res
}

func TestSlashingMsgs(t *testing.T) {
	f := initFixture(t)

	genTokens := sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	require.NoError(t, banktestutil.FundAccount(f.ctx, f.bankKeeper, addr1, sdk.NewCoins(genCoin)))

	description := stakingtypes.NewDescription("foo_moniker", "", "", "", "", &stakingtypes.Metadata{})
	commission := stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())

	addrStrVal, err := valaddrCodec.BytesToString(addr1)
	require.NoError(t, err)
	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		addrStrVal, valKey.PubKey(), bondCoin, description, commission, math.OneInt(),
	)
	require.NoError(t, err)

	_ = f.app.SignCheckDeliver(t, f.ctx, []sdk.Msg{createValidatorMsg}, "", []uint64{0}, []uint64{0}, []cryptotypes.PrivKey{priv1}, "")
	require.True(t, sdk.Coins{genCoin.Sub(bondCoin)}.Equal(f.bankKeeper.GetAllBalances(f.ctx, addr1)))

	validator, err := f.stakingKeeper.GetValidator(f.ctx, sdk.ValAddress(addr1))
	require.NoError(t, err)

	require.Equal(t, addrStrVal, validator.OperatorAddress)
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.True(math.IntEq(t, bondTokens, validator.BondedTokens()))
	unjailMsg := &slashingtypes.MsgUnjail{ValidatorAddr: addrStrVal}

	_, err = f.slashingKeeper.ValidatorSigningInfo.Get(f.ctx, sdk.ConsAddress(valAddr))
	require.NoError(t, err)

	// unjail should fail with validator not jailed error
	_ = f.app.SignCheckDeliver(t, f.ctx, []sdk.Msg{unjailMsg}, "", []uint64{0}, []uint64{1}, []cryptotypes.PrivKey{priv1}, slashingtypes.ErrValidatorNotJailed.Error())
}
