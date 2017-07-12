package counter

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with it's validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	NameCounter = "cntr"
	ByteTx      = 0x2F //TODO What does this byte represent should use typebytes probably
	TypeTx      = NameCounter + "/count"
)

func init() {
	basecoin.TxMapper.RegisterImplementation(Tx{}, TypeTx, ByteTx)
}

// Tx - struct for all counter transactions
type Tx struct {
	Valid    bool       `json:"valid"`
	Fee      coin.Coins `json:"fee"`
	Sequence int        `json:"sequence"`
}

// NewTx - return a new counter transaction struct wrapped as a basecoin transaction
func NewTx(valid bool, fee coin.Coins, sequence int) basecoin.Tx {
	return Tx{
		Valid:    valid,
		Fee:      fee,
		Sequence: sequence,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx, used to satisfy the XXX interface
func (c Tx) Wrap() basecoin.Tx {
	return basecoin.Tx{TxInner: c}
}

// ValidateBasic just makes sure the Fee is a valid, non-negative value
func (c Tx) ValidateBasic() error {
	if !c.Fee.IsValid() {
		return coin.ErrInvalidCoins()
	}
	if !c.Fee.IsNonnegative() {
		return coin.ErrInvalidCoins()
	}
	return nil
}

// Custom errors
//--------------------------------------------------------------------------------

var (
	errInvalidCounter = fmt.Errorf("Counter Tx marked invalid")
)

// ErrInvalidCounter - custom error class
func ErrInvalidCounter() error {
	return errors.WithCode(errInvalidCounter, abci.CodeType_BaseInvalidInput)
}

// IsInvalidCounterErr - custom error class check
func IsInvalidCounterErr(err error) bool {
	return errors.IsSameError(errInvalidCounter, err)
}

// ErrDecoding - This is just a helper function to return a generic "internal error"
func ErrDecoding() error {
	return errors.ErrInternal("Error decoding state")
}

// Counter Handler
//--------------------------------------------------------------------------------

// NewHandler returns a new counter transaction processing handler
func NewHandler() basecoin.Handler {
	// use the default stack
	coin := coin.NewHandler()
	counter := Handler{}
	dispatcher := stack.NewDispatcher(
		stack.WrapHandler(coin),
		counter,
	)
	return stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
	).Use(dispatcher)
}

// Handler the counter transaction processing handler
type Handler struct {
	stack.NopOption
}

var _ stack.Dispatchable = Handler{}

// Name - return counter namespace
func (Handler) Name() string {
	return NameCounter
}

// AssertDispatcher - placeholder to satisfy XXX
func (Handler) AssertDispatcher() {}

// CheckTx checks if the tx is properly structured
func (h Handler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, _ basecoin.Checker) (res basecoin.Result, err error) {
	_, err = checkTx(ctx, tx)
	return
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, dispatch basecoin.Deliver) (res basecoin.Result, err error) {
	ctr, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}
	// note that we don't assert this on CheckTx (ValidateBasic),
	// as we allow them to be writen to the chain
	if !ctr.Valid {
		return res, ErrInvalidCounter()
	}

	// handle coin movement.... like, actually decrement the other account
	if !ctr.Fee.IsZero() {
		// take the coins and put them in out account!
		senders := ctx.GetPermissions("", auth.NameSigs)
		if len(senders) == 0 {
			return res, errors.ErrMissingSignature()
		}
		in := []coin.TxInput{{Address: senders[0], Coins: ctr.Fee}}
		out := []coin.TxOutput{{Address: StoreActor(), Coins: ctr.Fee}}
		send := coin.NewSendTx(in, out)
		// if the deduction fails (too high), abort the command
		_, err = dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return res, err
		}
	}

	// update the counter
	state, err := LoadState(store)
	if err != nil {
		return res, err
	}
	state.Counter++
	state.TotalFees = state.TotalFees.Plus(ctr.Fee)
	err = SaveState(store, state)

	return res, err
}

func checkTx(ctx basecoin.Context, tx basecoin.Tx) (ctr Tx, err error) {
	ctr, ok := tx.Unwrap().(Tx)
	if !ok {
		return ctr, errors.ErrInvalidFormat(TypeTx, tx)
	}
	err = ctr.ValidateBasic()
	if err != nil {
		return ctr, err
	}
	return ctr, nil
}

// CounterStore
//--------------------------------------------------------------------------------

// StoreActor - return the basecoin actor for the account
func StoreActor() basecoin.Actor {
	return basecoin.Actor{App: NameCounter, Address: []byte{0x04, 0x20}} //XXX what do these bytes represent? - should use typebyte variables
}

// State - state of the counter applicaton
type State struct {
	Counter   int        `json:"counter"`
	TotalFees coin.Coins `json:"total_fees"`
}

// StateKey - store key for the counter state
func StateKey() []byte {
	return []byte("state")
}

// LoadState - retrieve the counter state from the store
func LoadState(store state.KVStore) (state State, err error) {
	bytes := store.Get(StateKey())
	if len(bytes) > 0 {
		err = wire.ReadBinaryBytes(bytes, &state)
		if err != nil {
			return state, errors.ErrDecoding()
		}
	}
	return state, nil
}

// SaveState - save the counter state to the provided store
func SaveState(store state.KVStore, state State) error {
	bytes := wire.BinaryBytes(state)
	store.Set(StateKey(), bytes)
	return nil
}
