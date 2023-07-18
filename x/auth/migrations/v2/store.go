// Package v2 creates in-place store migrations for fixing tracking
// delegations with vesting accounts.
// ref: https://github.com/cosmos/cosmos-sdk/issues/8601
// ref: https://github.com/cosmos/cosmos-sdk/issues/8812
//
// The migration script modifies x/auth state, hence lives in the `x/auth/legacy`
// folder. However, it needs access to staking and bank state. To avoid
// cyclic dependencies, we cannot import those 2 keepers in this file. To solve
// this, we use the baseapp router to do inter-module querying, by importing
// the `baseapp.QueryRouter grpc.Server`. This is really hacky.
//
// PLEASE DO NOT REPLICATE THIS PATTERN IN YOUR OWN APP.
//
// Proposals to refactor this file have been made in:
// https://github.com/cosmos/cosmos-sdk/issues/9070
// The preferred solution is to use inter-module communication (ADR-033), and
// this file will be refactored to use ADR-033 once it's ready.
package v2

import (
	"errors"
	"fmt"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	stakingv1beta1 "cosmossdk.io/api/cosmos/staking/v1beta1"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

const (
	delegatorDelegationPath           = "/cosmos.staking.v1beta1.Query/DelegatorDelegations"
	stakingParamsPath                 = "/cosmos.staking.v1beta1.Query/Params"
	delegatorUnbondingDelegationsPath = "/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations"
	balancesPath                      = "/cosmos.bank.v1beta1.Query/AllBalances"
)

// We use the baseapp.QueryRouter here to do inter-module state querying.
// PLEASE DO NOT REPLICATE THIS PATTERN IN YOUR OWN APP.
func migrateVestingAccounts(ctx sdk.Context, account sdk.AccountI, queryServer grpc.Server) (sdk.AccountI, error) {
	bondDenom, err := getBondDenom(ctx, queryServer)
	if err != nil {
		return nil, err
	}

	asVesting, ok := account.(exported.VestingAccount)
	if !ok {
		return nil, nil
	}

	addr := account.GetAddress().String()
	balance, err := getBalance(
		ctx,
		addr,
		queryServer,
	)
	if err != nil {
		return nil, err
	}

	delegations, err := getDelegatorDelegationsSum(
		ctx,
		addr,
		queryServer,
	)
	if err != nil {
		return nil, err
	}

	unbondingDelegations, err := getDelegatorUnbondingDelegationsSum(
		ctx,
		addr,
		bondDenom,
		queryServer,
	)
	if err != nil {
		return nil, err
	}

	delegations = delegations.Add(unbondingDelegations...)

	asVesting, ok = resetVestingDelegatedBalances(asVesting)
	if !ok {
		return nil, nil
	}

	// balance before any delegation includes balance of delegation
	for _, coin := range delegations {
		balance = balance.Add(coin)
	}

	asVesting.TrackDelegation(ctx.BlockTime(), balance, delegations)

	return asVesting.(sdk.AccountI), nil
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

// We use the baseapp.QueryRouter here to do inter-module state querying.
// PLEASE DO NOT REPLICATE THIS PATTERN IN YOUR OWN APP.
func getDelegatorDelegationsSum(ctx sdk.Context, address string, queryServer grpc.Server) (sdk.Coins, error) {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer)
	}

	queryFn := querier.Route(delegatorDelegationPath)

	q := &stakingv1beta1.QueryDelegatorDelegationsRequest{
		DelegatorAddr: address,
	}

	b, err := proto.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal staking type query request, %w", err)
	}
	req := abci.RequestQuery{
		Data: b,
		Path: delegatorDelegationPath,
	}
	resp, err := queryFn(ctx, &req)
	if err != nil {
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("staking query error, %w", err)
	}

	balance := new(stakingv1beta1.QueryDelegatorDelegationsResponse)
	if err := proto.Unmarshal(resp.Value, balance); err != nil {
		return nil, fmt.Errorf("unable to unmarshal delegator query delegations: %w", err)
	}

	res := sdk.NewCoins()
	for _, i := range balance.DelegationResponses {
		bal, err := strconv.Atoi(i.Balance.Amount)
		if err != nil {
			return nil, fmt.Errorf("cannot convert balance amount to int, %w", err)
		}
		coin := sdk.NewCoin(i.Balance.Denom, math.NewInt(int64(bal)))
		res = res.Add(coin)
	}

	return res, nil
}

// We use the baseapp.QueryRouter here to do inter-module state querying.
// PLEASE DO NOT REPLICATE THIS PATTERN IN YOUR OWN APP.
func getDelegatorUnbondingDelegationsSum(ctx sdk.Context, address, bondDenom string, queryServer grpc.Server) (sdk.Coins, error) {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer)
	}

	queryFn := querier.Route(delegatorUnbondingDelegationsPath)

	q := &stakingv1beta1.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: address,
	}

	b, err := proto.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal staking type query request, %w", err)
	}
	req := abci.RequestQuery{
		Data: b,
		Path: delegatorUnbondingDelegationsPath,
	}
	resp, err := queryFn(ctx, &req)
	if err != nil && !errors.Is(err, sdkerrors.ErrNotFound) {
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("staking query error, %w", err)
	}

	balance := new(stakingv1beta1.QueryDelegatorUnbondingDelegationsResponse)
	if err := proto.Unmarshal(resp.Value, balance); err != nil {
		return nil, fmt.Errorf("unable to unmarshal delegator query delegations: %w", err)
	}

	res := sdk.NewCoins()
	for _, i := range balance.UnbondingResponses {
		for _, r := range i.Entries {
			bal, err := strconv.Atoi(r.Balance)
			if err != nil {
				return nil, fmt.Errorf("unable to convert unbonding balance to int: %w", err)
			}
			res = res.Add(sdk.NewCoin(bondDenom, math.NewInt(int64(bal))))
		}
	}

	return res, nil
}

// We use the baseapp.QueryRouter here to do inter-module state querying.
// PLEASE DO NOT REPLICATE THIS PATTERN IN YOUR OWN APP.
func getBalance(ctx sdk.Context, address string, queryServer grpc.Server) (sdk.Coins, error) {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer)
	}

	queryFn := querier.Route(balancesPath)

	q := &bankv1beta1.QueryAllBalancesRequest{
		Address:    address,
		Pagination: nil,
	}
	b, err := proto.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal bank type query request, %w", err)
	}

	req := abci.RequestQuery{
		Data: b,
		Path: balancesPath,
	}
	resp, err := queryFn(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("bank query error, %w", err)
	}
	balance := new(bankv1beta1.QueryAllBalancesResponse)
	if err := proto.Unmarshal(resp.Value, balance); err != nil {
		return nil, fmt.Errorf("unable to unmarshal bank balance response: %w", err)
	}
	coins := make(sdk.Coins, len(balance.Balances))
	for i, b := range balance.Balances {
		amount, err := strconv.Atoi(b.Amount)
		if err != nil {
			return nil, fmt.Errorf("cannot convert balance amount to int, %w", err)
		}
		coins[i] = sdk.NewCoin(b.Denom, math.NewInt(int64(amount)))
	}
	return coins, nil
}

// We use the baseapp.QueryRouter here to do inter-module state querying.
// PLEASE DO NOT REPLICATE THIS PATTERN IN YOUR OWN APP.
func getBondDenom(ctx sdk.Context, queryServer grpc.Server) (string, error) {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		return "", fmt.Errorf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer)
	}

	queryFn := querier.Route(stakingParamsPath)

	q := &stakingv1beta1.QueryParamsRequest{}

	b, err := proto.Marshal(q)
	if err != nil {
		return "", fmt.Errorf("cannot marshal staking params query request, %w", err)
	}
	req := abci.RequestQuery{
		Data: b,
		Path: stakingParamsPath,
	}

	resp, err := queryFn(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("staking query error, %w", err)
	}

	params := new(stakingv1beta1.QueryParamsResponse)
	if err := proto.Unmarshal(resp.Value, params); err != nil {
		return "", fmt.Errorf("unable to unmarshal delegator query delegations: %w", err)
	}

	return params.Params.BondDenom, nil
}

// MigrateAccount migrates vesting account to make the DelegatedVesting and DelegatedFree fields correctly
// track delegations.
// References: https://github.com/cosmos/cosmos-sdk/issues/8601, https://github.com/cosmos/cosmos-sdk/issues/8812
//
// We use the baseapp.QueryRouter here to do inter-module state querying.
// PLEASE DO NOT REPLICATE THIS PATTERN IN YOUR OWN APP.
func MigrateAccount(ctx sdk.Context, account sdk.AccountI, queryServer grpc.Server) (sdk.AccountI, error) {
	return migrateVestingAccounts(ctx, account, queryServer)
}
