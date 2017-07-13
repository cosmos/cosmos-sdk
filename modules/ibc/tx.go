package ibc

import (
	abci "github.com/tendermint/abci/types"
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
	BytePacketCreate  = byte(0x32)
	BytePacketPost    = byte(0x33)

	TypeRegisterChain = NameIBC + "/register"
	TypeUpdateChain   = NameIBC + "/update"
	TypePacketCreate  = NameIBC + "/create"
	TypePacketPost    = NameIBC + "/post"

	IBCCodeEncodingError       = abci.CodeType(1001)
	IBCCodeChainAlreadyExists  = abci.CodeType(1002)
	IBCCodePacketAlreadyExists = abci.CodeType(1003)
	IBCCodeUnknownHeight       = abci.CodeType(1004)
	IBCCodeInvalidCommit       = abci.CodeType(1005)
	IBCCodeInvalidProof        = abci.CodeType(1006)
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(RegisterChainTx{}, TypeRegisterChain, ByteRegisterChain).
		RegisterImplementation(UpdateChainTx{}, TypeUpdateChain, ByteUpdateChain).
		RegisterImplementation(PacketCreateTx{}, TypePacketCreate, BytePacketCreate).
		RegisterImplementation(PacketPostTx{}, TypePacketPost, BytePacketPost)
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
	return r.Seed.ValidateBasic(r.ChainID())
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
	return u.Seed.ValidateBasic(u.ChainID())
}

// PacketCreateTx is meant to be called by IPC, another module...
//
// this is the tx that will be sent to another app and the permissions it
// comes with (which must be a subset of the permissions on the current tx)
//
// TODO: how to control who can create packets (can I just signed create packet?)
type PacketCreateTx struct {
	DestChain   string           `json:"dest_chain"`
	Permissions []basecoin.Actor `json:"permissions"`
	Tx          basecoin.Tx      `json:"tx"`
}

// ValidateBasic makes sure this is consistent - used to satisfy TxInner
func (p PacketCreateTx) ValidateBasic() error {
	if p.DestChain == "" {
		return errors.ErrNoChain()
	}
	// if len(p.Permissions) == 0 {
	//   return ErrNoPermissions()
	// }
	return nil
}

// Wrap - used to satisfy TxInner
func (p PacketCreateTx) Wrap() basecoin.Tx {
	return basecoin.Tx{p}
}

// PacketPostTx takes a wrapped packet from another chain and
// TODO!!!
type PacketPostTx struct {
	FromChainID     string // The immediate source of the packet, not always Packet.SrcChainID
	FromChainHeight uint64 // The block height in which Packet was committed, to check Proof
	Proof           *merkle.IAVLProof
	// Packet
}

// ValidateBasic makes sure this is consistent - used to satisfy TxInner
func (p PacketPostTx) ValidateBasic() error {
	// TODO
	return nil
}

// Wrap - used to satisfy TxInner
func (p PacketPostTx) Wrap() basecoin.Tx {
	return basecoin.Tx{p}
}

// proof := tx.Proof
// if proof == nil {
//   sm.res.Code = IBCCodeInvalidProof
//   sm.res.Log = "Proof is nil"
//   return
// }
// packetBytes := wire.BinaryBytes(packet)

// // Make sure packet's proof matches given (packet, key, blockhash)
// ok := proof.Verify(packetKeyEgress, packetBytes, header.AppHash)
// if !ok {
//   sm.res.Code = IBCCodeInvalidProof
//   sm.res.Log = fmt.Sprintf("Proof is invalid. key: %s; packetByes %X; header %v; proof %v", packetKeyEgress, packetBytes, header, proof)
//   return
// }

// // Execute payload
// switch payload := packet.Payload.(type) {
// case DataPayload:
//   // do nothing
// case CoinsPayload:
//   // Add coins to destination account
//   acc := types.GetAccount(sm.store, payload.Address)
//   if acc == nil {
//     acc = &types.Account{}
//   }
//   acc.Balance = acc.Balance.Plus(payload.Coins)
//   types.SetAccount(sm.store, payload.Address, acc)
// }
