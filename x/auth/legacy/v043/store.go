package v043

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	delegatorDelegationPath           = "/cosmos.staking.v1beta1.Query/DelegatorDelegations"
	stakingParamsPath                 = "/cosmos.staking.v1beta1.Query/Params"
	delegatorUnbondingDelegationsPath = "/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations"
	balancesPath                      = "/cosmos.bank.v1beta1.Query/AllBalances"
)

func migrateVestingAccounts(ctx sdk.Context, account types.AccountI, queryServer grpc.Server) (types.AccountI, error) {
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

	return asVesting.(types.AccountI), nil
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

func getDelegatorDelegationsSum(ctx sdk.Context, address string, queryServer grpc.Server) (sdk.Coins, error) {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer)
	}

	queryFn := querier.Route(delegatorDelegationPath)

	q := &stakingtypes.QueryDelegatorDelegationsRequest{
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
	resp, err := queryFn(ctx, req)
	if err != nil {
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("staking query error, %w", err)
	}

	balance := new(stakingtypes.QueryDelegatorDelegationsResponse)
	if err := proto.Unmarshal(resp.Value, balance); err != nil {
		return nil, fmt.Errorf("unable to unmarshal delegator query delegations: %w", err)
	}

	res := sdk.NewCoins()
	for _, i := range balance.DelegationResponses {
		res = res.Add(i.Balance)
	}

	return res, nil
}

func getDelegatorUnbondingDelegationsSum(ctx sdk.Context, address, bondDenom string, queryServer grpc.Server) (sdk.Coins, error) {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer)
	}

	queryFn := querier.Route(delegatorUnbondingDelegationsPath)

	q := &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
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
	resp, err := queryFn(ctx, req)
	if err != nil && !errors.Is(err, sdkerrors.ErrNotFound) {
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("staking query error, %w", err)
	}

	balance := new(stakingtypes.QueryDelegatorUnbondingDelegationsResponse)
	if err := proto.Unmarshal(resp.Value, balance); err != nil {
		return nil, fmt.Errorf("unable to unmarshal delegator query delegations: %w", err)
	}

	res := sdk.NewCoins()
	for _, i := range balance.UnbondingResponses {
		for _, r := range i.Entries {
			res = res.Add(sdk.NewCoin(bondDenom, r.Balance))
		}
	}

	return res, nil
}

func getBalance(ctx sdk.Context, address string, queryServer grpc.Server) (sdk.Coins, error) {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer)
	}

	queryFn := querier.Route(balancesPath)

	q := &banktypes.QueryAllBalancesRequest{
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
	resp, err := queryFn(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("bank query error, %w", err)
	}
	balance := new(banktypes.QueryAllBalancesResponse)
	if err := proto.Unmarshal(resp.Value, balance); err != nil {
		return nil, fmt.Errorf("unable to unmarshal bank balance response: %w", err)
	}
	return balance.Balances, nil
}

func getBondDenom(ctx sdk.Context, queryServer grpc.Server) (string, error) {
	querier, ok := queryServer.(*baseapp.GRPCQueryRouter)
	if !ok {
		return "", fmt.Errorf("unexpected type: %T wanted *baseapp.GRPCQueryRouter", queryServer)
	}

	queryFn := querier.Route(stakingParamsPath)

	q := &stakingtypes.QueryParamsRequest{}

	b, err := proto.Marshal(q)
	if err != nil {
		return "", fmt.Errorf("cannot marshal staking params query request, %w", err)
	}
	req := abci.RequestQuery{
		Data: b,
		Path: stakingParamsPath,
	}

	resp, err := queryFn(ctx, req)
	if err != nil {
		return "", fmt.Errorf("staking query error, %w", err)
	}

	params := new(stakingtypes.QueryParamsResponse)
	if err := proto.Unmarshal(resp.Value, params); err != nil {
		return "", fmt.Errorf("unable to unmarshal delegator query delegations: %w", err)
	}

	return params.Params.BondDenom, nil
}

// MigrateAccount migrates vesting account to make the DelegatedVesting and DelegatedFree fields correctly
// track delegations.
// References: https://github.com/cosmos/cosmos-sdk/issues/8601, https://github.com/cosmos/cosmos-sdk/issues/8812
func MigrateAccount(ctx sdk.Context, account types.AccountI, queryServer grpc.Server) (types.AccountI, error) {
	return migrateVestingAccounts(ctx, account, queryServer)
}
