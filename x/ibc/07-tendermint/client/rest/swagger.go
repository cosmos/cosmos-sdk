package rest

import (
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// nolint
type (
	PostCreateClient struct {
		Msgs       []ibctmtypes.MsgCreateClient `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                  `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature          `json:"signatures" yaml:"signatures"`
		Memo       string                       `json:"memo" yaml:"memo"`
	}

	PostUpdateClient struct {
		Msgs       []ibctmtypes.MsgUpdateClient `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                  `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature          `json:"signatures" yaml:"signatures"`
		Memo       string                       `json:"memo" yaml:"memo"`
	}

	PostSubmitMisbehaviour struct {
		Msgs       []evidence.MsgSubmitEvidence `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                  `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature          `json:"signatures" yaml:"signatures"`
		Memo       string                       `json:"memo" yaml:"memo"`
	}
)
