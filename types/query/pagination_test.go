package query_test

import (
	gocontext "context"
	"fmt"
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/params"
)

const (
	holder          = "holder"
	multiPerm       = "multiple permissions account"
	randomPerm      = "random permission"
	numBalances     = 235
	defaultLimit    = 100
	overLimit       = 101
	underLimit      = 10
	lastPageRecords = 35
)

type paginationTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	bankKeeper    bankkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper
	cdc           codec.Codec
	interfaceReg  codectypes.InterfaceRegistry
	app           *runtime.App
}

func TestPaginationTestSuite(t *testing.T) {
	suite.Run(t, new(paginationTestSuite))
}

func (s *paginationTestSuite) SetupTest() {
	var (
		bankKeeper    bankkeeper.Keeper
		accountKeeper authkeeper.AccountKeeper
		reg           codectypes.InterfaceRegistry
		cdc           codec.Codec
	)

	app, err := testutilsims.Setup(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.BankModule(),
				configurator.ParamsModule(),
				configurator.ConsensusModule(),
				configurator.OmitInitGenesis(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&bankKeeper, &accountKeeper, &reg, &cdc)

	s.NoError(err)

	ctx := app.NewContextLegacy(false, cmtproto.Header{Height: 1})

	s.ctx, s.bankKeeper, s.accountKeeper, s.cdc, s.app, s.interfaceReg = ctx, bankKeeper, accountKeeper, cdc, app, reg
}

func (s *paginationTestSuite) TestParsePagination() {
	s.T().Log("verify default values for empty page request")
	pageReq := &query.PageRequest{}
	page, limit, err := query.ParsePagination(pageReq)
	s.Require().NoError(err)
	s.Require().Equal(limit, query.DefaultLimit)
	s.Require().Equal(page, 1)

	s.T().Log("verify with custom values")
	pageReq = &query.PageRequest{
		Offset: 0,
		Limit:  10,
	}
	page, limit, err = query.ParsePagination(pageReq)
	s.Require().NoError(err)
	s.Require().Equal(page, 1)
	s.Require().Equal(limit, 10)
}

func (s *paginationTestSuite) TestPagination() {
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.interfaceReg)
	types.RegisterQueryServer(queryHelper, s.bankKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	var balances sdk.Coins

	for i := range numBalances {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	balances = balances.Sort()
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := s.accountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.accountKeeper.SetAccount(s.ctx, acc1)
	s.Require().NoError(testutil.FundAccount(s.ctx, s.bankKeeper, addr1, balances))

	s.T().Log("verify empty page request results a max of defaultLimit records and counts total records")
	pageReq := &query.PageRequest{}
	request := types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err := queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Pagination.Total, uint64(numBalances))
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().LessOrEqual(res.Balances.Len(), defaultLimit)

	s.T().Log("verify page request with limit > defaultLimit, returns less or equal to `limit` records")
	pageReq = &query.PageRequest{Limit: overLimit}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Pagination.Total, uint64(0))
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().LessOrEqual(res.Balances.Len(), overLimit)

	s.T().Log("verify paginate with custom limit and countTotal true")
	pageReq = &query.PageRequest{Limit: underLimit, CountTotal: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), underLimit)
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(numBalances))

	s.T().Log("verify paginate with custom limit and countTotal false")
	pageReq = &query.PageRequest{Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), defaultLimit)
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify paginate with custom limit, key and countTotal false")
	pageReq = &query.PageRequest{Key: res.Pagination.NextKey, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), defaultLimit)
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify paginate for last page, results in records less than max limit")
	pageReq = &query.PageRequest{Key: res.Pagination.NextKey, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().LessOrEqual(res.Balances.Len(), defaultLimit)
	s.Require().Equal(res.Balances.Len(), lastPageRecords)
	s.Require().Nil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify paginate with offset and limit")
	pageReq = &query.PageRequest{Offset: 200, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().LessOrEqual(res.Balances.Len(), defaultLimit)
	s.Require().Equal(res.Balances.Len(), lastPageRecords)
	s.Require().Nil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify paginate with offset and limit")
	pageReq = &query.PageRequest{Offset: 100, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().LessOrEqual(res.Balances.Len(), defaultLimit)
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify paginate with offset and key - error")
	pageReq = &query.PageRequest{Key: res.Pagination.NextKey, Offset: 100, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	_, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = InvalidArgument desc = paginate: invalid request, either offset or key is expected, got both", err.Error())

	s.T().Log("verify paginate with offset greater than total results")
	pageReq = &query.PageRequest{Offset: 300, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().LessOrEqual(res.Balances.Len(), 0)
	s.Require().Nil(res.Pagination.NextKey)
}

func (s *paginationTestSuite) TestReversePagination() {
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.interfaceReg)
	types.RegisterQueryServer(queryHelper, s.bankKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	var balances sdk.Coins

	for i := range numBalances {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	balances = balances.Sort()
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := s.accountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.accountKeeper.SetAccount(s.ctx, acc1)
	s.Require().NoError(testutil.FundAccount(s.ctx, s.bankKeeper, addr1, balances))

	s.T().Log("verify paginate with custom limit and countTotal, Reverse false")
	pageReq := &query.PageRequest{Limit: 2, CountTotal: true, Reverse: true, Key: nil}
	request := types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res1, err := queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(2, res1.Balances.Len())
	s.Require().NotNil(res1.Pagination.NextKey)

	s.T().Log("verify paginate with custom limit and countTotal, Reverse false")
	pageReq = &query.PageRequest{Limit: 150}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res1, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res1.Balances.Len(), 150)
	s.Require().NotNil(res1.Pagination.NextKey)
	s.Require().Equal(res1.Pagination.Total, uint64(0))

	s.T().Log("verify paginate with custom limit, key and Reverse true")
	pageReq = &query.PageRequest{Limit: defaultLimit, Reverse: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err := queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), defaultLimit)
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify paginate with custom limit, key and Reverse true")
	pageReq = &query.PageRequest{Offset: 100, Limit: defaultLimit, Reverse: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), defaultLimit)
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify paginate for last page, Reverse true")
	pageReq = &query.PageRequest{Offset: 200, Limit: defaultLimit, Reverse: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), lastPageRecords)
	s.Require().Nil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify page request with limit > defaultLimit, returns less or equal to `limit` records")
	pageReq = &query.PageRequest{Limit: overLimit, Reverse: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Pagination.Total, uint64(0))
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().LessOrEqual(res.Balances.Len(), overLimit)

	s.T().Log("verify paginate with custom limit, key, countTotal false and Reverse true")
	pageReq = &query.PageRequest{Key: res1.Pagination.NextKey, Limit: 50, Reverse: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), 50)
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify Reverse pagination returns valid result")
	s.Require().Equal(balances[101:151].String(), res.Balances.Sort().String())

	s.T().Log("verify paginate with custom limit, key, countTotal false and Reverse true")
	pageReq = &query.PageRequest{Key: res.Pagination.NextKey, Limit: 50, Reverse: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), 50)
	s.Require().NotNil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify Reverse pagination returns valid result")
	s.Require().Equal(balances[51:101].String(), res.Balances.Sort().String())

	s.T().Log("verify paginate for last page Reverse true")
	pageReq = &query.PageRequest{Key: res.Pagination.NextKey, Limit: defaultLimit, Reverse: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().Equal(res.Balances.Len(), 51)
	s.Require().Nil(res.Pagination.NextKey)
	s.Require().Equal(res.Pagination.Total, uint64(0))

	s.T().Log("verify Reverse pagination returns valid result")
	s.Require().Equal(balances[0:51].String(), res.Balances.Sort().String())

	s.T().Log("verify paginate with offset and key - error")
	pageReq = &query.PageRequest{Key: res1.Pagination.NextKey, Offset: 100, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	_, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().Error(err)
	s.Require().Equal("rpc error: code = InvalidArgument desc = paginate: invalid request, either offset or key is expected, got both", err.Error())

	s.T().Log("verify paginate with offset greater than total results")
	pageReq = &query.PageRequest{Offset: 300, Limit: defaultLimit, CountTotal: false, Reverse: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	s.Require().NoError(err)
	s.Require().LessOrEqual(res.Balances.Len(), 0)
	s.Require().Nil(res.Pagination.NextKey)
}

func (s *paginationTestSuite) TestPaginate() {
	var balances sdk.Coins

	for i := range 2 {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	balances = balances.Sort()
	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := s.accountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.accountKeeper.SetAccount(s.ctx, acc1)
	err := testutil.FundAccount(s.ctx, s.bankKeeper, addr1, balances)
	if err != nil { // should return no error
		fmt.Println(err)
	}
	// Paginate example
	pageReq := &query.PageRequest{Key: nil, Limit: 1, CountTotal: true}
	request := types.NewQueryAllBalancesRequest(addr1, pageReq, false)
	balResult := sdk.NewCoins()
	authStore := s.ctx.KVStore(s.app.UnsafeFindStoreKey(types.StoreKey))
	balancesStore := prefix.NewStore(authStore, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, address.MustLengthPrefix(addr1))
	pageRes, err := query.Paginate(accountStore, request.Pagination, func(key, value []byte) error {
		var amount math.Int
		err := amount.Unmarshal(value)
		if err != nil {
			return err
		}
		balResult = append(balResult, sdk.NewCoin(string(key), amount))
		return nil
	})
	if err != nil { // should return no error
		fmt.Println(err)
	}
	fmt.Println(&types.QueryAllBalancesResponse{Balances: balResult, Pagination: pageRes})
	// Output:
	// balances:<denom:"foo0denom" amount:"100" > pagination:<next_key:"foo1denom" total:2 >
}
