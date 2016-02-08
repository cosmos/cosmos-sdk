package app

import (
	"fmt"
	"github.com/tendermint/blackstar/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
	tmsp "github.com/tendermint/tmsp/types"
)

const version = "0.1"
const maxTxSize = 10240

type Blackstar struct {
	eyesCli *eyes.MerkleEyesClient
}

func NewBlackstar(eyesCli *eyes.MerkleEyesClient) *Blackstar {
	return &Blackstar{
		eyesCli: eyesCli,
	}
}

func (app *Blackstar) Info() string {
	return "Blackstar v" + version
}

func (app *Blackstar) SetOption(key string, value string) (log string) {
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

func (app *Blackstar) AppendTx(txBytes []byte) (code tmsp.CodeType, result []byte, log string) {
	if len(txBytes) > maxTxSize {
		return tmsp.CodeType_EncodingError, nil, "Tx size exceeds maximum"
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return tmsp.CodeType_EncodingError, nil, "Error decoding tx: " + err.Error()
	}
	// Validate tx
	code, errStr := validateTx(tx)
	if errStr != "" {
		return code, nil, "Error validating tx: " + errStr
	}
	// Load accounts
	accMap := loadAccounts(app.eyesCli, allPubKeys(tx))
	// Execute tx
	accs, code, errStr := execTx(tx, accMap, false)
	if errStr != "" {
		return code, nil, "Error executing tx: " + errStr
	}
	// Store accounts
	storeAccounts(app.eyesCli, accs)
	return tmsp.CodeType_OK, nil, "Success"
}

func (app *Blackstar) CheckTx(txBytes []byte) (code tmsp.CodeType, result []byte, log string) {
	if len(txBytes) > maxTxSize {
		return tmsp.CodeType_EncodingError, nil, "Tx size exceeds maximum"
	}
	// Decode tx
	var tx types.Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return tmsp.CodeType_EncodingError, nil, "Error decoding tx: " + err.Error()
	}
	// Validate tx
	code, errStr := validateTx(tx)
	if errStr != "" {
		return code, nil, "Error validating tx: " + errStr
	}
	// Load accounts
	accMap := loadAccounts(app.eyesCli, allPubKeys(tx))
	// Execute tx
	_, code, errStr = execTx(tx, accMap, false)
	if errStr != "" {
		return code, nil, "Error (mock) executing tx: " + errStr
	}
	return tmsp.CodeType_OK, nil, "Success"
}

func (app *Blackstar) Query(query []byte) (code tmsp.CodeType, result []byte, log string) {
	return tmsp.CodeType_OK, nil, ""
	value, err := app.eyesCli.GetSync(query)
	if err != nil {
		panic("Error making query: " + err.Error())
	}
	return tmsp.CodeType_OK, value, "Success"
}

func (app *Blackstar) GetHash() (hash []byte, log string) {
	hash, log, err := app.eyesCli.GetHashSync()
	if err != nil {
		panic("Error getting hash: " + err.Error())
	}
	return hash, "Success"
}

//----------------------------------------

func validateTx(tx types.Tx) (code tmsp.CodeType, errStr string) {
	if len(tx.Inputs) == 0 {
		return tmsp.CodeType_EncodingError, "Tx.Inputs length cannot be 0"
	}
	seenPubKeys := map[string]bool{}
	signBytes := txSignBytes(tx)
	for _, input := range tx.Inputs {
		code, errStr = validateInput(input, signBytes)
		if errStr != "" {
			return
		}
		keyString := input.PubKey.KeyString()
		if seenPubKeys[keyString] {
			return tmsp.CodeType_EncodingError, "Duplicate input pubKey"
		}
		seenPubKeys[keyString] = true
	}
	for _, output := range tx.Outputs {
		code, errStr = validateOutput(output)
		if errStr != "" {
			return
		}
		keyString := output.PubKey.KeyString()
		if seenPubKeys[keyString] {
			return tmsp.CodeType_EncodingError, "Duplicate output pubKey"
		}
		seenPubKeys[keyString] = true
	}
	sumInputs, overflow := sumAmounts(tx.Inputs, nil, 0)
	if overflow {
		return tmsp.CodeType_EncodingError, "Input amount overflow"
	}
	sumOutputsPlus, overflow := sumAmounts(nil, tx.Outputs, len(tx.Inputs)+len(tx.Outputs))
	if overflow {
		return tmsp.CodeType_EncodingError, "Output amount overflow"
	}
	if sumInputs < sumOutputsPlus {
		return tmsp.CodeType_InsufficientFees, "Insufficient fees"
	}
	return tmsp.CodeType_OK, ""
}

func txSignBytes(tx types.Tx) []byte {
	sigs := make([]crypto.Signature, len(tx.Inputs))
	for i, input := range tx.Inputs {
		sigs[i] = input.Signature
		input.Signature = nil
		tx.Inputs[i] = input
	}
	signBytes := wire.BinaryBytes(tx)
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = sigs[i]
	}
	return signBytes
}

func validateInput(input types.Input, signBytes []byte) (code tmsp.CodeType, errStr string) {
	if input.Amount == 0 {
		return tmsp.CodeType_EncodingError, "Input amount cannot be zero"
	}
	if input.PubKey == nil {
		return tmsp.CodeType_EncodingError, "Input pubKey cannot be nil"
	}
	if !input.PubKey.VerifyBytes(signBytes, input.Signature) {
		return tmsp.CodeType_Unauthorized, "Invalid signature"
	}
	return tmsp.CodeType_OK, ""
}

func validateOutput(output types.Output) (code tmsp.CodeType, errStr string) {
	if output.Amount == 0 {
		return tmsp.CodeType_EncodingError, "Output amount cannot be zero"
	}
	if output.PubKey == nil {
		return tmsp.CodeType_EncodingError, "Output pubKey cannot be nil"
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

func allPubKeys(tx types.Tx) (pubKeys []crypto.PubKey) {
	pubKeys = make([]crypto.PubKey, 0, len(tx.Inputs)+len(tx.Outputs))
	for _, input := range tx.Inputs {
		pubKeys = append(pubKeys, input.PubKey)
	}
	for _, output := range tx.Outputs {
		pubKeys = append(pubKeys, output.PubKey)
	}
	return pubKeys
}

// Returns accounts in order of types.Tx inputs and outputs
// appendTx: true if this is for AppendTx.
// TODO: create more intelligent sequence-checking.  Current impl is just for a throughput demo.
func execTx(tx types.Tx, accMap map[string]types.PubAccount, appendTx bool) (accs []types.PubAccount, code tmsp.CodeType, errStr string) {
	accs = make([]types.PubAccount, 0, len(tx.Inputs)+len(tx.Outputs))
	// Deduct from inputs
	for _, input := range tx.Inputs {
		var acc, ok = accMap[input.PubKey.KeyString()]
		if !ok {
			return nil, tmsp.CodeType_UnknownAccount, "Input account does not exist"
		}
		if appendTx {
			if acc.Sequence != input.Sequence {
				return nil, tmsp.CodeType_BadNonce, "Invalid sequence"
			}
		} else {
			if acc.Sequence > input.Sequence {
				return nil, tmsp.CodeType_BadNonce, "Invalid sequence (too low)"
			}
		}
		if acc.Balance < input.Amount {
			return nil, tmsp.CodeType_InsufficientFunds, "Insufficient funds"
		}
		// Good!
		acc.Sequence++
		acc.Balance -= input.Amount
		accs = append(accs, acc)
	}
	// Add to outputs
	for _, output := range tx.Outputs {
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
				return nil, tmsp.CodeType_InternalError, "Output balance overflow in execTx"
			}
			acc.Balance += output.Amount
			accs = append(accs, acc)
		}
	}
	return accs, tmsp.CodeType_OK, ""
}

//----------------------------------------

func loadAccounts(eyesCli *eyes.MerkleEyesClient, pubKeys []crypto.PubKey) (accMap map[string]types.PubAccount) {
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
func storeAccounts(eyesCli *eyes.MerkleEyesClient, accs []types.PubAccount) {
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
