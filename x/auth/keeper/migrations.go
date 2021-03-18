package keeper

import (
	"fmt"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	"github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	delegatorDelegationPath = "/cosmos.staking.v1beta1.Query/DelegatorDelegations"
	balancesPath            = "/cosmos.bank.v1beta1.Query/AllBalances"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper      AccountKeeper
	queryServer grpc.Server
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper AccountKeeper, queryServer grpc.Server) Migrator {
	return Migrator{keeper: keeper, queryServer: queryServer}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	m.keeper.IterateAccounts(ctx, func(account types.AccountI) (stop bool) {
		asVesting := vesting(account)
		if asVesting == nil {
			return false
		}

		addr := account.GetAddress().String()
		balance := getBalance(
			ctx,
			addr,
			m.queryServer,
		)

		delegations := getDelegatorDelegations(
			ctx,
			addr,
			m.queryServer,
		)

		if delegations.IsAllGTE(asVesting.GetOriginalVesting()) {
			delegations = asVesting.GetOriginalVesting()
		}

		if balance.IsAllLT(delegations) {
			balance = balance.Add(delegations...)
		}

		asVesting.TrackDelegation(ctx.BlockTime(), balance, delegations)

		m.keeper.SetAccount(ctx, account)

		return false
	})

	return nil
}

func vesting(account types.AccountI) exported.VestingAccount {
	v, ok := account.(exported.VestingAccount)
	if !ok {
		return nil
	}

	return v
}

func getDelegatorDelegations(ctx sdk.Context, address string, queryServer grpc.Server) sdk.Coins {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		panic(fmt.Sprintf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer))
	}

	queryFn := querier.Route(delegatorDelegationPath)

	q := &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: address,
	}

	b, err := proto.Marshal(q)
	if err != nil {
		panic(err)
	}
	req := abci.RequestQuery{
		Data: b,
		Path: delegatorDelegationPath,
	}
	resp, err := queryFn(ctx, req)
	if err != nil {
		panic(err)
	}
	balance := new(stakingtypes.QueryDelegatorDelegationsResponse)
	err = proto.Unmarshal(resp.Value, balance)

	res := sdk.NewCoins()
	for _, i := range balance.DelegationResponses {
		res = res.Add(i.Balance)
	}

	return res
}

func getBalance(ctx sdk.Context, address string, queryServer grpc.Server) sdk.Coins {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		panic(fmt.Sprintf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer))
	}

	queryFn := querier.Route(balancesPath)

	q := &banktypes.QueryAllBalancesRequest{
		Address:    address,
		Pagination: nil,
	}
	b, err := proto.Marshal(q)
	if err != nil {
		panic(err)
	}
	req := abci.RequestQuery{
		Data: b,
		Path: balancesPath,
	}
	resp, err := queryFn(ctx, req)
	if err != nil {
		panic(err)
	}
	balance := new(banktypes.QueryAllBalancesResponse)
	err = proto.Unmarshal(resp.Value, balance)
	return balance.Balances
}
