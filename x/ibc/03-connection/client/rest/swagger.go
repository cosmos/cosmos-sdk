package rest

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

type (
	QueryConnection struct {
		Height int64                    `json:"height"`
		Result types.ConnectionResponse `json:"result"`
	}

	QueryClientConnections struct {
		Height int64                           `json:"height"`
		Result types.ClientConnectionsResponse `json:"result"`
	}

	PostConnectionOpenInit struct {
		Msgs       []types.MsgConnectionOpenInit `json:"msg" yaml:"msg"`
		Fee        authtypes.StdFee                   `json:"fee" yaml:"fee"`
		Signatures []authtypes.StdSignature           `json:"signatures" yaml:"signatures"`
		Memo       string                        `json:"memo" yaml:"memo"`
	}

	PostConnectionOpenTry struct {
		Msgs       []types.MsgConnectionOpenTry `json:"msg" yaml:"msg"`
		Fee        authtypes.StdFee                  `json:"fee" yaml:"fee"`
		Signatures []authtypes.StdSignature          `json:"signatures" yaml:"signatures"`
		Memo       string                       `json:"memo" yaml:"memo"`
	}

	PostConnectionOpenAck struct {
		Msgs       []types.MsgConnectionOpenAck `json:"msg" yaml:"msg"`
		Fee        authtypes.StdFee                  `json:"fee" yaml:"fee"`
		Signatures []authtypes.StdSignature          `json:"signatures" yaml:"signatures"`
		Memo       string                       `json:"memo" yaml:"memo"`
	}

	PostConnectionOpenConfirm struct {
		Msgs       []types.MsgConnectionOpenConfirm `json:"msg" yaml:"msg"`
		Fee        authtypes.StdFee                      `json:"fee" yaml:"fee"`
		Signatures []authtypes.StdSignature              `json:"signatures" yaml:"signatures"`
		Memo       string                           `json:"memo" yaml:"memo"`
	}
)
