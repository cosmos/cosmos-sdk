package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type baseSuite struct {
	t gocuke.TestingT

	govAccount sdk.AccAddress
	alice      sdk.AccAddress

	cdc           codec.Codec
	authKeeper    *govtestutil.MockAccountKeeper
	bankKeeper    *govtestutil.MockBankKeeper
	govKeeper     *keeper.Keeper
	stakingKeeper *govtestutil.MockStakingKeeper
	ctx           sdk.Context
}

func setupBase(t gocuke.TestingT) *baseSuite {
	s := &baseSuite{t: t}

	key := sdk.NewKVStoreKey(types.StoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig(bank.AppModuleBasic{})
	// We register here the "/testdata.MsgCreateDog" message int he interface
	// registry, to be able to unmarshal it. But we do NOT register the
	// "testdata.Msg/CreateDog" endpoint in the MsgServiceRouter, because we
	// want it to fail.
	testdata.RegisterInterfaces(encCfg.InterfaceRegistry)
	s.cdc = encCfg.Codec
	msr := baseapp.NewMsgServiceRouter()
	msr.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	banktypes.RegisterMsgServer(msr, nil) // Nil is fine here as long as we never execute the proposal's Msgs.

	_, _, s.alice = testdata.KeyTestPubAddr()
	s.govAccount = authtypes.NewModuleAddress(types.ModuleName)

	s.ctx = testutil.DefaultContext(key, sdk.NewTransientStoreKey("transient_test"))

	// Gomock initializations
	ctrl := gomock.NewController(t)
	s.authKeeper = govtestutil.NewMockAccountKeeper(ctrl)
	s.bankKeeper = govtestutil.NewMockBankKeeper(ctrl)
	s.stakingKeeper = govtestutil.NewMockStakingKeeper(ctrl)
	s.authKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(s.govAccount)

	// Initialize the gov keeper
	s.govKeeper = keeper.NewKeeper(encCfg.Codec, key, s.authKeeper, s.bankKeeper, s.stakingKeeper, msr, types.DefaultConfig(), s.govAccount.String())
	s.govKeeper.SetProposalID(s.ctx, 1)
	s.govKeeper.SetParams(s.ctx, v1.DefaultParams())

	return s
}

// this is an example of how we will unit test the gov functionality with mocks
func TestKeeperExample(t *testing.T) {
	t.Parallel()
	s := setupBase(t)
	require.NotNil(t, s.govKeeper)
}
