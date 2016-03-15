package app

import (
	"fmt"

	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	gov "github.com/tendermint/governmint/gov"
	eyes "github.com/tendermint/merkleeyes/client"
	tmsp "github.com/tendermint/tmsp/types"
)

const version = "0.1"
const maxTxSize = 10240

type Basecoin struct {
	eyesCli *eyes.Client
	govMint *gov.Governmint
}

func NewBasecoin(eyesCli *eyes.Client) *Basecoin {
	return &Basecoin{
		eyesCli: eyesCli,
		govMint: gov.NewGovernmint(eyesCli),
	}
}

// TMSP::Info
func (app *Basecoin) Info() string {
	return Fmt("Basecoin v%v\n - %v", version, app.govMint.Info())
}

// TMSP::SetOption
func (app *Basecoin) SetOption(key string, value string) (log string) {
	if key == "setAccount" {
		var err error
		var setAccount types.PubAccount
		wire.ReadJSONPtr(&setAccount, []byte(value), &err)
		if err != nil {
			return "Error decoding setAccount message: " + err.Error()
		}
		pubKeyBytes := wire.BinaryBytes(setAccount.PubKey)
		accBytes := wire.BinaryBytes(setAccount.Account)
		err = app.eyesCli.SetSync(pubKeyBytes, accBytes)
		if err != nil {
			return "Error saving account: " + err.Error()
		}
		return "Success"
	}
	return "Unrecognized option key " + key
}

// TMSP::AppendTx
func (app *Basecoin) AppendTx(txBytes []byte) (code tmsp.CodeType, result []byte, log string) {
	if len(txBytes) > maxTxSize {
		return tmsp.CodeType_BaseEncodingError, nil, "Tx size exceeds maximum"
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return tmsp.CodeType_BaseEncodingError, nil, "Error decoding tx: " + err.Error()
	}
	// Validate tx
	code, errStr := validateTx(tx)
	if errStr != "" {
		return code, nil, "Error validating tx: " + errStr
	}
	// Load accounts
	accMap := loadAccounts(app.eyesCli, allPubKeys(tx))
	// Execute tx
	accs, code, errStr := runTx(tx, accMap, false)
	if errStr != "" {
		return code, nil, "Error executing tx: " + errStr
	}
	// Store accounts
	storeAccounts(app.eyesCli, accs)
	return tmsp.CodeType_OK, nil, "Success"
}

// TMSP::CheckTx
func (app *Basecoin) CheckTx(txBytes []byte) (code tmsp.CodeType, result []byte, log string) {
	if len(txBytes) > maxTxSize {
		return tmsp.CodeType_BaseEncodingError, nil, "Tx size exceeds maximum"
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return tmsp.CodeType_BaseEncodingError, nil, "Error decoding tx: " + err.Error()
	}
	// Validate tx
	code, errStr := validateTx(tx)
	if errStr != "" {
		return code, nil, "Error validating tx: " + errStr
	}
	// Load accounts
	accMap := loadAccounts(app.eyesCli, allPubKeys(tx))
	// Execute tx
	_, code, errStr = runTx(tx, accMap, false)
	if errStr != "" {
		return code, nil, "Error (mock) executing tx: " + errStr
	}
	return tmsp.CodeType_OK, nil, "Success"
}

// TMSP::Query
func (app *Basecoin) Query(query []byte) (code tmsp.CodeType, result []byte, log string) {
	return tmsp.CodeType_OK, nil, ""
	value, err := app.eyesCli.GetSync(query)
	if err != nil {
		panic("Error making query: " + err.Error())
	}
	return tmsp.CodeType_OK, value, "Success"
}

// TMSP::Commit
func (app *Basecoin) Commit() (hash []byte, log string) {
	hash, log, err := app.eyesCli.CommitSync()
	if err != nil {
		panic("Error getting hash: " + err.Error())
	}
	return hash, "Success"
}

// TMSP::InitChain
func (app *Basecoin) InitChain(validators []*tmsp.Validator) {
	app.govMint.InitChain(validators)
}

// TMSP::EndBlock
func (app *Basecoin) EndBlock(height uint64) []*tmsp.Validator {
	return app.govMint.EndBlock(height)
}

//----------------------------------------

func validateTx(tx types.Tx) (code tmsp.CodeType, errStr string) {
	inputs, outputs := tx.GetInputs(), tx.GetOutputs()
	if len(inputs) == 0 {
		return tmsp.CodeType_BaseEncodingError, "Tx.Inputs length cannot be 0"
	}
	seenPubKeys := map[string]bool{}
	signBytes := tx.SignBytes()
	for _, input := range inputs {
		code, errStr = validateInput(input, signBytes)
		if errStr != "" {
			return
		}
		keyString := input.PubKey.KeyString()
		if seenPubKeys[keyString] {
			return tmsp.CodeType_BaseEncodingError, "Duplicate input pubKey"
		}
		seenPubKeys[keyString] = true
	}
	for _, output := range outputs {
		code, errStr = validateOutput(output)
		if errStr != "" {
			return
		}
		keyString := output.PubKey.KeyString()
		if seenPubKeys[keyString] {
			return tmsp.CodeType_BaseEncodingError, "Duplicate output pubKey"
		}
		seenPubKeys[keyString] = true
	}
	sumInputs, overflow := sumAmounts(inputs, nil, 0)
	if overflow {
		return tmsp.CodeType_BaseEncodingError, "Input amount overflow"
	}
	sumOutputsPlus, overflow := sumAmounts(nil, outputs, len(inputs)+len(outputs))
	if overflow {
		return tmsp.CodeType_BaseEncodingError, "Output amount overflow"
	}
	if sumInputs < sumOutputsPlus {
		return tmsp.CodeType_BaseInsufficientFees, "Insufficient fees"
	}
	return tmsp.CodeType_OK, ""
}

func validateInput(input types.Input, signBytes []byte) (code tmsp.CodeType, errStr string) {
	if input.Amount == 0 {
		return tmsp.CodeType_BaseEncodingError, "Input amount cannot be zero"
	}
	if input.PubKey == nil {
		return tmsp.CodeType_BaseEncodingError, "Input pubKey cannot be nil"
	}
	if !input.PubKey.VerifyBytes(signBytes, input.Signature) {
		return tmsp.CodeType_BaseUnauthorized, "Invalid signature"
	}
	return tmsp.CodeType_OK, ""
}

func validateOutput(output types.Output) (code tmsp.CodeType, errStr string) {
	if output.Amount == 0 {
		return tmsp.CodeType_BaseEncodingError, "Output amount cannot be zero"
	}
	if output.PubKey == nil {
		return tmsp.CodeType_BaseEncodingError, "Output pubKey cannot be nil"
	}
	return tmsp.CodeType_OK, ""
}

func sumAmounts(inputs []types.Input, outputs []types.Output, more int) (total uint64, overflow bool) {
	total = uint64(more)
	for _, input := range inputs {
		total2 := total + input.Amount
		if total2 < total {
			return 0, true
		}
		total = total2
	}
	for _, output := range outputs {
		total2 := total + output.Amount
		if total2 < total {
			return 0, true
		}
		total = total2
	}
	return total, false
}

// Returns accounts in order of types.Tx inputs and outputs
// appendTx: true if this is for AppendTx.
// TODO: create more intelligent sequence-checking.  Current impl is just for a throughput demo.
func runTx(tx types.Tx, accMap map[string]types.PubAccount, appendTx bool) (accs []types.PubAccount, code tmsp.CodeType, errStr string) {
	switch tx := tx.(type) {
	case *types.SendTx:
		return runSendTx(tx, accMap, appendTx)
	case *types.GovTx:
		return runGovTx(tx, accMap, appendTx)
	}
	return nil, tmsp.CodeType_InternalError, "Unknown transaction type"
}

func processInputsOutputs(tx types.Tx, accMap map[string]types.PubAccount, appendTx bool) (accs []types.PubAccount, code tmsp.CodeType, errStr string) {
	inputs, outputs := tx.GetInputs(), tx.GetOutputs()
	accs = make([]types.PubAccount, 0, len(inputs)+len(outputs))
	// Deduct from inputs
	// TODO refactor, duplicated code.
	for _, input := range inputs {
		var acc, ok = accMap[input.PubKey.KeyString()]
		if !ok {
			return nil, tmsp.CodeType_BaseUnknownAccount, "Input account does not exist"
		}
		if appendTx {
			if acc.Sequence != input.Sequence {
				return nil, tmsp.CodeType_BaseBadNonce, "Invalid sequence"
			}
		} else {
			if acc.Sequence > input.Sequence {
				return nil, tmsp.CodeType_BaseBadNonce, "Invalid sequence (too low)"
			}
		}
		if acc.Balance < input.Amount {
			return nil, tmsp.CodeType_BaseInsufficientFunds, "Insufficient funds"
		}
		// Good!
		acc.Sequence++
		acc.Balance -= input.Amount
		accs = append(accs, acc)
	}
	// Add to outputs
	for _, output := range outputs {
		var acc, ok = accMap[output.PubKey.KeyString()]
		if !ok {
			// Create new account if it doesn't already exist.
			acc = types.PubAccount{
				PubKey: output.PubKey,
				Account: types.Account{
					Balance: output.Amount,
				},
			}
			accMap[output.PubKey.KeyString()] = acc
			accs = append(accs, acc)
		} else {
			// Good!
			if (acc.Balance + output.Amount) < acc.Balance {
				return nil, tmsp.CodeType_InternalError, "Output balance overflow in runTx"
			}
			acc.Balance += output.Amount
			accs = append(accs, acc)
		}
	}
	return accs, tmsp.CodeType_OK, ""
}

func runSendTx(tx types.Tx, accMap map[string]types.PubAccount, appendTx bool) (accs []types.PubAccount, code tmsp.CodeType, errStr string) {
	return processInputsOutputs(tx, accMap, appendTx)
}

func runGovTx(tx *types.GovTx, accMap map[string]types.PubAccount, appendTx bool) (accs []types.PubAccount, code tmsp.CodeType, errStr string) {
	accs, code, errStr = processInputsOutputs(tx, accMap, appendTx)
	// XXX run GovTx
	return
}

//----------------------------------------

func loadAccounts(eyesCli *eyes.Client, pubKeys []crypto.PubKey) (accMap map[string]types.PubAccount) {
	accMap = make(map[string]types.PubAccount, len(pubKeys))
	for _, pubKey := range pubKeys {
		keyString := pubKey.KeyString()
		accBytes, err := eyesCli.GetSync([]byte(keyString))
		if err != nil {
			panic("Error loading account: " + err.Error())
		}
		if len(accBytes) == 0 {
			continue
		}
		var acc types.Account
		err = wire.ReadBinaryBytes(accBytes, &acc)
		if err != nil {
			panic("Error reading account: " + err.Error())
		}
		accMap[keyString] = types.PubAccount{
			Account: acc,
			PubKey:  pubKey,
		}
	}
	return
}

// NOTE: accs must be stored in deterministic order.
func storeAccounts(eyesCli *eyes.Client, accs []types.PubAccount) {
	fmt.Println("STORE ACCOUNTS", accs)
	for _, acc := range accs {
		accBytes := wire.BinaryBytes(acc.Account)
		err := eyesCli.SetSync([]byte(acc.PubKey.KeyString()), accBytes)
		if err != nil {
			panic("Error storing account: " + err.Error())
		}
	}
}

//----------------------------------------

func allPubKeys(tx types.Tx) (pubKeys []crypto.PubKey) {
	inputs := tx.GetInputs()
	outputs := tx.GetOutputs()
	pubKeys = make([]crypto.PubKey, 0, len(inputs)+len(outputs))
	for _, input := range inputs {
		pubKeys = append(pubKeys, input.PubKey)
	}
	for _, output := range outputs {
		pubKeys = append(pubKeys, output.PubKey)
	}
	return pubKeys
}
