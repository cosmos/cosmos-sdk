package migrations

import (
	"fmt"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

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
	keeper      keeper.AccountKeeper
	queryServer grpc.Server
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper keeper.AccountKeeper, queryServer grpc.Server) Migrator {
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

		delegations := getDelegatorDelegationsSum(
			ctx,
			addr,
			m.queryServer,
		)

		asVesting, ok := resetVestingDelegatedBalances(asVesting)
		if !ok {
			return false
		}

		// balance before any delegation includes balance of delegation
		for _, coin := range delegations {
			balance = balance.Add(coin)
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

func resetVestingDelegatedBalances(evacct exported.VestingAccount) (exported.VestingAccount, bool) {
	// reset `DelegatedVesting` and `DelegatedFree` to zero
	df := sdk.NewCoins()
	dv := sdk.NewCoins()

	switch vacct := evacct.(type) {
	case *vestingtypes.ContinuousVestingAccount:
		vacct.DelegatedVesting = dv
		vacct.DelegatedFree = df
		return vacct, true
	case *vestingtypes.DelayedVestingAccount:
		vacct.DelegatedVesting = dv
		vacct.DelegatedFree = df
		return vacct, true
	case *vestingtypes.PeriodicVestingAccount:
		vacct.DelegatedVesting = dv
		vacct.DelegatedFree = df
		return vacct, true
	default:
		return nil, false
	}
}

func getDelegatorDelegationsSum(ctx sdk.Context, address string, queryServer grpc.Server) sdk.Coins {
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
	if err := proto.Unmarshal(resp.Value, balance); err != nil {
		panic(fmt.Errorf("unable to unmarshal delegator query delegations: %w", err))
	}

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
	if err := proto.Unmarshal(resp.Value, balance); err != nil {
		panic(fmt.Errorf("unable to unmarshal bank balance response: %w", err))
	}
	return balance.Balances
}
