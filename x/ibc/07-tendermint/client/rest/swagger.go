package rest

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// nolint
type (
	PostCreateClient struct {
		Msgs       []ibctmtypes.MsgCreateClient `json:"msg" yaml:"msg"`
		Fee        authtypes.StdFee             `json:"fee" yaml:"fee"`
		Signatures []authtypes.StdSignature     `json:"signatures" yaml:"signatures"`
		Memo       string                       `json:"memo" yaml:"memo"`
	}

	PostUpdateClient struct {
		Msgs       []ibctmtypes.MsgUpdateClient `json:"msg" yaml:"msg"`
		Fee        authtypes.StdFee             `json:"fee" yaml:"fee"`
		Signatures []authtypes.StdSignature     `json:"signatures" yaml:"signatures"`
		Memo       string                       `json:"memo" yaml:"memo"`
	}

	PostSubmitMisbehaviour struct {
		Msgs       []ibctmtypes.MsgSubmitClientMisbehaviour `json:"msg" yaml:"msg"`
		Fee        authtypes.StdFee                         `json:"fee" yaml:"fee"`
		Signatures []authtypes.StdSignature                 `json:"signatures" yaml:"signatures"`
		Memo       string                                   `json:"memo" yaml:"memo"`
	}
)
