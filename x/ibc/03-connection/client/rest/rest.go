package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	RestConnectionID = "connection-id"
	RestClientID     = "client-id"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	registerQueryRoutes(clientCtx, r)
	registerTxRoutes(clientCtx, r)
}

// ConnectionOpenInitReq defines the properties of a connection open init request's body.
type ConnectionOpenInitReq struct {
	BaseReq                  rest.BaseReq                 `json:"base_req" yaml:"base_req"`
	ConnectionID             string                       `json:"connection_id" yaml:"connection_id"`
	ClientID                 string                       `json:"client_id" yaml:"client_id"`
	CounterpartyClientID     string                       `json:"counterparty_client_id" yaml:"counterparty_client_id"`
	CounterpartyConnectionID string                       `json:"counterparty_connection_id" yaml:"counterparty_connection_id"`
	CounterpartyPrefix       commitmenttypes.MerklePrefix `json:"counterparty_prefix" yaml:"counterparty_prefix"`
}

// ConnectionOpenTryReq defines the properties of a connection open try request's body.
type ConnectionOpenTryReq struct {
	BaseReq                  rest.BaseReq                 `json:"base_req" yaml:"base_req"`
	ConnectionID             string                       `json:"connection_id" yaml:"connection_id"`
	ClientID                 string                       `json:"client_id" yaml:"client_id"`
	CounterpartyClientID     string                       `json:"counterparty_client_id" yaml:"counterparty_client_id"`
	CounterpartyConnectionID string                       `json:"counterparty_connection_id" yaml:"counterparty_connection_id"`
	CounterpartyPrefix       commitmenttypes.MerklePrefix `json:"counterparty_prefix" yaml:"counterparty_prefix"`
	CounterpartyVersions     []string                     `json:"counterparty_versions" yaml:"counterparty_versions"`
	ProofInit                []byte                       `json:"proof_init" yaml:"proof_init"`
	ProofConsensus           []byte                       `json:"proof_consensus" yaml:"proof_consensus"`
	ProofHeight              uint64                       `json:"proof_height" yaml:"proof_height"`
	ConsensusHeight          uint64                       `json:"consensus_height" yaml:"consensus_height"`
}

// ConnectionOpenAckReq defines the properties of a connection open ack request's body.
type ConnectionOpenAckReq struct {
	BaseReq         rest.BaseReq `json:"base_req" yaml:"base_req"`
	ProofTry        []byte       `json:"proof_try" yaml:"proof_try"`
	ProofConsensus  []byte       `json:"proof_consensus" yaml:"proof_consensus"`
	ProofHeight     uint64       `json:"proof_height" yaml:"proof_height"`
	ConsensusHeight uint64       `json:"consensus_height" yaml:"consensus_height"`
	Version         string       `json:"version" yaml:"version"`
}

// ConnectionOpenConfirmReq defines the properties of a connection open confirm request's body.
type ConnectionOpenConfirmReq struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	ProofAck    []byte       `json:"proof_ack" yaml:"proof_ack"`
	ProofHeight uint64       `json:"proof_height" yaml:"proof_height"`
}
