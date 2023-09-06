package testnet

import (
	"fmt"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmttypes "github.com/cometbft/cometbft/types"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ValidatorPrivKeys is a slice of [*ValidatorPrivKey].
type ValidatorPrivKeys []*ValidatorPrivKey

// ValidatorPrivKey holds a validator key (a comet ed25519 key)
// and the validator's delegator or account key (a Cosmos SDK secp256k1 key).
type ValidatorPrivKey struct {
	Val cmted25519.PrivKey
	Del *secp256k1.PrivKey
}

// NewValidatorPrivKeys returns a ValidatorPrivKeys of length n,
// where each set of keys is dynamically generated.
//
// If writing a test where deterministic keys are required,
// the caller should manually construct a slice and assign each key as needed.
func NewValidatorPrivKeys(n int) ValidatorPrivKeys {
	vpk := make(ValidatorPrivKeys, n)

	for i := range vpk {
		vpk[i] = &ValidatorPrivKey{
			Val: cmted25519.GenPrivKey(),
			Del: secp256k1.GenPrivKey(),
		}
	}

	return vpk
}

// CometGenesisValidators derives the CometGenesisValidators belonging to vpk.
func (vpk ValidatorPrivKeys) CometGenesisValidators() CometGenesisValidators {
	cgv := make(CometGenesisValidators, len(vpk))

	for i, pk := range vpk {
		pubKey := pk.Val.PubKey()

		const votingPower = 1
		cmtVal := cmttypes.NewValidator(pubKey, votingPower)

		cgv[i] = &CometGenesisValidator{
			V: cmttypes.GenesisValidator{
				Address: cmtVal.Address,
				PubKey:  cmtVal.PubKey,
				Power:   cmtVal.VotingPower,
				Name:    fmt.Sprintf("val-%d", i),
			},
			PK: pk,
		}
	}

	return cgv
}

// CometGenesisValidators is a slice of [*CometGenesisValidator].
type CometGenesisValidators []*CometGenesisValidator

// CometGenesisValidator holds a comet GenesisValidator
// and a reference to the ValidatorPrivKey from which the CometGenesisValidator was derived.
type CometGenesisValidator struct {
	V  cmttypes.GenesisValidator
	PK *ValidatorPrivKey
}

// ToComet returns a new slice of [cmttypes.GenesisValidator],
// useful for some interactions.
func (cgv CometGenesisValidators) ToComet() []cmttypes.GenesisValidator {
	vs := make([]cmttypes.GenesisValidator, len(cgv))
	for i, v := range cgv {
		vs[i] = v.V
	}
	return vs
}

// StakingValidators derives the StakingValidators belonging to cgv.
func (cgv CometGenesisValidators) StakingValidators() StakingValidators {
	vals := make(StakingValidators, len(cgv))
	for i, v := range cgv {
		pk, err := cryptocodec.FromCmtPubKeyInterface(v.V.PubKey)
		if err != nil {
			panic(fmt.Errorf("failed to extract comet pub key: %w", err))
		}

		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			panic(fmt.Errorf("failed to wrap pub key in any type: %w", err))
		}

		vals[i] = &StakingValidator{
			V: stakingtypes.Validator{
				OperatorAddress:   sdk.ValAddress(v.V.Address).String(), // TODO: this relies on global bech32 config.
				ConsensusPubkey:   pkAny,
				Status:            stakingtypes.Bonded,
				Tokens:            sdk.DefaultPowerReduction,
				DelegatorShares:   sdkmath.LegacyOneDec(),
				MinSelfDelegation: sdkmath.ZeroInt(),

				// more fields uncopied from testutil/sims/app_helpers.go:220
			},
			C:  v,
			PK: v.PK,
		}
	}

	return vals
}

// StakingValidators is a slice of [*StakingValidator].
type StakingValidators []*StakingValidator

// StakingValidator holds a [stakingtypes.Validator],
// and the CometGenesisValidator and ValidatorPrivKey required to derive the StakingValidator.
type StakingValidator struct {
	V  stakingtypes.Validator
	C  *CometGenesisValidator
	PK *ValidatorPrivKey
}

// ToStakingType returns a new slice of [stakingtypes.Validator],
// useful for some interactions.
func (sv StakingValidators) ToStakingType() []stakingtypes.Validator {
	vs := make([]stakingtypes.Validator, len(sv))
	for i, v := range sv {
		vs[i] = v.V
	}
	return vs
}

// BaseAccounts returns the BaseAccounts for this set of StakingValidators.
// The base accounts are important for [*GenesisBuilder.BaseAccounts].
func (sv StakingValidators) BaseAccounts() BaseAccounts {
	ba := make(BaseAccounts, len(sv))

	for i, v := range sv {
		const accountNumber = 0
		const sequenceNumber = 0

		pubKey := v.PK.Del.PubKey()
		bech, err := bech32.ConvertAndEncode("cosmos", pubKey.Address().Bytes()) // TODO: this shouldn't be hardcoded to cosmos!
		if err != nil {
			panic(err)
		}
		accAddr, err := sdk.AccAddressFromBech32(bech)
		if err != nil {
			panic(err)
		}
		ba[i] = authtypes.NewBaseAccount(
			accAddr, pubKey, accountNumber, sequenceNumber,
		)
	}

	return ba
}

// Balances returns the balances held by this set of StakingValidators.
func (sv StakingValidators) Balances() []banktypes.Balance {
	bals := make([]banktypes.Balance, len(sv))

	for i, v := range sv {
		addr, err := bech32.ConvertAndEncode("cosmos", v.PK.Del.PubKey().Address().Bytes()) // TODO: this shouldn't be hardcoded to cosmos!
		if err != nil {
			panic(err)
		}
		bals[i] = banktypes.Balance{
			Address: addr,
			Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, v.V.Tokens)},
		}
	}

	return bals
}
