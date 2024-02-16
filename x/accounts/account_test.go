package accounts

import (
	"context"
	"strconv"

	"github.com/cosmos/gogoproto/types"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"
)

var _ implementation.Account = (*TestAccount)(nil)

func NewTestAccount(d accountstd.Dependencies) (*TestAccount, error) {
	return &TestAccount{
		Counter: collections.NewSequence(d.SchemaBuilder, collections.NewPrefix(0), "counter"),
	}, nil
}

type TestAccount struct {
	Counter collections.Sequence
}

func (t TestAccount) RegisterInitHandler(builder *implementation.InitBuilder) {
	implementation.RegisterInitHandler(builder, func(ctx context.Context, _ *types.Empty) (*types.Empty, error) {
		// we also force a module call here to test things work as expected.
		_, err := implementation.QueryModule[bankv1beta1.QueryBalanceResponse](ctx, &bankv1beta1.QueryBalanceRequest{
			Address: string(implementation.Whoami(ctx)),
			Denom:   "atom",
		})
		return &types.Empty{}, err
	})
}

func (t TestAccount) RegisterExecuteHandlers(builder *implementation.ExecuteBuilder) {
	implementation.RegisterExecuteHandler(builder, func(_ context.Context, _ *types.Empty) (*types.Empty, error) {
		return &types.Empty{}, nil
	})

	implementation.RegisterExecuteHandler(builder, func(_ context.Context, req *types.StringValue) (*types.UInt64Value, error) {
		value, err := strconv.ParseUint(req.Value, 10, 64)
		if err != nil {
			return nil, err
		}

		return &types.UInt64Value{Value: value}, nil
	})

	// this is for intermodule comms testing, we simulate a bank send
	implementation.RegisterExecuteHandler(builder, func(ctx context.Context, req *types.Int64Value) (*types.Empty, error) {
		resp, err := implementation.ExecModule[bankv1beta1.MsgSendResponse](ctx, &bankv1beta1.MsgSend{
			FromAddress: string(implementation.Whoami(ctx)),
			ToAddress:   "recipient",
			Amount: []*basev1beta1.Coin{
				{
					Denom:  "test",
					Amount: strconv.FormatInt(req.Value, 10),
				},
			},
		})
		if err != nil {
			return nil, err
		}
		if resp == nil {
			panic("nil response") // should never happen
		}

		return &types.Empty{}, nil
	})

	// genesis testing
	implementation.RegisterExecuteHandler(builder, func(ctx context.Context, req *types.UInt64Value) (*types.Empty, error) {
		return &types.Empty{}, t.Counter.Set(ctx, req.Value)
	})
}

func (t TestAccount) RegisterQueryHandlers(builder *implementation.QueryBuilder) {
	implementation.RegisterQueryHandler(builder, func(_ context.Context, _ *types.Empty) (*types.Empty, error) {
		return &types.Empty{}, nil
	})

	implementation.RegisterQueryHandler(builder, func(_ context.Context, req *types.UInt64Value) (*types.StringValue, error) {
		return &types.StringValue{Value: strconv.FormatUint(req.Value, 10)}, nil
	})

	// test intermodule comms, we simulate someone is sending the account a request for the accounts balance
	// of a given denom.
	implementation.RegisterQueryHandler(builder, func(ctx context.Context, req *types.StringValue) (*types.Int64Value, error) {
		resp, err := implementation.QueryModule[bankv1beta1.QueryBalanceResponse](ctx, &bankv1beta1.QueryBalanceRequest{
			Address: string(implementation.Whoami(ctx)),
			Denom:   req.Value,
		})
		if err != nil {
			return nil, err
		}

		amt, err := strconv.ParseInt(resp.Balance.Amount, 10, 64)
		if err != nil {
			return nil, err
		}
		return &types.Int64Value{Value: amt}, nil
	})

	// genesis testing; DoubleValue does not make sense as a request type for this query, but empty is already taken
	// and this is only used for testing.
	implementation.RegisterQueryHandler(builder, func(ctx context.Context, _ *types.DoubleValue) (*types.UInt64Value, error) {
		v, err := t.Counter.Peek(ctx)
		if err != nil {
			return nil, err
		}
		return &types.UInt64Value{Value: v}, nil
	})
}
