package counter

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/accountstd"
	counterv1 "cosmossdk.io/x/accounts/testing/counter/v1"
)

var (
	OwnerPrefix   = collections.NewPrefix(0)
	CounterPrefix = collections.NewPrefix(1)
)

var _ accountstd.Interface = (*Account)(nil)

// NewAccount creates a new account.
func NewAccount(d accountstd.Dependencies) (Account, error) {
	return Account{
		Owner:   collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		Counter: collections.NewItem(d.SchemaBuilder, CounterPrefix, "counter", collections.Uint64Value),
	}, nil
}

// Account implements the Account interface. It is an account
// who can be used to increase a counter.
type Account struct {
	// Owner is the address of the account owner.
	Owner collections.Item[[]byte]
	// Counter is the counter value.
	Counter collections.Item[uint64]
}

func (a Account) Init(ctx context.Context, msg *counterv1.MsgInit) (*counterv1.MsgInitResponse, error) {
	err := a.Owner.Set(ctx, accountstd.Sender(ctx))
	if err != nil {
		return nil, err
	}
	err = a.Counter.Set(ctx, msg.InitialValue)
	if err != nil {
		return nil, err
	}
	return &counterv1.MsgInitResponse{}, nil
}

func (a Account) IncreaseCounter(ctx context.Context, msg *counterv1.MsgIncreaseCounter) (*counterv1.MsgIncreaseCounterResponse, error) {
	sender := accountstd.Sender(ctx)
	owner, err := a.Owner.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(sender, owner) {
		return nil, fmt.Errorf("sender is not the owner of the account")
	}
	counter, err := a.Counter.Get(ctx)
	if err != nil {
		return nil, err
	}
	counter += msg.Amount
	err = a.Counter.Set(ctx, counter)
	if err != nil {
		return nil, err
	}
	return &counterv1.MsgIncreaseCounterResponse{
		NewAmount: counter,
	}, nil
}

func (a Account) QueryCounter(ctx context.Context, _ *counterv1.QueryCounterRequest) (*counterv1.QueryCounterResponse, error) {
	counter, err := a.Counter.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &counterv1.QueryCounterResponse{
		Value: counter,
	}, nil
}

func (a Account) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

func (a Account) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.IncreaseCounter)
}

func (a Account) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QueryCounter)
}
