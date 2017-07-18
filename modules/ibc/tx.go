package ibc

import (
	"github.com/tendermint/light-client/certifiers"
	merkle "github.com/tendermint/merkleeyes/iavl"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
)

// nolint
const (
	// 0x3? series for ibc
	ByteRegisterChain = byte(0x30)
	ByteUpdateChain   = byte(0x31)
	ByteCreatePacket  = byte(0x32)
	BytePostPacket    = byte(0x33)

	TypeRegisterChain = NameIBC + "/register"
	TypeUpdateChain   = NameIBC + "/update"
	TypeCreatePacket  = NameIBC + "/create"
	TypePostPacket    = NameIBC + "/post"
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(RegisterChainTx{}, TypeRegisterChain, ByteRegisterChain).
		RegisterImplementation(UpdateChainTx{}, TypeUpdateChain, ByteUpdateChain).
		RegisterImplementation(CreatePacketTx{}, TypeCreatePacket, ByteCreatePacket).
		RegisterImplementation(PostPacketTx{}, TypePostPacket, BytePostPacket)
}

// RegisterChainTx allows you to register a new chain on this blockchain
type RegisterChainTx struct {
	Seed certifiers.Seed `json:"seed"`
}

// ChainID helps get the chain this tx refers to
func (r RegisterChainTx) ChainID() string {
	return r.Seed.Header.ChainID
}

// ValidateBasic makes sure this is consistent, without checking the sigs
func (r RegisterChainTx) ValidateBasic() error {
	err := r.Seed.ValidateBasic(r.ChainID())
	if err != nil {
		err = ErrInvalidCommit(err)
	}
	return err
}

// Wrap - used to satisfy TxInner
func (r RegisterChainTx) Wrap() basecoin.Tx {
	return basecoin.Tx{r}
}

// UpdateChainTx updates the state of this chain
type UpdateChainTx struct {
	Seed certifiers.Seed `json:"seed"`
}

// ChainID helps get the chain this tx refers to
func (u UpdateChainTx) ChainID() string {
	return u.Seed.Header.ChainID
}

// ValidateBasic makes sure this is consistent, without checking the sigs
func (u UpdateChainTx) ValidateBasic() error {
	err := u.Seed.ValidateBasic(u.ChainID())
	if err != nil {
		err = ErrInvalidCommit(err)
	}
	return err
}

// Wrap - used to satisfy TxInner
func (u UpdateChainTx) Wrap() basecoin.Tx {
	return basecoin.Tx{u}
}

// CreatePacketTx is meant to be called by IPC, another module...
//
// this is the tx that will be sent to another app and the permissions it
// comes with (which must be a subset of the permissions on the current tx)
//
// If must have the special `AllowIBC` permission from the app
// that can send this packet (so only coins can request SendTx packet)
type CreatePacketTx struct {
	DestChain   string           `json:"dest_chain"`
	Permissions []basecoin.Actor `json:"permissions"`
	Tx          basecoin.Tx      `json:"tx"`
}

// ValidateBasic makes sure this is consistent - used to satisfy TxInner
func (p CreatePacketTx) ValidateBasic() error {
	if p.DestChain == "" {
		return errors.ErrNoChain()
	}
	// if len(p.Permissions) == 0 {
	//   return ErrNoPermissions()
	// }
	return nil
}

// Wrap - used to satisfy TxInner
func (p CreatePacketTx) Wrap() basecoin.Tx {
	return basecoin.Tx{p}
}

// PostPacketTx takes a wrapped packet from another chain and
// TODO!!!
// also think... which chains can relay packets???
// right now, enforce that these packets are only sent directly,
// not routed over the hub.  add routing later.
type PostPacketTx struct {
	// make sure we have this header...
	FromChainID     string // The immediate source of the packet, not always Packet.SrcChainID
	FromChainHeight uint64 // The block height in which Packet was committed, to check Proof
	// this proof must match the header and the packet.Bytes()
	Proof  *merkle.IAVLProof
	Key    []byte
	Packet Packet
}

// ValidateBasic makes sure this is consistent - used to satisfy TxInner
func (p PostPacketTx) ValidateBasic() error {
	// TODO
	return nil
}

// Wrap - used to satisfy TxInner
func (p PostPacketTx) Wrap() basecoin.Tx {
	return basecoin.Tx{p}
}
