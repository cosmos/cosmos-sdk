package rest

import (
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

type (
	QueryNextSequenceRecv struct {
		Height int64  `json:"height"`
		Result uint64 `json:"result"`
	}

	PostTransfer struct {
		Msgs       []types.MsgTransfer `json:"msg" yaml:"msg"`
		Fee        auth.StdFee         `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature `json:"signatures" yaml:"signatures"`
		Memo       string              `json:"memo" yaml:"memo"`
	}

	PostRecvPacket struct {
		Msgs       []types.MsgRecvPacket `json:"msg" yaml:"msg"`
		Fee        auth.StdFee           `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature   `json:"signatures" yaml:"signatures"`
		Memo       string                `json:"memo" yaml:"memo"`
	}
)
