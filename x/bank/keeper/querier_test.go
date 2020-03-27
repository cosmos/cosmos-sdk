package keeper_test

import (
	gocontext "context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type QueryServerTestHelper struct {
	ctx      sdk.Context
	handlers map[string]struct {
		querier interface{}
		svcDesc *grpc.ServiceDesc
	}
}

func NewQueryServerTestHelper(ctx sdk.Context) *QueryServerTestHelper {
	return &QueryServerTestHelper{ctx: ctx, handlers: map[string]struct {
		querier interface{}
		svcDesc *grpc.ServiceDesc
	}{}}
}

func (q *QueryServerTestHelper) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	q.handlers[sd.ServiceName] = struct {
		querier interface{}
		svcDesc *grpc.ServiceDesc
	}{querier: ss, svcDesc: sd}
}

func (q *QueryServerTestHelper) Invoke(ctx gocontext.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	path := strings.Split(method, "/")
	if len(path) != 3 {
		return fmt.Errorf("unexpected method name %s", method)
	}
	handler, ok := q.handlers[path[1]]
	if !ok {
		return fmt.Errorf("handler not found for %s", path[2])
	}
	for _, m := range handler.svcDesc.Methods {
		if m.MethodName == path[2] {
			req := args.(proto.Message)
			if req == nil {
				return fmt.Errorf("empty request")
			}
			reqBz, err := proto.Marshal(req)
			if err != nil {
				return err
			}
			res, err := m.Handler(handler.querier, sdk.WrapSDKContext(q.ctx), func(i interface{}) error {
				req := i.(proto.Message)
				return proto.Unmarshal(reqBz, req)
			}, nil)
			resProto := res.(proto.Message)
			resBz, err := proto.Marshal(resProto)
			resProto2 := reply.(proto.Message)
			return proto.Unmarshal(resBz, resProto2)
		}
	}
	return fmt.Errorf("method not found")
}

func (q *QueryServerTestHelper) NewStream(ctx gocontext.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("not supported")
}

var _ module.GRPCServer = &QueryServerTestHelper{}
var _ context.GRPCClientConn = &QueryServerTestHelper{}

func (suite *IntegrationTestSuite) TestQuerier_QueryBalance() {
	app, ctx := suite.app, suite.ctx
	_, _, addr := authtypes.KeyTestPubAddr()

	queryHelper := NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, keeper.Querier{app.BankKeeper})
	queryClient := types.NewQueryClient(queryHelper)

	res, err := queryClient.QueryBalance(nil, nil)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req := types.NewQueryBalanceParams(addr, fooDenom)
	balance, err := queryClient.QueryBalance(gocontext.Background(), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	suite.True(balance.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	balance, err = queryClient.QueryBalance(gocontext.Background(), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.True(balance.IsEqual(newFooCoin(50)))
}

//func (suite *IntegrationTestSuite) TestQuerier_QueryAllBalances() {
//	app, ctx := suite.app, suite.ctx
//	_, _, addr := authtypes.KeyTestPubAddr()
//	req := abci.RequestQuery{
//		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryAllBalances),
//		Data: []byte{},
//	}
//
//	querier := keeper.NewQuerier(app.BankKeeper)
//
//	res, err := querier(ctx, []string{types.QueryAllBalances}, req)
//	suite.Require().NotNil(err)
//	suite.Require().Nil(res)
//
//	req.Data = app.Codec().MustMarshalJSON(types.NewQueryAllBalancesParams(addr))
//	res, err = querier(ctx, []string{types.QueryAllBalances}, req)
//	suite.Require().NoError(err)
//	suite.Require().NotNil(res)
//
//	var balances sdk.Coins
//	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &balances))
//	suite.True(balances.IsZero())
//
//	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
//	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
//
//	app.AccountKeeper.SetAccount(ctx, acc)
//	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))
//
//	res, err = querier(ctx, []string{types.QueryAllBalances}, req)
//	suite.Require().NoError(err)
//	suite.Require().NotNil(res)
//	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &balances))
//	suite.True(balances.IsEqual(origCoins))
//}

//func (suite *IntegrationTestSuite) TestQuerierRouteNotFound() {
//	app, ctx := suite.app, suite.ctx
//	req := abci.RequestQuery{
//		Path: fmt.Sprintf("custom/%s/invalid", types.ModuleName),
//		Data: []byte{},
//	}
//
//	querier := keeper.NewQuerier(app.BankKeeper)
//	_, err := querier(ctx, []string{"invalid"}, req)
//	suite.Error(err)
//}
