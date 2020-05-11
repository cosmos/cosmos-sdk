package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StdSignMsg is a convenience structure for passing along
// a Msg with the other requirements for a StdSignDoc before
// it is signed. For use in the CLI.
// XXX This is confusing, as it isn't a "Msg" which implies sdk.Msg.
// The only difference between this and StdSignDoc is that StdFee is
// enforced (which is fine, but why is Fee a RawMessage in StdSignDoc anyways?
// it could allow us to support non-Coins fees, for instance, or other fee
// mechanics than what is default, but those could be supported via protobuf
// field additions to StdFee anyways, so this approach seems better.
// But still, otherwise, why do we need to have Msgs that isn't just RawMessage?
// Sometimes you want to write tools where the Msg cannot be parsed
// (see multisign tool).
// IOW, Why not just write a convenience function for Msgs -> json.RawMessage?
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
