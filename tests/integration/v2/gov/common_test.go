package gov

import (
	"bytes"
	"context"
	"log"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	sdklog "cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/runtime/v2"
	_ "cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus"
	_ "cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
	_ "cosmossdk.io/x/mint"
	_ "cosmossdk.io/x/protocolpool"
	_ "cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	valTokens           = sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	TestProposal        = v1beta1.NewTextProposal("Test", "description")
	TestDescription     = stakingtypes.NewDescription("T", "E", "S", "T", "Z", &stakingtypes.Metadata{})
	TestCommissionRates = stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
)

// mkTestLegacyContent creates a MsgExecLegacyContent for testing purposes.
func mkTestLegacyContent(t *testing.T) *v1.MsgExecLegacyContent {
	t.Helper()
	msgContent, err := v1.NewLegacyContent(TestProposal, authtypes.NewModuleAddress(types.ModuleName).String())
	assert.NilError(t, err)

	return msgContent
}

var pubkeys = []cryptotypes.PubKey{
	ed25519.GenPrivKey().PubKey(),
	ed25519.GenPrivKey().PubKey(),
	ed25519.GenPrivKey().PubKey(),
}

// SortAddresses - Sorts Addresses
func SortAddresses(addrs []sdk.AccAddress) {
	byteAddrs := make([][]byte, len(addrs))

	for i, addr := range addrs {
		byteAddrs[i] = addr.Bytes()
	}

	SortByteArrays(byteAddrs)

	for i, byteAddr := range byteAddrs {
		addrs[i] = byteAddr
	}
}

// implement `Interface` in sort package.
type sortByteArrays [][]byte

func (b sortByteArrays) Len() int {
	return len(b)
}

func (b sortByteArrays) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i], b[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortByteArrays) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// SortByteArrays - sorts the provided byte array
func SortByteArrays(src [][]byte) [][]byte {
	sorted := sortByteArrays(src)
	sort.Sort(sorted)
	return sorted
}

type suite struct {
	cdc codec.Codec
	app *integration.App

	ctx context.Context

	AuthKeeper    authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	GovKeeper     *keeper.Keeper
	StakingKeeper *stakingkeeper.Keeper

	txConfigOptions tx.ConfigOptions
}

func createTestSuite(t *testing.T, genesisBehavior int) suite {
	t.Helper()
	res := suite{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.BankModule(),
		configurator.GovModule(),
		configurator.MintModule(),
		configurator.ConsensusModule(),
		configurator.ProtocolPoolModule(),
	}

	startupCfg := integration.DefaultStartUpConfig(t)

	msgRouterService := integration.NewRouterService()
	res.registerMsgRouterService(msgRouterService)

	var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
		return msgRouterService
	}

	queryRouterService := integration.NewRouterService()
	res.registerQueryRouterService(queryRouterService)
	serviceBuilder := runtime.NewRouterBuilder(routerFactory, queryRouterService)

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = &integration.HeaderService{}
	startupCfg.GasService = &integration.GasService{}
	startupCfg.GenesisBehavior = genesisBehavior

	app, err := integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(sdklog.NewNopLogger())),
		startupCfg,
		&res.AuthKeeper, &res.BankKeeper, &res.GovKeeper, &res.StakingKeeper, &res.cdc, &res.txConfigOptions,
	)
	require.NoError(t, err)

	res.ctx = app.StateLatestContext(t)
	res.app = app
	return res
}

func (s *suite) registerMsgRouterService(router *integration.RouterService) {
	// register custom router service
	bankSendHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*banktypes.MsgSend)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := bankkeeper.NewMsgServerImpl(s.BankKeeper)
		resp, err := msgServer.Send(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(bankSendHandler, "/cosmos.bank.v1beta1.MsgSend")

	// register custom router service

	govSubmitProposalHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*v1.MsgExecLegacyContent)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := keeper.NewMsgServerImpl(s.GovKeeper)
		resp, err := msgServer.ExecLegacyContent(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(govSubmitProposalHandler, "/cosmos.gov.v1.MsgExecLegacyContent")
}

func (f *suite) registerQueryRouterService(router *integration.RouterService) {
}
