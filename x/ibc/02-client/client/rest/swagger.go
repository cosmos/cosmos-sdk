package rest

import (
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type (
	QueryConsensusState struct {
		Height int64                        `json:"height"`
		Result types.ConsensusStateResponse `json:"result"`
	}

	QueryHeader struct {
		Height int64             `json:"height"`
		Result tendermint.Header `json:"result"`
	}

	QueryClientState struct {
		Height int64               `json:"height"`
		Result types.StateResponse `json:"result"`
	}

	QueryRoot struct {
		Height int64              `json:"height"`
		Result types.RootResponse `json:"result"`
	}

	QueryNodeConsensusState struct {
		Height int64                     `json:"height"`
		Result tendermint.ConsensusState `json:"result"`
	}

	QueryPath struct {
		Height int64             `json:"height"`
		Result commitment.Prefix `json:"result"`
	}

	PostCreateClient struct {
		Msgs       []types.MsgCreateClient `json:"msg" yaml:"msg"`
		Fee        auth.StdFee             `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature     `json:"signatures" yaml:"signatures"`
		Memo       string                  `json:"memo" yaml:"memo"`
	}

	PostUpdateClient struct {
		Msgs       []types.MsgUpdateClient `json:"msg" yaml:"msg"`
		Fee        auth.StdFee             `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature     `json:"signatures" yaml:"signatures"`
		Memo       string                  `json:"memo" yaml:"memo"`
	}

	PostSubmitMisbehaviour struct {
		Msgs       []types.MsgSubmitMisbehaviour `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                   `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature           `json:"signatures" yaml:"signatures"`
		Memo       string                        `json:"memo" yaml:"memo"`
	}
)
