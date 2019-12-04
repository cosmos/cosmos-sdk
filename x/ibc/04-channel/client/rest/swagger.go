package rest

import (
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

type (
	QueryChannel struct {
		Height int64                 `json:"height"`
		Result types.ChannelResponse `json:"result"`
	}

	PostChannelOpenInit struct {
		Msgs       []types.MsgChannelOpenInit `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature        `json:"signatures" yaml:"signatures"`
		Memo       string                     `json:"memo" yaml:"memo"`
	}

	PostChannelOpenTry struct {
		Msgs       []types.MsgChannelOpenTry `json:"msg" yaml:"msg"`
		Fee        auth.StdFee               `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature       `json:"signatures" yaml:"signatures"`
		Memo       string                    `json:"memo" yaml:"memo"`
	}

	PostChannelOpenAck struct {
		Msgs       []types.MsgChannelOpenAck `json:"msg" yaml:"msg"`
		Fee        auth.StdFee               `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature       `json:"signatures" yaml:"signatures"`
		Memo       string                    `json:"memo" yaml:"memo"`
	}

	PostChannelOpenConfirm struct {
		Msgs       []types.MsgChannelOpenConfirm `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                   `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature           `json:"signatures" yaml:"signatures"`
		Memo       string                        `json:"memo" yaml:"memo"`
	}

	PostChannelCloseInit struct {
		Msgs       []types.MsgChannelCloseInit `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                 `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature         `json:"signatures" yaml:"signatures"`
		Memo       string                      `json:"memo" yaml:"memo"`
	}

	PostChannelCloseConfirm struct {
		Msgs       []types.MsgChannelCloseConfirm `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                    `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature            `json:"signatures" yaml:"signatures"`
		Memo       string                         `json:"memo" yaml:"memo"`
	}
)
