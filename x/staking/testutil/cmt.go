package testutil

import (
	"cosmossdk.io/math"
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	cmttypes "github.com/cometbft/cometbft/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetCmtConsPubKey gets the validator's public key as a cmtcrypto.PubKey.
func GetCmtConsPubKey(v types.Validator) (cmtcrypto.PubKey, error) {
	pk, err := v.ConsPubKey()
	if err != nil {
		return nil, err
	}

	return cryptocodec.ToTmPubKeyInterface(pk)
}

// ToCmtValidator casts an SDK validator to a CometBFT type Validator.
func ToCmtValidator(v types.Validator, r math.Int) (*cmttypes.Validator, error) {
	tmPk, err := GetCmtConsPubKey(v)
	if err != nil {
		return nil, err
	}

	return cmttypes.NewValidator(tmPk, v.ConsensusPower(r)), nil
}

// ToCmtValidators casts all validators to the corresponding CometBFT type.
func ToCmtValidators(v types.Validators, r math.Int) ([]*cmttypes.Validator, error) {
	validators := make([]*cmttypes.Validator, len(v))
	var err error
	for i, val := range v {
		validators[i], err = ToCmtValidator(val, r)
		if err != nil {
			return nil, err
		}
	}

	return validators, nil
}
