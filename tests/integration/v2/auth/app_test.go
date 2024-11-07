package auth

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/accounts" // import as blank for app wiring
	"cosmossdk.io/x/accounts/accountstd"
	baseaccount "cosmossdk.io/x/accounts/defaults/base"
	"cosmossdk.io/x/bank"
	_ "cosmossdk.io/x/bank" // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus" // import as blank for app wiring
	minttypes "cosmossdk.io/x/mint/types"
	_ "cosmossdk.io/x/staking" // import as blank for app wirings
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting" // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"      // import as blank for app wiring
	"github.com/stretchr/testify/require"
)

var _ signing.SignModeHandler = directHandler{}

type directHandler struct{}

func (s directHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT
}

func (s directHandler) GetSignBytes(_ context.Context, _ signing.SignerData, _ signing.TxData) ([]byte, error) {
	panic("not implemented")
}

type suite struct {
	app *integration.App

	cdc codec.Codec
	ctx sdk.Context

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.Keeper
}

func (s suite) mustAddr(address []byte) string {
	str, _ := s.authKeeper.AddressCodec().BytesToString(address)
	return str
}

func createTestSuite(t *testing.T, genesisAccounts []authtypes.GenesisAccount, extraAccs map[string]accountstd.Interface) *suite {
	t.Helper()
	res := suite{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.VestingModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
	}

	var err error
	startupCfg := integration.DefaultStartUpConfig(t)
	var genAccounts []integration.GenesisAccount
	for _, acc := range genesisAccounts {
		genAccounts = append(genAccounts, integration.GenesisAccount{GenesisAccount: acc})
	}
	startupCfg.GenesisAccounts = genAccounts

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, accounts.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, accounts.AppModule{})
	cdc := encodingCfg.Codec

	handler := directHandler{}
	account := baseaccount.NewAccount("base", signing.NewHandlerMap(handler), baseaccount.WithSecp256K1PubKey())

	var accs []accountstd.AccountCreatorFunc
	for name, acc := range extraAccs {
		f := accountstd.AddAccount(name, func(_ accountstd.Dependencies) (accountstd.Interface, error) {
			return acc, nil
		})
		accs = append(accs, f)
	}
	res.accountsKeeper, err = accounts.NewKeeper(
		cdc,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[accounts.StoreKey]), log.NewNopLogger(), runtime.EnvWithQueryRouterService(queryRouter), runtime.EnvWithMsgRouterService(router)),
		addresscodec.NewBech32Codec("cosmos"),
		cdc.InterfaceRegistry(),
		nil,
		append(accs, account)...,
	)
	assert.NilError(t, err)

	authority := authtypes.NewModuleAddress("gov")

	res.authKeeper = authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authtypes.ProtoBaseAccount,
		res.accountsKeeper,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		res.authKeeper.GetAuthority(): false,
	}
	res.bankKeeper = bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), log.NewNopLogger()),
		cdc,
		res.authKeeper,
		blockedAddresses,
		authority.String(),
	)

	res.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.bankKeeper, &res.accountsKeeper)
	require.NoError(t, err)

	return &res
}
