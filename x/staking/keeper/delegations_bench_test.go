package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
)

func (s *KeeperTestSuite) BenchmarkDelegationsByValidator(b *testing.B) {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, valAddrs := createValAddrs(3)
	numOfDels := 1000
	delAddrs := simtestutil.CreateIncrementalAccounts(numOfDels)

	for _, addr := range addrDels {
		s.accountKeeper.EXPECT().StringToBytes(addr.String()).Return(addr, nil).AnyTimes()
		s.accountKeeper.EXPECT().BytesToString(addr).Return(addr.String(), nil).AnyTimes()
		s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), addr, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	}

	// construct the validators
	amts := []math.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]stakingtypes.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(s.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)

		validators[i] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[i], true)
	}

	_, err := s.msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(delAddrs[0], valAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2))))
	require.NoError(err)
}
