package rest

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// nolint
type (
	QueryConsensusState struct {
		Height int64                        `json:"height"`
		Result types.ConsensusStateResponse `json:"result"`
	}

	QueryHeader struct {
		Height int64             `json:"height"`
		Result ibctmtypes.Header `json:"result"`
	}

	QueryClientState struct {
		Height int64               `json:"height"`
		Result types.StateResponse `json:"result"`
	}

	QueryNodeConsensusState struct {
		Height int64                     `json:"height"`
		Result ibctmtypes.ConsensusState `json:"result"`
	}

	QueryPath struct {
		Height int64             `json:"height"`
		Result commitment.Prefix `json:"result"`
	}
)
