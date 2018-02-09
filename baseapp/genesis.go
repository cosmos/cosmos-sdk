package baseapp

import (
	"encoding/json"
	"io/ioutil"
	"time"

	crypto "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

// TODO this is dup code from tendermint-core to avoid dep issues
// should probably remove from both SDK / Tendermint and move to tmlibs
//  ^^^^^^^^^^^^^ This is my preference to avoid DEP hell
// or sync up versioning and just reference tendermint

// GenesisDoc defines the initial conditions for a tendermint blockchain, in particular its validator set.
type GenesisDoc struct {
	GenesisTime     time.Time          `json:"genesis_time"`
	ChainID         string             `json:"chain_id"`
	ConsensusParams *ConsensusParams   `json:"consensus_params,omitempty"`
	Validators      []GenesisValidator `json:"validators"`
	AppHash         cmn.HexBytes       `json:"app_hash"`
	AppState        json.RawMessage    `json:"app_state,omitempty"`
}

//nolint TODO remove
type ConsensusParams struct {
	BlockSize      `json:"block_size_params"`
	TxSize         `json:"tx_size_params"`
	BlockGossip    `json:"block_gossip_params"`
	EvidenceParams `json:"evidence_params"`
}
type GenesisValidator struct {
	PubKey crypto.PubKey `json:"pub_key"`
	Power  int64         `json:"power"`
	Name   string        `json:"name"`
}
type BlockSize struct {
	MaxBytes int   `json:"max_bytes"` // NOTE: must not be 0 nor greater than 100MB
	MaxTxs   int   `json:"max_txs"`
	MaxGas   int64 `json:"max_gas"`
}
type TxSize struct {
	MaxBytes int   `json:"max_bytes"`
	MaxGas   int64 `json:"max_gas"`
}
type BlockGossip struct {
	BlockPartSizeBytes int `json:"block_part_size_bytes"` // NOTE: must not be 0
}
type EvidenceParams struct {
	MaxAge int64 `json:"max_age"` // only accept new evidence more recent than this
}

// GenesisDocFromFile reads JSON data from a file and unmarshalls it into a GenesisDoc.
func GenesisDocFromFile(genDocFile string) (*GenesisDoc, error) {
	if genDocFile == "" {
		var g GenesisDoc
		return &g, nil
	}
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, err
	}
	genDoc := GenesisDoc{}
	err = json.Unmarshal(jsonBlob, &genDoc)
	if err != nil {
		return nil, err
	}
	return &genDoc, nil
}
