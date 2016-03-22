package state

import (
	"bytes"
	"strings"

	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-events"
	tmsp "github.com/tendermint/tmsp/types"
)

// If the tx is invalid, a TMSP error will be returned.
func ExecTx(state *State, tx types.Tx, isCheckTx bool, evc events.Fireable) tmsp.Result {

	// TODO: do something with fees
	fees := int64(0)

	// Exec tx
	switch tx := tx.(type) {
	case *types.SendTx:
		// First, get inputs
		accounts, res := getInputs(state, tx.Inputs)
		if !res.IsOK() {
			return res
		}

		// Then, get or make outputs.
		accounts, res = getOrMakeOutputs(state, accounts, tx.Outputs)
		if !res.IsOK() {
			return res
		}

		// Validate inputs and outputs
		signBytes := tx.SignBytes(state.GetChainID())
		inTotal, res := validateInputs(state, accounts, signBytes, tx.Inputs)
		if !res.IsOK() {
			return res
		}
		outTotal, res := validateOutputs(tx.Outputs)
		if !res.IsOK() {
			return res
		}
		if outTotal > inTotal {
			return types.ErrInsufficientFunds
		}
		fee := inTotal - outTotal
		fees += fee

		// TODO: Fee validation for SendTx

		// Good! Adjust accounts
		adjustByInputs(state, accounts, tx.Inputs, isCheckTx)
		adjustByOutputs(state, accounts, tx.Outputs, isCheckTx)

		/*
			// Fire events
			if !isCheckTx {
				if evc != nil {
					for _, i := range tx.Inputs {
						evc.FireEvent(types.EventStringAccInput(i.Address), types.EventDataTx{tx, nil, ""})
					}
					for _, o := range tx.Outputs {
						evc.FireEvent(types.EventStringAccOutput(o.Address), types.EventDataTx{tx, nil, ""})
					}
				}
			}
		*/

		return types.ResultOK

	case *types.CallTx:
		// First, get input account
		inAcc := state.GetAccount(tx.Input.Address)
		if inAcc == nil {
			log.Info(Fmt("Can't find in account %X", tx.Input.Address))
			return types.ErrInvalidAddress
		}

		// Validate input
		// pubKey should be present in either "inAcc" or "tx.Input"
		if res := checkInputPubKey(tx.Input.Address, inAcc, tx.Input); !res.IsOK() {
			log.Info(Fmt("Can't find pubkey for %X", tx.Input.Address))
			return res
		}
		signBytes := tx.SignBytes(state.GetChainID())
		res := validateInput(state, inAcc, signBytes, tx.Input)
		if !res.IsOK() {
			log.Info(Fmt("validateInput failed on %X: %v", tx.Input.Address, res))
			return res
		}
		if tx.Input.Amount < tx.Fee {
			log.Info(Fmt("Sender did not send enough to cover the fee %X", tx.Input.Address))
			return types.ErrInsufficientFunds
		}

		// Validate call address
		if strings.HasPrefix(string(tx.Address), "gov/") {
			// This is a gov call.
		} else {
			return types.ErrInvalidAddress.AppendLog(Fmt("Unrecognized address %X", tx.Address))
		}

		// Good!
		value := tx.Input.Amount - tx.Fee
		inAcc.Sequence += 1
		inAcc.Balance -= tx.Input.Amount
		state.SetCheckAccount(tx.Input.Address, inAcc.Sequence, inAcc.Balance)

		// If this is AppendTx, actually save accounts
		if !isCheckTx {
			state.SetAccount(inAcc)
			// NOTE: value is dangling.
			// XXX: don't just give it back
			inAcc.Balance += value
			// TODO: logic.
			// TODO: persist
			// state.SetAccount(inAcc)
			log.Info("Successful execution")
			// Fire events
			/*
				if evc != nil {
					exception := ""
					if !res.IsOK() {
						exception = res.Error()
					}
					evc.FireEvent(types.EventStringAccInput(tx.Input.Address), types.EventDataTx{tx, ret, exception})
					evc.FireEvent(types.EventStringAccOutput(tx.Address), types.EventDataTx{tx, ret, exception})
				}
			*/
		}

		return types.ResultOK

	default:
		PanicSanity("Unknown Tx type")
		return types.ErrInternalError
	}
}

//--------------------------------------------------------------------------------

// The accounts from the TxInputs must either already have
// crypto.PubKey.(type) != nil, (it must be known),
// or it must be specified in the TxInput.  If redeclared,
// the TxInput is modified and input.PubKey set to nil.
func getInputs(state types.AccountGetter, ins []types.TxInput) (map[string]*types.Account, tmsp.Result) {
	accounts := map[string]*types.Account{}
	for _, in := range ins {
		// Account shouldn't be duplicated
		if _, ok := accounts[string(in.Address)]; ok {
			return nil, types.ErrDuplicateAddress
		}
		acc := state.GetAccount(in.Address)
		if acc == nil {
			return nil, types.ErrInvalidAddress
		}
		// PubKey should be present in either "account" or "in"
		if res := checkInputPubKey(in.Address, acc, in); !res.IsOK() {
			return nil, res
		}
		accounts[string(in.Address)] = acc
	}
	return accounts, types.ResultOK
}

func getOrMakeOutputs(state types.AccountGetter, accounts map[string]*types.Account, outs []types.TxOutput) (map[string]*types.Account, tmsp.Result) {
	if accounts == nil {
		accounts = make(map[string]*types.Account)
	}

	for _, out := range outs {
		// Account shouldn't be duplicated
		if _, ok := accounts[string(out.Address)]; ok {
			return nil, types.ErrDuplicateAddress
		}
		acc := state.GetAccount(out.Address)
		// output account may be nil (new)
		if acc == nil {
			acc = &types.Account{
				PubKey:   nil,
				Sequence: 0,
				Balance:  0,
			}
		}
		accounts[string(out.Address)] = acc
	}
	return accounts, types.ResultOK
}

// Input must not have a redundant PubKey (i.e. Account already has PubKey).
// NOTE: Account has PubKey if Sequence > 0
func checkInputPubKey(address []byte, acc *types.Account, in types.TxInput) tmsp.Result {
	if acc.PubKey == nil {
		if in.PubKey == nil {
			return types.ErrUnknownPubKey
		}
		if !bytes.Equal(in.PubKey.Address(), address) {
			return types.ErrInvalidPubKey
		}
		acc.PubKey = in.PubKey
	} else {
		if in.PubKey != nil {
			return types.ErrInvalidPubKey
		}
	}
	return types.ResultOK
}

// Validate inputs and compute total amount
func validateInputs(state *State, accounts map[string]*types.Account, signBytes []byte, ins []types.TxInput) (total int64, res tmsp.Result) {

	for _, in := range ins {
		acc := accounts[string(in.Address)]
		if acc == nil {
			PanicSanity("validateInputs() expects account in accounts")
		}
		res = validateInput(state, acc, signBytes, in)
		if !res.IsOK() {
			return
		}
		// Good. Add amount to total
		total += in.Amount
	}
	return total, types.ResultOK
}

func validateInput(state *State, acc *types.Account, signBytes []byte, in types.TxInput) (res tmsp.Result) {
	// Check TxInput basic
	if res := in.ValidateBasic(); !res.IsOK() {
		return res
	}
	// Check sequence/balance
	seq, balance := state.GetCheckAccount(in.Address, acc.Sequence, acc.Balance)
	if seq+1 != in.Sequence {
		return types.ErrInvalidSequence.AppendLog(Fmt("Got %v, expected %v. (acc.seq=%v)", in.Sequence, seq+1, acc.Sequence))
	}
	// Check amount
	if balance < in.Amount {
		return types.ErrInsufficientFunds
	}
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, in.Signature) {
		return types.ErrInvalidSignature
	}
	return types.ResultOK
}

func validateOutputs(outs []types.TxOutput) (total int64, res tmsp.Result) {
	for _, out := range outs {
		// Check TxOutput basic
		if res := out.ValidateBasic(); !res.IsOK() {
			return 0, res
		}
		// Good. Add amount to total
		total += out.Amount
	}
	return total, types.ResultOK
}

func adjustByInputs(state *State, accounts map[string]*types.Account, ins []types.TxInput, isCheckTx bool) {
	for _, in := range ins {
		acc := accounts[string(in.Address)]
		if acc == nil {
			PanicSanity("adjustByInputs() expects account in accounts")
		}
		if acc.Balance < in.Amount {
			PanicSanity("adjustByInputs() expects sufficient funds")
		}
		acc.Balance -= in.Amount
		acc.Sequence += 1
		state.SetCheckAccount(in.Address, acc.Sequence, acc.Balance)
		if !isCheckTx {
			// NOTE: Must be set in deterministic order
			state.SetAccount(acc)
		}
	}
}

func adjustByOutputs(state *State, accounts map[string]*types.Account, outs []types.TxOutput, isCheckTx bool) {
	for _, out := range outs {
		acc := accounts[string(out.Address)]
		if acc == nil {
			PanicSanity("adjustByOutputs() expects account in accounts")
		}
		acc.Balance += out.Amount
		if !isCheckTx {
			state.SetCheckAccount(out.Address, acc.Sequence, acc.Balance)
			// NOTE: Must be set in deterministic order
			state.SetAccount(acc)
		}
	}
}
