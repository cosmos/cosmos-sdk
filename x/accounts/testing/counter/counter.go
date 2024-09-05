package counter

import (
	"bytes"
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/accounts/accountstd"
	counterv1 "cosmossdk.io/x/accounts/testing/counter/v1"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	OwnerPrefix          = collections.NewPrefix(0)
	CounterPrefix        = collections.NewPrefix(1)
	TestStateCodecPrefix = collections.NewPrefix(2)
)

var _ accountstd.Interface = (*Account)(nil)

// NewAccount creates a new account.
func NewAccount(d accountstd.Dependencies) (Account, error) {
	return Account{
		Owner:          collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		Counter:        collections.NewItem(d.SchemaBuilder, CounterPrefix, "counter", collections.Uint64Value),
		TestStateCodec: collections.NewItem(d.SchemaBuilder, TestStateCodecPrefix, "test_state_codec", codec.CollValue[counterv1.MsgTestDependencies](d.LegacyStateCodec)),
		addressCodec:   d.AddressCodec,
		hs:             d.Environment.HeaderService,
		gs:             d.Environment.GasService,
	}, nil
}

// Account implements the Account interface. It is an account
// who can be used to increase a counter.
type Account struct {
	// Owner is the address of the account owner.
	Owner collections.Item[[]byte]
	// Counter is the counter value.
	Counter collections.Item[uint64]
	// TestStateCodec is used to test the binary codec provided by the runtime.
	// It simply stores the MsgInit.
	TestStateCodec collections.Item[counterv1.MsgTestDependencies]

	hs           header.Service
	addressCodec address.Codec
	gs           gas.Service
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
	// check funds
	return &counterv1.MsgInitResponse{}, nil
}

func (a Account) IncreaseCounter(ctx context.Context, msg *counterv1.MsgIncreaseCounter) (*counterv1.MsgIncreaseCounterResponse, error) {
	sender := accountstd.Sender(ctx)
	owner, err := a.Owner.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(sender, owner) {
		return nil, errors.New("sender is not the owner of the account")
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

func (a Account) TestDependencies(ctx context.Context, _ *counterv1.MsgTestDependencies) (*counterv1.MsgTestDependenciesResponse, error) {
	// test binary codec
	err := a.TestStateCodec.Set(ctx, counterv1.MsgTestDependencies{})
	if err != nil {
		return nil, err
	}

	// test address codec
	me := accountstd.Whoami(ctx)
	meStr, err := a.addressCodec.BytesToString(me)
	if err != nil {
		return nil, err
	}

	// test header service
	chainID := a.hs.HeaderInfo(ctx).ChainID

	// test gas meter
	gm := a.gs.GasMeter(ctx)
	gasBefore := gm.Limit() - gm.Remaining()
	if err := gm.Consume(10, "test"); err != nil {
		return nil, err
	}
	gasAfter := gm.Limit() - gm.Remaining()

	// test funds
	funds := accountstd.Funds(ctx)
	if len(funds) == 0 {
		return nil, errors.New("expected funds")
	}

	return &counterv1.MsgTestDependenciesResponse{
		ChainId:   chainID,
		Address:   meStr,
		BeforeGas: gasBefore,
		AfterGas:  gasAfter,
		Funds:     funds,
	}, nil
}

func (a Account) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

func (a Account) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.IncreaseCounter)
	accountstd.RegisterExecuteHandler(builder, a.TestDependencies)
}

func (a Account) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QueryCounter)
}
