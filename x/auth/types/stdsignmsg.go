package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StdSignMsg is a convenience structure for passing along
// a Msg with the other requirements for a StdSignDoc before
// it is signed. For use in the CLI.
type StdSignMsg struct {
	ChainID       string    `json:"chain_id" yaml:"chain_id"`
	AccountNumber uint64    `json:"account_number" yaml:"account_number"`
	Sequence      uint64    `json:"sequence" yaml:"sequence"`
	Fee           StdFee    `json:"fee" yaml:"fee"`
	Msgs          []sdk.Msg `json:"msgs" yaml:"msgs"`
	Memo          string    `json:"memo" yaml:"memo"`
}

// get message bytes
func (msg StdSignMsg) Bytes() []byte {
	return StdSignBytes(msg.ChainID, msg.AccountNumber, msg.Sequence, msg.Fee, msg.Msgs, msg.Memo)
}
