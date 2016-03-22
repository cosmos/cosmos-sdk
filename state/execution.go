package state

import (
	"bytes"

	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-events"
)

// The accounts from the TxInputs must either already have
// crypto.PubKey.(type) != nil, (it must be known),
// or it must be specified in the TxInput.  If redeclared,
// the TxInput is modified and input.PubKey set to nil.
func getInputs(state types.AccountGetter, ins []types.TxInput) (map[string]*types.Account, error) {
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
		if err := checkInputPubKey(in.Address, acc, in); err != nil {
			return nil, err
		}
		accounts[string(in.Address)] = acc
	}
	return accounts, nil
}

func getOrMakeOutputs(state types.AccountGetter, accounts map[string]*types.Account, outs []types.TxOutput) (map[string]*types.Account, error) {
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
	return accounts, nil
}

// Input must not have a redundant PubKey (i.e. Account already has PubKey).
// NOTE: Account has PubKey if Sequence > 0
func checkInputPubKey(address []byte, acc *types.Account, in types.TxInput) error {
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
	return nil
}

func validateInputs(accounts map[string]*types.Account, signBytes []byte, ins []types.TxInput) (total int64, err error) {
	for _, in := range ins {
		acc := accounts[string(in.Address)]
		if acc == nil {
			PanicSanity("validateInputs() expects account in accounts")
		}
		err = validateInput(acc, signBytes, in)
		if err != nil {
			return
		}
		// Good. Add amount to total
		total += in.Amount
	}
	return total, nil
}

func validateInput(acc *types.Account, signBytes []byte, in types.TxInput) (err error) {
	// Check TxInput basic
	if err := in.ValidateBasic(); err != nil {
		return err
	}
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, in.Signature) {
		return types.ErrInvalidSignature
	}
	// Check sequences
	if acc.Sequence+1 != in.Sequence {
		return types.ErrInvalidSequence.AppendLog(Fmt("Got %v, expected %v", in.Sequence, acc.Sequence+1))
	}
	// Check amount
	if acc.Balance < in.Amount {
		return types.ErrInsufficientFunds
	}
	return nil
}

func validateOutputs(outs []types.TxOutput) (total int64, err error) {
	for _, out := range outs {
		// Check TxOutput basic
		if err := out.ValidateBasic(); err != nil {
			return 0, err
		}
		// Good. Add amount to total
		total += out.Amount
	}
	return total, nil
}

func adjustByInputs(accounts map[string]*types.Account, ins []types.TxInput) {
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
	}
}

func adjustByOutputs(accounts map[string]*types.Account, outs []types.TxOutput) {
	for _, out := range outs {
		acc := accounts[string(out.Address)]
		if acc == nil {
			PanicSanity("adjustByOutputs() expects account in accounts")
		}
		acc.Balance += out.Amount
	}
}

// If the tx is invalid, an error will be returned.
// Unlike ExecBlock(), state will not be altered.
func ExecTx(state *State, tx types.Tx, runCall bool, evc events.Fireable) (err error) {

	// TODO: do something with fees
	fees := int64(0)

	// Exec tx
	switch tx := tx.(type) {
	case *types.SendTx:
		accounts, err := getInputs(state, tx.Inputs)
		if err != nil {
			return err
		}

		// add outputs to accounts map
		// if any outputs don't exist, all inputs must have CreateAccount perm
		accounts, err = getOrMakeOutputs(state, accounts, tx.Outputs)
		if err != nil {
			return err
		}

		signBytes := tx.SignBytes(state.ChainID())
		inTotal, err := validateInputs(accounts, signBytes, tx.Inputs)
		if err != nil {
			return err
		}
		outTotal, err := validateOutputs(tx.Outputs)
		if err != nil {
			return err
		}
		if outTotal > inTotal {
			return types.ErrInsufficientFunds
		}
		fee := inTotal - outTotal
		fees += fee

		// Good! Adjust accounts
		adjustByInputs(accounts, tx.Inputs)
		adjustByOutputs(accounts, tx.Outputs)
		for _, acc := range accounts {
			state.SetAccount(acc)
		}

		// if the evc is nil, nothing will happen
		/*
			if evc != nil {
				for _, i := range tx.Inputs {
					evc.FireEvent(types.EventStringAccInput(i.Address), types.EventDataTx{tx, nil, ""})
				}
				for _, o := range tx.Outputs {
					evc.FireEvent(types.EventStringAccOutput(o.Address), types.EventDataTx{tx, nil, ""})
				}
			}
		*/
		return nil

	case *types.CallTx:
		var inAcc, outAcc *types.Account

		// Validate input
		inAcc = state.GetAccount(tx.Input.Address)
		if inAcc == nil {
			log.Info(Fmt("Can't find in account %X", tx.Input.Address))
			return types.ErrInvalidAddress
		}

		// pubKey should be present in either "inAcc" or "tx.Input"
		if err := checkInputPubKey(tx.Input.Address, inAcc, tx.Input); err != nil {
			log.Info(Fmt("Can't find pubkey for %X", tx.Input.Address))
			return err
		}
		signBytes := tx.SignBytes(state.ChainID())
		err := validateInput(inAcc, signBytes, tx.Input)
		if err != nil {
			log.Info(Fmt("validateInput failed on %X: %v", tx.Input.Address, err))
			return err
		}
		if tx.Input.Amount < tx.Fee {
			log.Info(Fmt("Sender did not send enough to cover the fee %X", tx.Input.Address))
			return types.ErrInsufficientFunds
		}

		if len(tx.Address) == 0 {
			return types.ErrInvalidAddress.AppendLog("Address cannot be zero")
		}
		// Validate output
		if len(tx.Address) != 20 {
			log.Info(Fmt("Destination address is not 20 bytes %X", tx.Address))
			return types.ErrInvalidAddress
		}
		// check if its a native contract
		// XXX if IsNativeContract(tx.Address) {...}

		// Output account may be nil if we are still in mempool and contract was created in same block as this tx
		// but that's fine, because the account will be created properly when the create tx runs in the block
		// and then this won't return nil. otherwise, we take their fee
		outAcc = state.GetAccount(tx.Address)

		log.Info(Fmt("Out account: %v", outAcc))

		// Good!
		value := tx.Input.Amount - tx.Fee
		inAcc.Sequence += 1
		inAcc.Balance -= tx.Fee
		state.SetAccount(inAcc)

		// The logic in runCall MUST NOT return.
		if runCall {

			// VM call variables
			var (
				// gas int64 = tx.GasLimit
				err error = nil
				// caller  *vm.Account = toVMAccount(inAcc)
				// callee  *vm.Account = nil // initialized below
				// code    []byte = nil
				// ret     []byte = nil
				// txCache = NewTxCache(state)
				/*
					params  = vm.Params{
						BlockHeight: int64(state.LastBlockHeight),
						BlockHash:   LeftPadWord256(state.LastBlockHash),
						BlockTime:   state.LastBlockTime.Unix(),
						GasLimit:    state.GetGasLimit(),
					}
				*/
			)

			// if you call an account that doesn't exist
			// or an account with no code then we take fees (sorry pal)
			// NOTE: it's fine to create a contract and call it within one
			// block (nonce will prevent re-ordering of those txs)
			// but to create with one contract and call with another
			// you have to wait a block to avoid a re-ordering attack
			// that will take your fees
			if outAcc == nil {
				log.Info(Fmt("%X tries to call %X but it does not exist.", tx.Input.Address, tx.Address))
				err = types.ErrInvalidAddress
				goto CALL_COMPLETE
			}
			/*
				if len(outAcc.Code) == 0 {
					log.Info(Fmt("%X tries to call %X but code is blank.", inAcc.Address, tx.Address))
					err = types.ErrInvalidAddress
					goto CALL_COMPLETE
				}
				log.Info(Fmt("Code for this contract: %X", code))
			*/

			// Run VM call and sync txCache to state.
			{ // Capture scope for goto.
				// Write caller/callee to txCache.
				// txCache.SetAccount(caller)
				// txCache.SetAccount(callee)
				// vmach := vm.NewVM(txCache, params, caller.Address, types.TxID(state.ChainID(), tx))
				// vmach.SetFireable(evc)
				// NOTE: Call() transfers the value from caller to callee iff call succeeds.
				// ret, err = vmach.Call(caller, callee, code, tx.Data, value, &gas)
				if err != nil {
					// Failure. Charge the gas fee. The 'value' was otherwise not transferred.
					log.Info(Fmt("Error on execution: %v", err))
					goto CALL_COMPLETE
				}
				log.Info("Successful execution")
				// txCache.Sync()
			}

		CALL_COMPLETE: // err may or may not be nil.

			// Create a receipt from the ret and whether errored.
			// log.Notice("VM call complete", "caller", caller, "callee", callee, "return", ret, "err", err)

			// Fire Events for sender and receiver
			// a separate event will be fired from vm for each additional call
			/*
				if evc != nil {
					exception := ""
					if err != nil {
						exception = err.Error()
					}
					evc.FireEvent(types.EventStringAccInput(tx.Input.Address), types.EventDataTx{tx, ret, exception})
					evc.FireEvent(types.EventStringAccOutput(tx.Address), types.EventDataTx{tx, ret, exception})
				}
			*/
		} else {
			// The mempool does not call txs until
			// the proposer determines the order of txs.
			// So mempool will skip the actual .Call(),
			// and only deduct from the caller's balance.
			inAcc.Balance -= value
			state.SetAccount(inAcc)
		}

		return nil

	default:
		// binary decoding should not let this happen
		PanicSanity("Unknown Tx type")
		return nil
	}
}

//-----------------------------------------------------------------------------

type InvalidTxError struct {
	Tx     types.Tx
	Reason error
}

func (txErr InvalidTxError) Error() string {
	return Fmt("Invalid tx: [%v] reason: [%v]", txErr.Tx, txErr.Reason)
}
