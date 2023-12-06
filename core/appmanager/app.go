package appmanager

import (
	"context"
	"time"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/core/validator"
)

type Header interface {
	Height() int64
	Time() time.Time
	GetChainID() string
	Hash() []byte
}

type CelestiaHeader struct {
	Header
	sequencers []byte
	proofs     []byte
	evidence   []byte
}

func (ch CelestiaHeader) Height() int64 {
	return 1
}

func (ch CelestiaHeader) Time() time.Time {
	return time.Now()
}

func (ch CelestiaHeader) GetChainID() string {
	return "celestia"
}

func (ch CelestiaHeader) Hash() []byte {
	return []byte("celestia")
}

func (ch CelestiaHeader) Sequencers() []byte {
	return ch.sequencers
}

func (ch CelestiaHeader) Proofs() []byte {
	return ch.proofs
}

func (ch CelestiaHeader) Evidence() []byte {
	return ch.evidence
}

type CometHeader struct {
	Header
	Evidence []comet.Evidence // Evidence misbehavior of the block
	// ValidatorsHash returns the hash of the validators
	// For Comet, it is the hash of the next validator set
	ValidatorsHash  []byte
	ProposerAddress []byte           // ProposerAddress is  the address of the block proposer
	LastCommit      comet.CommitInfo // DecidedLastCommit returns the last commit info
}

/*
// NewStakingModule(distributionmodule, slashingmodule)

func ValidatorSomething() error {

	evi := ctx.GetHeader[CometHeader].Evidence()

	return nil
}

*/

type App[H Header, T transaction.Tx] interface {
	ChainID() string
	AppVersion() (uint64, error)

	InitChain(context.Context, RequestInitChain) (ResponseInitChain, error)
	DeliverBlock(context.Context, RequestDeliverBlock[H, T]) (ResponseDeliverBlock, error)
}

type RequestDeliverBlock[H Header, T transaction.Tx] struct {
	Header H
	Txs    []T
}

type ResponseDeliverBlock struct {
	Apphash          []byte
	ValidatorUpdates []validator.Update
	TxResults        []TxResult
	Events           []event.Event
}

type RequestInitChain struct {
	Time          time.Time
	ChainId       string
	Validators    []validator.Update
	AppStateBytes []byte
	InitialHeight int64
}

type ResponseInitChain struct {
	Validators []validator.Update
	AppHash    []byte
}

type TxResult struct {
	GasWanted int64
	GasUsed   int64
	Log       string
	Data      string
	Events    event.Event
}
