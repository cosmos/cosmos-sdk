package types

import (
	fmt "fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate validates a consensus key rotation history entry and returns
// validated fields.
func (h ConsensusKeyRotationHistory) Validate() (sdk.ValAddress, sdk.ConsAddress, error) {
	valAddr, err := sdk.ValAddressFromBech32(h.ValidatorAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid consensus key rotation validator address %s: %w", h.ValidatorAddress, err)
	}

	oldConsAddr, err := sdk.ConsAddressFromBech32(h.OldConsensusAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid consensus key rotation old consensus address %s: %w", h.OldConsensusAddress, err)
	}

	if h.MaturityTime.IsZero() {
		return nil, nil, fmt.Errorf("consensus key rotation history for validator %s has zero maturity time", h.ValidatorAddress)
	}

	return valAddr, oldConsAddr, nil
}

// Validate validates a pending consensus key rotation entry and returns
// validated fields.
func (r PendingConsensusKeyRotation) Validate() (sdk.ValAddress, cryptotypes.PubKey, error) {
	valAddr, err := sdk.ValAddressFromBech32(r.ValidatorAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid pending consensus key rotation validator address %s: %w", r.ValidatorAddress, err)
	}

	if r.ApplyHeight < 0 {
		return nil, nil, fmt.Errorf("pending consensus key rotation for validator %s has negative apply height", r.ValidatorAddress)
	}

	if r.NewPubkey == nil {
		return nil, nil, fmt.Errorf("pending consensus key rotation for validator %s has nil new pubkey", r.ValidatorAddress)
	}
	newPubKey, ok := r.NewPubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, nil, fmt.Errorf("pending consensus key rotation for validator %s has invalid new pubkey type %T", r.ValidatorAddress, r.NewPubkey.GetCachedValue())
	}

	return valAddr, newPubKey, nil
}
