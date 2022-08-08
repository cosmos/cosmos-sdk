package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type baseSuite struct {
	t gocuke.TestingT

	alice         sdk.AccAddress
	authKeeper    *govtestutil.MockAccountKeeper
	bankKeeper    *govtestutil.MockBankKeeper
	govKeeper     *keeper.Keeper
	stakingKeeper *govtestutil.MockStakingKeeper
	ctx           sdk.Context
}

func setupBase(t gocuke.TestingT) *baseSuite {
	s := &baseSuite{t: t}

	key := sdk.NewKVStoreKey(types.StoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	_, _, s.alice = testdata.KeyTestPubAddr()
	_, _, moduleAddress := testdata.KeyTestPubAddr()

	s.ctx = testutil.DefaultContext(key, sdk.NewTransientStoreKey("transient_test"))

	// gomock initializations
	ctrl := gomock.NewController(t)
	s.authKeeper = govtestutil.NewMockAccountKeeper(ctrl)
	s.bankKeeper = govtestutil.NewMockBankKeeper(ctrl)
	s.stakingKeeper = govtestutil.NewMockStakingKeeper(ctrl)
	s.authKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(moduleAddress)

	s.govKeeper = keeper.NewKeeper(encCfg.Codec, key, s.authKeeper, s.bankKeeper, s.stakingKeeper, nil, types.DefaultConfig(), moduleAddress.String())
	s.govKeeper.SetProposalID(s.ctx, 0)

	return s
}

// this is an example of how we will unit test the gov functionality with mocks
func TestKeeperExample(t *testing.T) {
	t.Parallel()
	s := setupBase(t)
	require.NotNil(t, s.govKeeper)
}
