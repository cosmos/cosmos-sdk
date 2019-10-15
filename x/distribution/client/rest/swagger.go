package rest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.
//
//nolint:deadcode,unused
type (
	coinsReturn struct {
		Height int64     `json:"height"`
		Result sdk.Coins `json:"result"`
	}
	// delegatorWithdrawalAddr helps generate documentation for delegatorWithdrawalAddrHandlerFn
	delegatorWithdrawalAddr struct {
		Height int64       `json:"height"`
		Result sdk.Address `json:"result"`
	}

	// validatorInfo helps generate documentation for validatorInfoHandlerFn
	validatorInfo struct {
		Height int64             `json:"height"`
		Result ValidatorDistInfo `json:"result"`
	}

	// params helps generate documentation for paramsHandlerFn
	params struct {
		Height int64               `json:"height"`
		Result common.PrettyParams `json:"result"`
	}

	// >----------------
	// setDelegatorWithdrawalAddr is used to generate documentation for setDelegatorWithdrawalAddrHandllerFn
	setDelegatorWithdrawalAddr struct {
		Msgs       []types.MsgSetWithdrawAddress `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                   `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature           `json:"signatures" yaml:"signatures"`
		Memo       string                        `json:"memo" yaml:"memo"`
	}

	// withdrawDelegatorReward is used to generate documentation for withdrawDelegatorRewardHandlerFn
	withdrawDelegatorReward struct {
		Msgs       []types.MsgWithdrawDelegatorReward `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                        `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature                `json:"signatures" yaml:"signatures"`
		Memo       string                             `json:"memo" yaml:"memo"`
	}

	// withdrawValidatorRewards is used to generate documentation for withdrawValidatorRewardsHandllerFn
	withdrawValidatorRewards struct {
		Msgs       []types.MsgWithdrawValidatorCommission `json:"msg" yaml:"msg"`
		Fee        auth.StdFee                            `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature                    `json:"signatures" yaml:"signatures"`
		Memo       string                                 `json:"memo" yaml:"memo"`
	}
)
