package keeper_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/accounts"
	baseaccount "cosmossdk.io/x/accounts/defaults/base"
	"cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	authsims "cosmossdk.io/x/auth/simulation"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"
	minttypes "cosmossdk.io/x/mint/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type fixture struct {
	app *integration.App

	cdc codec.Codec
	ctx sdk.Context

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.Keeper
}

var _ signing.SignModeHandler = directHandler{}

type directHandler struct{}

func (s directHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT
}

func (s directHandler) GetSignBytes(_ context.Context, _ signing.SignerData, _ signing.TxData) ([]byte, error) {
	panic("not implemented")
}

func initFixture(t *testing.T) *fixture {
	t.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, accounts.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, accounts.AppModule{})
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, true, logger)

	router := baseapp.NewMsgServiceRouter()
	queryRouter := baseapp.NewGRPCQueryRouter()

	handler := directHandler{}
	account := baseaccount.NewAccount("base", signing.NewHandlerMap(handler), baseaccount.WithSecp256K1PubKey())
	accountsKeeper, err := accounts.NewKeeper(
		cdc,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[accounts.StoreKey]), log.NewNopLogger(), runtime.EnvWithQueryRouterService(queryRouter), runtime.EnvWithMsgRouterService(router)),
		addresscodec.NewBech32Codec("cosmos"),
		cdc.InterfaceRegistry(),
		account,
	)
	assert.NilError(t, err)

	authority := authtypes.NewModuleAddress("gov")

	authKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authtypes.ProtoBaseAccount,
		accountsKeeper,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		authKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authKeeper,
		blockedAddresses,
		authority.String(),
	)

	params := banktypes.DefaultParams()
	assert.NilError(t, bankKeeper.SetParams(newCtx, params))

	accountsModule := accounts.NewAppModule(cdc, accountsKeeper)
	authModule := auth.NewAppModule(cdc, authKeeper, accountsKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, authKeeper)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc,
		encodingCfg.InterfaceRegistry.SigningContext().AddressCodec(),
		encodingCfg.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			accounts.ModuleName:  accountsModule,
			authtypes.ModuleName: authModule,
			banktypes.ModuleName: bankModule,
		}, router, queryRouter)

	authtypes.RegisterInterfaces(cdc.InterfaceRegistry())
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry())

	authtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), authkeeper.NewMsgServerImpl(authKeeper))
	authtypes.RegisterQueryServer(integrationApp.QueryHelper(), authkeeper.NewQueryServer(authKeeper))

	banktypes.RegisterMsgServer(router, bankkeeper.NewMsgServerImpl(bankKeeper))

	return &fixture{
		app:            integrationApp,
		cdc:            cdc,
		ctx:            newCtx,
		accountsKeeper: accountsKeeper,
		authKeeper:     authKeeper,
		bankKeeper:     bankKeeper,
	}
}

func TestAsyncExec(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	addrs := simtestutil.CreateIncrementalAccounts(2)
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10)))

	assert.NilError(t, testutil.FundAccount(f.ctx, f.bankKeeper, addrs[0], sdk.NewCoins(sdk.NewInt64Coin("stake", 500))))

	msg := &banktypes.MsgSend{
		FromAddress: addrs[0].String(),
		ToAddress:   addrs[1].String(),
		Amount:      coins,
	}
	msg2 := &banktypes.MsgSend{
		FromAddress: addrs[1].String(),
		ToAddress:   addrs[0].String(),
		Amount:      coins,
	}
	failingMsg := &banktypes.MsgSend{
		FromAddress: addrs[0].String(),
		ToAddress:   addrs[1].String(),
		Amount:      sdk.NewCoins(sdk.NewCoin("stake", sdkmath.ZeroInt())), // No amount specified
	}

	msgAny, err := codectypes.NewAnyWithValue(msg)
	assert.NilError(t, err)

	msgAny2, err := codectypes.NewAnyWithValue(msg2)
	assert.NilError(t, err)

	failingMsgAny, err := codectypes.NewAnyWithValue(failingMsg)
	assert.NilError(t, err)

	testCases := []struct {
		name      string
		req       *authtypes.MsgNonAtomicExec
		expectErr bool
		expErrMsg string
	}{
		{
			name: "empty signer address",
			req: &authtypes.MsgNonAtomicExec{
				Signer: "",
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "empty signer address string is not allowed",
		},
		{
			name: "invalid signer address",
			req: &authtypes.MsgNonAtomicExec{
				Signer: "invalid",
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "invalid signer address",
		},
		{
			name: "empty msgs",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "messages cannot be empty",
		},
		{
			name: "valid msg",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny},
			},
			expectErr: false,
		},
		{
			name: "multiple messages being executed",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny, msgAny},
			},
			expectErr: false,
		},
		{
			name: "multiple messages with different signers",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny, msgAny2},
			},
			expectErr: false,
			expErrMsg: "unauthorized: sender does not match expected sender",
		},
		{
			name: "multi msg with one failing being executed",
			req: &authtypes.MsgNonAtomicExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny, failingMsgAny},
			},
			expectErr: false,
			expErrMsg: "invalid coins",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.req,
				integration.WithAutomaticFinalizeBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expectErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := authtypes.MsgNonAtomicExecResponse{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				if tc.expErrMsg != "" {
					for _, res := range result.Results {
						if res.Error != "" {
							assert.Assert(t, strings.Contains(res.Error, tc.expErrMsg), fmt.Sprintf("res.Error %s does not contain %s", res.Error, tc.expErrMsg))
						}
						continue
					}
				}
			}
		})
	}
}
