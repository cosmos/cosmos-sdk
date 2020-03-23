package rest

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

type (
	QueryNextSequenceRecv struct {
		Height int64  `json:"height"`
		Result uint64 `json:"result"`
	}

	PostTransfer struct {
		Msgs       []types.MsgTransfer      `json:"msg" yaml:"msg"`
		Fee        authtypes.StdFee         `json:"fee" yaml:"fee"`
		Signatures []authtypes.StdSignature `json:"signatures" yaml:"signatures"`
		Memo       string                   `json:"memo" yaml:"memo"`
	}
)
