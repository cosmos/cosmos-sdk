package stake

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

type StakeParams struct {
	UnbondingPeriod uint
	TokenDenom      string
}

type StakePlugin struct {
	params StakeParams
}

func New(params StakeParams) *StakePlugin {
	return &StakePlugin{params}
}

func (sp *StakePlugin) Name() string {
	return "stake"
}

func (sp *StakePlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	state := loadState(store)
	if state == nil {
		state = &StakeState{}
		state.Collateral = make(Collaterals, 0)
		state.Unbonding = make([]Unbond, 0)
	}

	if key == "collateral" {
		// TODO: unmarshal, then state.Collateral.Add()
		saveState(store, state)
		return fmt.Sprintf("got collateral option:\n%s\n", value)
	}

	panic(fmt.Sprintf("Unknown option key '%s'", key))
}

func (sp *StakePlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
	var tx Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	if _, isBondTx := tx.(BondTx); isBondTx {
		return sp.runBondTx(tx.(BondTx), store, ctx)
	}
	return sp.runUnbondTx(tx.(UnbondTx), store, ctx)
}

func (sp *StakePlugin) runBondTx(tx BondTx, store types.KVStore, ctx types.CallContext) (res abci.Result) {
	if len(ctx.Coins) != 1 {
		log := "Must only use one denomination"
		return abci.ErrInternalError.AppendLog(log)
	}
	if ctx.Coins[0].Denom != sp.params.TokenDenom {
		log := fmt.Sprintf("Collateral must be denomination '%s'", sp.params.TokenDenom)
		return abci.ErrInternalError.AppendLog(log)
	}
	amount := ctx.Coins[0].Amount
	if amount <= 0 {
		log := "Amount must be > 0"
		return abci.ErrInternalError.AppendLog(log)
	}

	state := loadState(store)
	state.Collateral = state.Collateral.Add(Collateral{
		ValidatorPubKey: tx.ValidatorPubKey,
		Address:         ctx.CallerAddress,
		Amount:          uint64(amount),
	})
	saveState(store, state)

	return abci.OK
}

func (sp *StakePlugin) runUnbondTx(tx UnbondTx, store types.KVStore, ctx types.CallContext) (res abci.Result) {
	if tx.Amount <= 0 {
		log := "Unbond amount must be > 0"
		return abci.ErrInternalError.AppendLog(log)
	}

	// TODO

	return abci.OK
}

func (sp *StakePlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
	// TODO: ensure genesis validator set has bonded collateral,
	// and voting power matches collateral amounts
}

func (sp *StakePlugin) BeginBlock(store types.KVStore, height uint64) {}

func (sp *StakePlugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	return loadState(store).Collateral.Validators()
}

func loadState(store types.KVStore) (state *StakeState) {
	bytes := store.Get([]byte("state"))
	if len(bytes) == 0 {
		return nil
	}
	err := wire.ReadBinaryBytes(bytes, state)
	if err != nil {
		panic(err)
	}
	return state
}

func saveState(store types.KVStore, state *StakeState) {
	bytes := wire.BinaryBytes(*state)
	store.Set([]byte("state"), bytes)
}
