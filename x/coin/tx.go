package coin

import (
	"fmt"
)

/*func init() {
	sdk.TxMapper.
		RegisterImplementation(SendTx{}, TypeSend, ByteSend).
		RegisterImplementation(CreditTx{}, TypeCredit, ByteCredit)
}*/

// we reserve the 0x20-0x3f range for standard modules
const (
	NameCoin = "coin"

	ByteSend   = 0x20
	TypeSend   = NameCoin + "/send"
	ByteCredit = 0x21
	TypeCredit = NameCoin + "/credit"
)

//-----------------------------------------------------------------------------

// TxInput - expected coin movement outputs, used with SendTx
type TxInput struct {
	Address Actor `json:"address"`
	Coins   Coins `json:"coins"`
}

// ValidateBasic - validate transaction input
func (txIn TxInput) ValidateBasic() error {
	if txIn.Address.App == "" {
		return ErrInvalidAddress()
	}
	// TODO: knowledge of app-specific codings?
	if len(txIn.Address.Address) == 0 {
		return ErrInvalidAddress()
	}
	if !txIn.Coins.IsValid() {
		return ErrInvalidCoins()
	}
	if !txIn.Coins.IsPositive() {
		return ErrInvalidCoins()
	}
	return nil
}

func (txIn TxInput) String() string {
	return fmt.Sprintf("TxInput{%v,%v}", txIn.Address, txIn.Coins)
}

// NewTxInput - create a transaction input, used with SendTx
func NewTxInput(addr Actor, coins Coins) TxInput {
	input := TxInput{
		Address: addr,
		Coins:   coins,
	}
	return input
}

//-----------------------------------------------------------------------------

// TxOutput - expected coin movement output, used with SendTx
type TxOutput struct {
	Address Actor `json:"address"`
	Coins   Coins `json:"coins"`
}

// ValidateBasic - validate transaction output
func (txOut TxOutput) ValidateBasic() error {
	if txOut.Address.App == "" {
		return ErrInvalidAddress()
	}
	// TODO: knowledge of app-specific codings?
	if len(txOut.Address.Address) == 0 {
		return ErrInvalidAddress()
	}
	if !txOut.Coins.IsValid() {
		return ErrInvalidCoins()
	}
	if !txOut.Coins.IsPositive() {
		return ErrInvalidCoins()
	}
	return nil
}

func (txOut TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%X,%v}", txOut.Address, txOut.Coins)
}

// NewTxOutput - create a transaction output, used with SendTx
func NewTxOutput(addr Actor, coins Coins) TxOutput {
	output := TxOutput{
		Address: addr,
		Coins:   coins,
	}
	return output
}

//-----------------------------------------------------------------------------

// SendTx - high level transaction of the coin module
// Satisfies: TxInner
type SendTx struct {
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

// var _ types.Tx = NewSendTx(nil, nil)

// NewSendTx - construct arbitrary multi-in, multi-out sendtx
func NewSendTx(in []TxInput, out []TxOutput) SendTx { // types.Tx {
	return SendTx{Inputs: in, Outputs: out}
}

// NewSendOneTx is a helper for the standard (?) case where there is exactly
// one sender and one recipient
func NewSendOneTx(sender, recipient Actor, amount Coins) SendTx {
	in := []TxInput{{Address: sender, Coins: amount}}
	out := []TxOutput{{Address: recipient, Coins: amount}}
	return SendTx{Inputs: in, Outputs: out}
}

// ValidateBasic - validate the send transaction
func (tx SendTx) ValidateBasic() error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(tx.Inputs) == 0 {
		return ErrNoInputs()
	}
	if len(tx.Outputs) == 0 {
		return ErrNoOutputs()
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut Coins
	for _, in := range tx.Inputs {
		if err := in.ValidateBasic(); err != nil {
			return err
		}
		totalIn = totalIn.Plus(in.Coins)
	}
	for _, out := range tx.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return err
		}
		totalOut = totalOut.Plus(out.Coins)
	}
	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return ErrInvalidCoins()
	}
	return nil
}

func (tx SendTx) String() string {
	return fmt.Sprintf("SendTx{%v->%v}", tx.Inputs, tx.Outputs)
}

//-----------------------------------------------------------------------------

// CreditTx - this allows a special issuer to give an account credit
// Satisfies: TxInner
type CreditTx struct {
	Debitor Actor `json:"debitor"`
	// Credit is the amount to change the credit...
	// This may be negative to remove some over-issued credit,
	// but can never bring the credit or the balance to negative
	Credit Coins `json:"credit"`
}

// NewCreditTx - modify the credit granted to a given account
func NewCreditTx(debitor Actor, credit Coins) CreditTx {
	return CreditTx{Debitor: debitor, Credit: credit}
}

// ValidateBasic - used to satisfy TxInner
func (tx CreditTx) ValidateBasic() error {
	return nil
}
