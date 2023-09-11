package accounts

import (
	"context"
	"strconv"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/x/accounts/internal/implementation"
)

var _ implementation.Account = (*TestAccount)(nil)

type TestAccount struct{}

func (t TestAccount) RegisterInitHandler(builder *implementation.InitBuilder) {
	implementation.RegisterInitHandler(builder, func(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	})
}

func (t TestAccount) RegisterExecuteHandlers(builder *implementation.ExecuteBuilder) {
	implementation.RegisterExecuteHandler(builder, func(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	})

	implementation.RegisterExecuteHandler(builder, func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.UInt64Value, error) {
		value, err := strconv.ParseUint(req.Value, 10, 64)
		if err != nil {
			return nil, err
		}

		return wrapperspb.UInt64(value), nil
	})

	// this is for intermodule comms testing, we simulate a bank send
	implementation.RegisterExecuteHandler(builder, func(ctx context.Context, req *wrapperspb.Int64Value) (*emptypb.Empty, error) {
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

		return &emptypb.Empty{}, nil
	})
}

func (t TestAccount) RegisterQueryHandlers(builder *implementation.QueryBuilder) {
	implementation.RegisterQueryHandler(builder, func(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	})

	implementation.RegisterQueryHandler(builder, func(_ context.Context, req *wrapperspb.UInt64Value) (*wrapperspb.StringValue, error) {
		return wrapperspb.String(strconv.FormatUint(req.Value, 10)), nil
	})

	// test intermodule comms, we simulate someone is sending the account a request for the accounts balance
	// of a given denom.
	implementation.RegisterQueryHandler(builder, func(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.Int64Value, error) {
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
		return wrapperspb.Int64(amt), nil
	})
}
