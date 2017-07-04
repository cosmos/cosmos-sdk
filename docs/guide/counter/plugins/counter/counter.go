package counter

import (
	rawerr "errors"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/types"
)

// CounterTx
//--------------------------------------------------------------------------------

// register the tx type with it's validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	NameCounter = "cntr"
	ByteTx      = 0x21
	TypeTx      = NameCounter + "/count"
)

func init() {
	basecoin.TxMapper.RegisterImplementation(CounterTx{}, TypeTx, ByteTx)
}

type CounterTx struct {
	Valid bool        `json:"valid"`
	Fee   types.Coins `json:"fee"`
}

func NewCounterTx(valid bool, fee types.Coins) basecoin.Tx {
	return CounterTx{
		Valid: valid,
		Fee:   fee,
	}.Wrap()
}

func (c CounterTx) Wrap() basecoin.Tx {
	return basecoin.Tx{c}
}

// ValidateBasic just makes sure the Fee is a valid, non-negative value
func (c CounterTx) ValidateBasic() error {
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
	errInvalidCounter = rawerr.New("Counter Tx marked invalid")
)

// This is a custom error class
func ErrInvalidCounter() error {
	return errors.WithCode(errInvalidCounter, abci.CodeType_BaseInvalidInput)
}
func IsInvalidCounterErr(err error) bool {
	return errors.IsSameError(errInvalidCounter, err)
}

// This is just a helper function to return a generic "internal error"
func ErrDecoding() error {
	return errors.ErrInternal("Error decoding state")
}

// CounterHandler
//--------------------------------------------------------------------------------

func NewCounterHandler() basecoin.Handler {
	// use the default stack
	coin := coin.NewHandler()
	counter := CounterHandler{}
	dispatcher := stack.NewDispatcher(
		stack.WrapHandler(coin),
		stack.WrapHandler(counter),
	)
	return stack.NewDefault().Use(dispatcher)
}

type CounterHandler struct {
	basecoin.NopOption
}

var _ basecoin.Handler = CounterHandler{}

func (_ CounterHandler) Name() string {
	return NameCounter
}

// CheckTx checks if the tx is properly structured
func (h CounterHandler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	_, err = checkTx(ctx, tx)
	return
}

// DeliverTx executes the tx if valid
func (h CounterHandler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	ctr, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}
	// note that we don't assert this on CheckTx (ValidateBasic),
	// as we allow them to be writen to the chain
	if !ctr.Valid {
		return res, ErrInvalidCounter()
	}

	// TODO: handle coin movement.... ugh, need sequence to do this, right?
	// like, actually decrement the other account

	// update the counter
	state, err := LoadState(store)
	if err != nil {
		return res, err
	}
	state.Counter += 1
	state.TotalFees = state.TotalFees.Plus(ctr.Fee)
	err = StoreState(store, state)

	return res, err
}

func checkTx(ctx basecoin.Context, tx basecoin.Tx) (ctr CounterTx, err error) {
	ctr, ok := tx.Unwrap().(CounterTx)
	if !ok {
		return ctr, errors.ErrInvalidFormat(tx)
	}
	err = ctr.ValidateBasic()
	if err != nil {
		return ctr, err
	}
	return ctr, nil
}

// CounterStore
//--------------------------------------------------------------------------------

type CounterState struct {
	Counter   int         `json:"counter"`
	TotalFees types.Coins `json:"total_fees"`
}

func StateKey() []byte {
	return []byte(NameCounter + "/state")
}

func LoadState(store types.KVStore) (state CounterState, err error) {
	bytes := store.Get(StateKey())
	if len(bytes) > 0 {
		err = wire.ReadBinaryBytes(bytes, &state)
		if err != nil {
			return state, errors.ErrDecoding()
		}
	}
	return state, nil
}

func StoreState(store types.KVStore, state CounterState) error {
	bytes := wire.BinaryBytes(state)
	store.Set(StateKey(), bytes)
	return nil
}
