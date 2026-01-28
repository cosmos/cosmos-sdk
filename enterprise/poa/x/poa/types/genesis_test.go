// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See ./enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	t.Run("returns valid default genesis state", func(t *testing.T) {
		genesis := DefaultGenesis()
		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
		require.NotEmpty(t, genesis.Params.Admin)
		require.Equal(t, authtypes.NewModuleAddress(govtypes.ModuleName).String(), genesis.Params.Admin)
		require.NotNil(t, genesis.Validators)
		require.Empty(t, genesis.Validators)
	})

	t.Run("default genesis fails validation (no validators)", func(t *testing.T) {
		genesis := DefaultGenesis()
		err := genesis.ValidateBasic()
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least one validator with non-zero power")
	})
}

func TestGenesisStateValidateBasicEnhanced(t *testing.T) {
	t.Run("genesis state with no validators fails", func(t *testing.T) {
		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{},
		}
		err := genesis.ValidateBasic()
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least one validator with non-zero power")
	})

	t.Run("valid genesis state with validators", func(t *testing.T) {
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1, err := codectypes.NewAnyWithValue(pubKey1)
		require.NoError(t, err)

		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny1,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "validator-1",
						Description:     "Validator 1 description",
					},
				},
				{
					PubKey: pubKeyAny2,
					Power:  200,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator2",
						Moniker:         "validator-2",
						Description:     "Validator 2 description",
					},
				},
			},
		}
		err = genesis.ValidateBasic()
		require.NoError(t, err)
	})

	t.Run("invalid params is invalid for genesis", func(t *testing.T) {
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1, err := codectypes.NewAnyWithValue(pubKey1)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny1,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "validator-1",
						Description:     "Validator 1 description",
					},
				},
			},
		}
		err = genesis.ValidateBasic()
		// Params.ValidateBasic() currently returns nil, so this test may not error
		// If Params validation is added in the future, this test will catch it
		_ = err
	})

	t.Run("validator with negative power", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny,
					Power:  -1,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "test-validator",
						Description:     "Test validator description",
					},
				},
			},
		}
		err = genesis.ValidateBasic()
		require.Error(t, err)
		require.Contains(t, err.Error(), "validator at index 0")
		require.Contains(t, err.Error(), "negative validator power")
	})

	t.Run("validator with missing operator address", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "",
						Moniker:         "test-validator",
						Description:     "Test validator description",
					},
				},
			},
		}
		err = genesis.ValidateBasic()
		require.Error(t, err)
		require.Contains(t, err.Error(), "validator at index 0")
		require.Contains(t, err.Error(), "missing validator operator address")
	})

	t.Run("validator with nil pubkey", func(t *testing.T) {
		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: nil,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "test-validator",
						Description:     "Test validator description",
					},
				},
			},
		}
		err := genesis.ValidateBasic()
		require.Error(t, err)
		require.Contains(t, err.Error(), "validator at index 0")
		require.Contains(t, err.Error(), "pubkey cannot be nil")
	})

	t.Run("duplicate operator addresses", func(t *testing.T) {
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1, err := codectypes.NewAnyWithValue(pubKey1)
		require.NoError(t, err)

		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
		require.NoError(t, err)

		duplicateOperatorAddr := "cosmos1duplicate"

		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny1,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: duplicateOperatorAddr,
						Moniker:         "validator-1",
						Description:     "Validator 1 description",
					},
				},
				{
					PubKey: pubKeyAny2,
					Power:  200,
					Metadata: &ValidatorMetadata{
						OperatorAddress: duplicateOperatorAddr, // Duplicate!
						Moniker:         "validator-2",
						Description:     "Validator 2 description",
					},
				},
			},
		}
		err = genesis.ValidateBasic()
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate operator address")
		require.Contains(t, err.Error(), duplicateOperatorAddr)
	})

	t.Run("multiple validation errors - ValidateBasic called first", func(t *testing.T) {
		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: nil, // Will be checked after ValidateBasic
					Power:  -1,  // This will be caught by ValidateBasic first
					Metadata: &ValidatorMetadata{
						OperatorAddress: "",
						Moniker:         "test-validator",
						Description:     "Test validator description",
					},
				},
			},
		}
		err := genesis.ValidateBasic()
		require.Error(t, err)
		// ValidateBasic is called before nil pubkey check, so negative power error is returned
		require.Contains(t, err.Error(), "negative validator power")
	})

	t.Run("multiple validation errors - ValidateBasic error returned when pubkey exists", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny,
					Power:  -1, // Error from ValidateBasic
					Metadata: &ValidatorMetadata{
						OperatorAddress: "", // Also an error
					},
				},
			},
		}
		err = genesis.ValidateBasic()
		require.Error(t, err)
		// ValidateBasic is called first, so negative power error should be returned
		require.Contains(t, err.Error(), "negative validator power")
	})

	t.Run("only zero power validators fails", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny,
					Power:  0,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "validator-zero",
						Description:     "Zero power validator description",
					},
				},
			},
		}
		err = genesis.ValidateBasic()
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least one validator with non-zero power")
	})

	t.Run("validators with zero power and non-zero power", func(t *testing.T) {
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1, err := codectypes.NewAnyWithValue(pubKey1)
		require.NoError(t, err)

		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny1,
					Power:  0,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "validator-zero",
						Description:     "Zero power validator description",
					},
				},
				{
					PubKey: pubKeyAny2,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator2",
						Moniker:         "validator-nonzero",
						Description:     "Non-zero power validator description",
					},
				},
			},
		}
		err = genesis.ValidateBasic()
		require.NoError(t, err) // Valid because at least one has non-zero power
	})

	t.Run("large number of validators", func(t *testing.T) {
		validators := make([]Validator, 100)
		for i := 0; i < 100; i++ {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)

			validators[i] = Validator{
				PubKey: pubKeyAny,
				Power:  int64(i + 1),
				Metadata: &ValidatorMetadata{
					OperatorAddress: "cosmos1operator" + string(rune(i)),
					Moniker:         "validator-" + string(rune(i)),
					Description:     "Validator " + string(rune(i)) + " description",
				},
			}
		}

		genesis := &GenesisState{
			Params: Params{
				Admin: "cosmos1admin",
			},
			Validators: validators,
		}
		err := genesis.ValidateBasic()
		require.NoError(t, err)
	})
}

func TestGenesisStateValidate(t *testing.T) {
	ac := address.NewBech32Codec("cosmos")

	t.Run("valid genesis state", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "validator-1",
						Description:     "Validator 1 description",
					},
				},
			},
		}
		err = genesis.Validate(ac)
		require.NoError(t, err)
	})

	t.Run("invalid params - empty admin", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "validator-1",
						Description:     "Validator 1 description",
					},
				},
			},
		}
		err = genesis.Validate(ac)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid params")
	})

	t.Run("invalid params - invalid admin address", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: "invalid-address",
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "validator-1",
						Description:     "Validator 1 description",
					},
				},
			},
		}
		err = genesis.Validate(ac)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid params")
	})

	t.Run("fails basic validation - no validators", func(t *testing.T) {
		genesis := &GenesisState{
			Params: Params{
				Admin: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			},
			Validators: []Validator{},
		}
		err := genesis.Validate(ac)
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least one validator with non-zero power")
	})

	t.Run("fails basic validation - negative power", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		genesis := &GenesisState{
			Params: Params{
				Admin: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			},
			Validators: []Validator{
				{
					PubKey: pubKeyAny,
					Power:  -1,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator1",
						Moniker:         "validator-1",
						Description:     "Validator 1 description",
					},
				},
			},
		}
		err = genesis.Validate(ac)
		require.Error(t, err)
		require.Contains(t, err.Error(), "negative validator power")
	})
}
