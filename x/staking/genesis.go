package staking

import (
	"fmt"

	cmttypes "github.com/cometbft/cometbft/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx sdk.Context, keeper *keeper.Keeper) (vals []cmttypes.GenesisValidator, returnErr error) {
	exportedInitialHeight := ctx.BlockHeight() + 1
	pendingRotations, err := keeper.PendingConsKeyRotations(ctx)
	if err != nil {
		return nil, err
	}

	err = keeper.IterateLastValidators(ctx, func(_ int64, validator types.ValidatorI) (stop bool) {
		valAddr, err := keeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
		if err != nil {
			returnErr = err
			return true
		}

		// if this validator has a pending rotation happening while we are
		// exporting the validators, we respect the apply height for the
		// rotation and export using the new key if it would be applied at a
		// height >= exportedInitialHeight.
		pk, err := pendingRotations.EffectiveKeyForGenesis(valAddr, validator, exportedInitialHeight)
		if err != nil {
			returnErr = err
			return true
		}
		cmtPk, err := cryptocodec.ToCmtPubKeyInterface(pk)
		if err != nil {
			returnErr = err
			return true
		}

		vals = append(vals, cmttypes.GenesisValidator{
			Address: sdk.ConsAddress(cmtPk.Address()).Bytes(),
			PubKey:  cmtPk,
			Power:   validator.GetConsensusPower(keeper.PowerReduction(ctx)),
			Name:    validator.GetMoniker(),
		})

		return false
	})
	if err != nil {
		return nil, err
	}

	return vals, returnErr
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants hold. (i.e. params in correct bounds, no duplicate validators)
func ValidateGenesis(data *types.GenesisState) error {
	if err := validateGenesisStateValidators(data.Validators); err != nil {
		return err
	}

	if err := validateGenesisStateConsKeyRotations(data); err != nil {
		return err
	}

	return data.Params.Validate()
}

func validateGenesisStateValidators(validators []types.Validator) error {
	addrMap := make(map[string]bool, len(validators))

	for i := range validators {
		val := validators[i]
		consPk, err := val.ConsPubKey()
		if err != nil {
			return err
		}

		strKey := string(consPk.Bytes())

		if _, ok := addrMap[strKey]; ok {
			consAddr, err := val.GetConsAddr()
			if err != nil {
				return err
			}
			return fmt.Errorf("duplicate validator in genesis state: moniker %v, address %v", val.Description.Moniker, consAddr)
		}

		if val.Jailed && val.IsBonded() {
			consAddr, err := val.GetConsAddr()
			if err != nil {
				return err
			}
			return fmt.Errorf("validator is bonded and jailed in genesis state: moniker %v, address %v", val.Description.Moniker, consAddr)
		}

		if val.DelegatorShares.IsZero() && !val.IsUnbonding() {
			return fmt.Errorf("bonded/unbonded genesis validator cannot have zero delegator shares, validator: %v", val)
		}

		addrMap[strKey] = true
	}

	return nil
}

// validateGenesisStateConsKeyRotations validates the imported key rotation
// queues against the same uniqueness and lock invariants enforced by the tx
// path.
func validateGenesisStateConsKeyRotations(data *types.GenesisState) error {
	validatorIndexes, err := buildGenesisValidatorIndex(data)
	if err != nil {
		return err
	}

	historyIndexes, err := validateConsKeyRotationHistory(data, validatorIndexes)
	if err != nil {
		return err
	}

	return validatePendingConsKeyRotations(data, validatorIndexes, historyIndexes)
}

// consKeyRotationValidatorIndexes tracks validator consensus keys that are live
// in genesis so pending rotations cannot target keys already in use.
type consKeyRotationValidatorIndexes struct {
	byValidatorAddress                 map[string]struct{}
	validatorAddressByConsensusAddress map[string]string
}

// consKeyRotationHistoryIndexes tracks rotation history entries that should
// continue locking old consensus keys during the unbonding window.
type consKeyRotationHistoryIndexes struct {
	byValidatorAddress                 map[string]struct{}
	byConsensusAddress                 map[string]struct{}
	consensusAddressByValidatorAddress map[string]string
}

// buildGenesisValidatorIndex creates a lookup of validators listed in the
// genesis.
func buildGenesisValidatorIndex(data *types.GenesisState) (consKeyRotationValidatorIndexes, error) {
	vals := data.GetValidators()
	indexes := consKeyRotationValidatorIndexes{
		byValidatorAddress:                 make(map[string]struct{}, len(vals)),
		validatorAddressByConsensusAddress: make(map[string]string, len(vals)),
	}

	for _, validator := range vals {
		// index the validator operator address so history and pending entries
		// cannot reference missing validators
		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			return consKeyRotationValidatorIndexes{}, fmt.Errorf("invalid validator address: %w", err)
		}
		indexes.byValidatorAddress[string(valAddr)] = struct{}{}

		// index the live consensus address so pending rotations cannot target
		// a key already assigned to a validator
		consPubKey, err := validator.ConsPubKey()
		if err != nil {
			return consKeyRotationValidatorIndexes{}, err
		}
		indexes.validatorAddressByConsensusAddress[string(consPubKey.Address())] = string(valAddr)
	}

	return indexes, nil
}

// validateConsKeyRotationHistory validates already applied rotations that are
// still inside the unbonding window and returns the old consensus key locks
// restored from genesis.
func validateConsKeyRotationHistory(
	data *types.GenesisState,
	validatorIndexes consKeyRotationValidatorIndexes,
) (consKeyRotationHistoryIndexes, error) {
	history := data.GetConsensusKeyRotationHistory()
	indexes := consKeyRotationHistoryIndexes{
		byValidatorAddress:                 make(map[string]struct{}, len(history)),
		byConsensusAddress:                 make(map[string]struct{}, len(history)),
		consensusAddressByValidatorAddress: make(map[string]string, len(history)),
	}

	for _, h := range history {
		// decode both addresses before using them as canonical map keys
		valAddr, oldConsAddr, err := h.Validate()
		if err != nil {
			return consKeyRotationHistoryIndexes{}, err
		}

		// a validator can only have one history record inside the unbonding
		// window, and history is only valid for validators present in genesis
		if _, found := indexes.byValidatorAddress[string(valAddr)]; found {
			return consKeyRotationHistoryIndexes{}, fmt.Errorf("duplicate consensus key rotation history for validator %s", h.ValidatorAddress)
		}
		if _, found := validatorIndexes.byValidatorAddress[string(valAddr)]; !found {
			return consKeyRotationHistoryIndexes{}, fmt.Errorf("consensus key rotation history for unknown validator %s", h.ValidatorAddress)
		}
		indexes.byValidatorAddress[string(valAddr)] = struct{}{}

		// a rotated-away consensus address can only be locked by one history
		// entry
		consAddrKey := string(oldConsAddr)
		if _, found := indexes.byConsensusAddress[consAddrKey]; found {
			return consKeyRotationHistoryIndexes{}, fmt.Errorf("duplicate consensus key rotation old consensus address %s", h.OldConsensusAddress)
		}
		indexes.byConsensusAddress[consAddrKey] = struct{}{}
		indexes.consensusAddressByValidatorAddress[string(valAddr)] = consAddrKey
	}

	return indexes, nil
}

// validatePendingConsKeyRotations validates deferred key swaps against live
// validator keys, rotation history locks, and duplicate pending targets.
func validatePendingConsKeyRotations(
	data *types.GenesisState,
	validatorIndexes consKeyRotationValidatorIndexes,
	historyIndexes consKeyRotationHistoryIndexes,
) error {
	pendingRotations := data.GetPendingConsensusKeyRotations()
	pendingByValidator := make(map[string]struct{}, len(pendingRotations))
	pendingNewConsAddrs := make(map[string]struct{}, len(pendingRotations))

	for _, rotation := range pendingRotations {
		// decode the validator address before using it as the pending map key
		valAddr, newPubKey, err := rotation.Validate()
		if err != nil {
			return err
		}

		// every pending rotation must correspond to a real validator and its
		// restored history marker
		if _, found := historyIndexes.byValidatorAddress[string(valAddr)]; !found {
			return fmt.Errorf("pending consensus key rotation for validator %s missing rotation history", rotation.ValidatorAddress)
		}
		if _, found := validatorIndexes.byValidatorAddress[string(valAddr)]; !found {
			return fmt.Errorf("pending consensus key rotation for unknown validator %s", rotation.ValidatorAddress)
		}

		// each validator can have at most one deferred key swap
		if _, found := pendingByValidator[string(valAddr)]; found {
			return fmt.Errorf("duplicate pending consensus key rotation for validator %s", rotation.ValidatorAddress)
		}
		pendingByValidator[string(valAddr)] = struct{}{}

		// the target consensus key must not be pending, live, or locked by
		// rotation history
		newConsAddrKey := string(newPubKey.Address())
		if _, found := pendingNewConsAddrs[newConsAddrKey]; found {
			return fmt.Errorf("duplicate pending consensus key rotation new consensus address for validator %s", rotation.ValidatorAddress)
		}
		if _, found := validatorIndexes.validatorAddressByConsensusAddress[newConsAddrKey]; found {
			return fmt.Errorf("pending consensus key rotation for validator %s targets a live consensus key", rotation.ValidatorAddress)
		}
		if _, found := historyIndexes.byConsensusAddress[newConsAddrKey]; found {
			return fmt.Errorf("pending consensus key rotation for validator %s targets a consensus key in rotation history", rotation.ValidatorAddress)
		}
		pendingNewConsAddrs[newConsAddrKey] = struct{}{}
	}

	// history can lock a live key only while that same validator has a pending
	// rotation. Once the rotation has applied, the old key should no longer be
	// live in the validator set.
	for valAddr, historyConsAddr := range historyIndexes.consensusAddressByValidatorAddress {
		liveValAddr, found := validatorIndexes.validatorAddressByConsensusAddress[historyConsAddr]
		if !found {
			continue
		}
		if liveValAddr != valAddr {
			return fmt.Errorf("consensus key rotation history for validator %s targets another validator's live consensus key", sdk.ValAddress(valAddr).String())
		}
		if _, found := pendingByValidator[valAddr]; !found {
			return fmt.Errorf("consensus key rotation history for validator %s targets a live consensus key", sdk.ValAddress(valAddr).String())
		}
	}

	return nil
}
