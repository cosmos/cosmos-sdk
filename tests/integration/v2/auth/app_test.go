package auth

import (
	"context"
	"io"
	"testing"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/testing/msgrouter"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/x/accounts"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus" // import as blank for app wiring
	"cosmossdk.io/x/gov"
	_ "cosmossdk.io/x/staking" // import as blank for app wirings

	"github.com/cosmos/cosmos-sdk/client"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
)

type suite struct {
	app *integration.App
	ctx context.Context

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.Keeper

	clientCtx client.Context
	val, val1 sdk.AccAddress
}

func (s suite) mustAddr(address []byte) string {
	str, _ := s.authKeeper.AddressCodec().BytesToString(address)
	return str
}

func createTestSuite(t *testing.T) *suite {
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

	msgRouterService := msgrouter.NewRouterService()
	res.registerMsgRouterService(msgRouterService)

	var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
		return msgRouterService
	}

	queryRouterService := msgrouter.NewRouterService()
	res.registerQueryRouterService(queryRouterService)

	serviceBuilder := runtime.NewRouterBuilder(routerFactory, queryRouterService)

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = services.NewGenesisHeaderService(stf.HeaderService{})

	res.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Provide(
			// inject desired account types:
			basedepinject.ProvideAccount,

			// provide base account options
			basedepinject.ProvideSecp256K1PubKey,

			// provide extra accounts
			ProvideMockRetroCompatAccountValid,
			ProvideMockRetroCompatAccountNoInfo,
			ProvideMockRetroCompatAccountNoImplement,
		), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.bankKeeper, &res.accountsKeeper, &res.authKeeper)
	require.NoError(t, err)

	res.ctx = res.app.StateLatestContext(t)

	// setup client context
	encCfg := testutilmod.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, gov.AppModule{})
	clientCtx := client.Context{}.
		WithKeyring(keyring.NewInMemory(encCfg.Codec)).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain").
		WithAddressCodec(addresscodec.NewBech32Codec("cosmos")).
		WithValidatorAddressCodec(addresscodec.NewBech32Codec("cosmosvaloper")).
		WithConsensusAddressCodec(addresscodec.NewBech32Codec("cosmosvalcons"))

	ctxGen := func() client.Context {
		bz, _ := encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := clitestutil.NewMockCometRPC(abci.QueryResponse{
			Value: bz,
		})
		return clientCtx.WithClient(c)
	}
	res.clientCtx = ctxGen()

	// setup keyring
	valAcc, _, err := res.clientCtx.Keyring.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	account1, _, err := res.clientCtx.Keyring.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	account2, _, err := res.clientCtx.Keyring.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	_, _, err = res.clientCtx.Keyring.NewMnemonic("dummyAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1) // Create a dummy account for testing purpose
	require.NoError(t, err)

	res.val, err = valAcc.GetAddress()
	require.NoError(t, err)
	require.NoError(t, err)
	res.val1, err = account1.GetAddress()
	require.NoError(t, err)
	pub1, err := account1.GetPubKey()
	require.NoError(t, err)
	pub2, err := account2.GetPubKey()
	require.NoError(t, err)

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{pub1, pub2})
	_, err = res.clientCtx.Keyring.SaveMultisig("multi", multi)
	require.NoError(t, err)

	return &res
}

func (s *suite) registerMsgRouterService(router *msgrouter.RouterService) {
	// register custom router service
	bankSendHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*banktypes.MsgSend)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := bankkeeper.NewMsgServerImpl(s.bankKeeper)
		resp, err := msgServer.Send(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(bankSendHandler, "cosmos.bank.v1beta1.MsgSend")
}

func (s *suite) registerQueryRouterService(router *msgrouter.RouterService) {
	// register custom router service
	queryHandler := func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
		req, ok := msg.(*accountsv1.AccountNumberRequest)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		qs := accounts.NewQueryServer(s.accountsKeeper)
		resp, err := qs.AccountNumber(ctx, req)
		return resp, err
	}

	router.RegisterHandler(queryHandler, "cosmos.accounts.v1.AccountNumberRequest")
}
