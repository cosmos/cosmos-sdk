package stake

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	bcs "github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

type StakeParams struct {
	UnbondingPeriod uint64
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

	state := loadState(store)

	coll, i := state.Collateral.Get(ctx.CallerAddress, tx.ValidatorPubKey)
	if coll == nil {
		log := fmt.Sprintf(
			"Address %X does not have any collateral delegated to validator %X",
			ctx.CallerAddress,
			tx.ValidatorPubKey,
		)
		return abci.ErrBaseUnknownAddress.AppendLog(log)
	}
	if coll.Amount < tx.Amount {
		log := fmt.Sprintf(
			"Not enough coins bonded (requested=%v, balance=%v)",
			tx.Amount,
			coll.Amount,
		)
		return abci.ErrBaseInsufficientFunds.AppendLog(log)
	}

	// subtract coins from collateral
	if coll.Amount > tx.Amount {
		state.Collateral[i].Amount -= tx.Amount
	} else {
		state.Collateral = state.Collateral.Remove(i)
	}

	// create new unbond record
	state.Unbonding = append(state.Unbonding, Unbond{
		ValidatorPubKey: tx.ValidatorPubKey,
		Amount:          tx.Amount,
		Address:         ctx.CallerAddress,
		Height:          0, // TODO
	})

	saveState(store, state)

	return abci.OK
}

func (sp *StakePlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
	state := loadState(store)

	// create collateral for initial validators
	for _, v := range vals {
		state.Collateral.Add(Collateral{
			ValidatorPubKey: v.PubKey,
			Address:         crypto.Ripemd160(v.PubKey),
			Amount:          v.Power,
		})
	}

	saveState(store, state)
}

func (sp *StakePlugin) BeginBlock(store types.KVStore, hash []byte, header *abci.Header) {
	state := loadState(store)

	// if any unbonding requests have reached maturity,
	// pay out coins into their basecoin accounts
	unbonding := state.Unbonding
	height := header.GetHeight()
	for len(unbonding) > 0 {
		if height-unbonding[0].Height < sp.params.UnbondingPeriod {
			break
		}
		unbond := unbonding[0]
		unbonding = unbonding[1:]
		account := bcs.GetAccount(store, unbond.Address)
		account.Balance = account.Balance.Plus(types.Coins{
			types.Coin{
				Denom:  sp.params.TokenDenom,
				Amount: int64(unbond.Amount),
			},
		})
		bcs.SetAccount(store, unbond.Address, account)
	}

	saveState(store, state)
}

func (sp *StakePlugin) EndBlock(store types.KVStore, height uint64) (res abci.ResponseEndBlock) {
	res.Diffs = loadState(store).Collateral.Validators()
	return
}

func loadState(store types.KVStore) *StakeState {
	bytes := store.Get([]byte("state"))
	if len(bytes) == 0 {
		return &StakeState{}
	}
	var state *StakeState
	err := wire.ReadBinaryBytes(bytes, &state)
	if err != nil {
		panic(err)
	}
	return state
}

func saveState(store types.KVStore, state *StakeState) {
	bytes := wire.BinaryBytes(state)
	store.Set([]byte("state"), bytes)
}
