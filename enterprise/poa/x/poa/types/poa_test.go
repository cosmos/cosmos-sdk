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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package types

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidatorValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		validator *Validator
		wantErr   error
	}{
		{
			name: "valid validator",
			validator: func() *Validator {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &Validator{
					PubKey: pubKeyAny,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator",
						Moniker:         "test-validator",
						Description:     "Test validator description",
					},
				}
			}(),
			wantErr: nil,
		},
		{
			name: "negative power",
			validator: &Validator{
				Power: -1,
				Metadata: &ValidatorMetadata{
					OperatorAddress: "cosmos1operator",
					Moniker:         "test-validator",
					Description:     "Test validator description",
				},
			},
			wantErr: ErrNegativeValidatorPower,
		},
		{
			name: "missing operator address",
			validator: func() *Validator {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &Validator{
					PubKey: pubKeyAny,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "",
						Moniker:         "test-validator",
						Description:     "Test validator description",
					},
				}
			}(),
			wantErr: ErrMissingOperatorAddress,
		},
		{
			name: "zero power is valid",
			validator: func() *Validator {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &Validator{
					PubKey: pubKeyAny,
					Power:  0,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator",
						Moniker:         "zero-power-validator",
						Description:     "Zero power validator description",
					},
				}
			}(),
			wantErr: nil,
		},
		{
			name: "pubkey too short",
			validator: func() *Validator {
				// Create Any with Value shorter than MinPubKeyLength
				pubKeyAny := &codectypes.Any{TypeUrl: "/cosmos.crypto.ed25519.PubKey", Value: make([]byte, 10)}
				return &Validator{
					PubKey: pubKeyAny,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator",
						Moniker:         "test-validator",
						Description:     "Test validator description",
					},
				}
			}(),
			wantErr: ErrInvalidPubKeyLength,
		},
		{
			name: "pubkey too long",
			validator: func() *Validator {
				// Create Any with Value longer than MaxPubKeyLength
				pubKeyAny := &codectypes.Any{TypeUrl: "/cosmos.crypto.ed25519.PubKey", Value: make([]byte, MaxPubKeyLength+1)}
				return &Validator{
					PubKey: pubKeyAny,
					Power:  100,
					Metadata: &ValidatorMetadata{
						OperatorAddress: "cosmos1operator",
						Moniker:         "test-validator",
						Description:     "Test validator description",
					},
				}
			}(),
			wantErr: ErrInvalidPubKeyLength,
		},
		{
			name: "nil metadata",
			validator: func() *Validator {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &Validator{
					PubKey:   pubKeyAny,
					Power:    100,
					Metadata: nil,
				}
			}(),
			wantErr: ErrInvalidMetadata,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.ValidateBasic()
			if tt.wantErr != nil {
				require.Error(t, err)
				// Error is wrapped, so check that the error message contains the expected error
				require.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateValidatorSet(t *testing.T) {
	makeValidator := func(power int64, operator string) Validator {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		return Validator{
			PubKey: pubKeyAny,
			Power:  power,
			Metadata: &ValidatorMetadata{
				OperatorAddress: operator,
				Moniker:         "validator-" + operator,
				Description:     "desc",
			},
		}
	}

	tests := []struct {
		name       string
		validators []Validator
		wantErr    string
	}{
		{
			name:       "empty set invalid",
			validators: []Validator{},
			wantErr:    "total power must be greater than zero",
		},
		{
			name: "all zero power invalid",
			validators: []Validator{
				makeValidator(0, "cosmos1operator1"),
				makeValidator(0, "cosmos1operator2"),
			},
			wantErr: "total power must be greater than zero",
		},
		{
			name: "duplicate operator invalid",
			validators: []Validator{
				makeValidator(100, "cosmos1duplicate"),
				makeValidator(200, "cosmos1duplicate"),
			},
			wantErr: "duplicate operator address",
		},
		{
			name: "total power overflow invalid",
			validators: []Validator{
				makeValidator(math.MaxInt64/2+1, "cosmos1operator1"),
				makeValidator(math.MaxInt64/2+1, "cosmos1operator2"),
			},
			wantErr: ErrTotalPowerOverflow.Error(),
		},
		{
			name: "valid set",
			validators: []Validator{
				makeValidator(100, "cosmos1operator1"),
				makeValidator(200, "cosmos1operator2"),
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateValidatorSet(tt.validators)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatorMetadataValidateBasic(t *testing.T) {
	tests := []struct {
		name     string
		metadata *ValidatorMetadata
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid metadata",
			metadata: &ValidatorMetadata{
				OperatorAddress: "cosmos1operator",
				Moniker:         "validator",
				Description:     "description",
			},
			wantErr: false,
		},
		{
			name: "missing operator address",
			metadata: &ValidatorMetadata{
				OperatorAddress: "",
				Moniker:         "validator",
				Description:     "description",
			},
			wantErr: true,
			errMsg:  ErrMissingOperatorAddress.Error(),
		},
		{
			name: "empty moniker",
			metadata: &ValidatorMetadata{
				OperatorAddress: "cosmos1operator",
				Moniker:         "",
				Description:     "description",
			},
			wantErr: true,
			errMsg:  "moniker cannot be empty",
		},
		{
			name: "moniker at max length is valid",
			metadata: &ValidatorMetadata{
				OperatorAddress: "cosmos1operator",
				Moniker:         strings.Repeat("m", 256),
				Description:     "description",
			},
			wantErr: false,
		},
		{
			name: "moniker too long",
			metadata: &ValidatorMetadata{
				OperatorAddress: "cosmos1operator",
				Moniker:         strings.Repeat("m", 257),
				Description:     "description",
			},
			wantErr: true,
			errMsg:  "moniker too long",
		},
		{
			name: "empty description is valid",
			metadata: &ValidatorMetadata{
				OperatorAddress: "cosmos1operator",
				Moniker:         "validator",
				Description:     "",
			},
			wantErr: false,
		},
		{
			name: "description at max length is valid",
			metadata: &ValidatorMetadata{
				OperatorAddress: "cosmos1operator",
				Moniker:         "validator",
				Description:     strings.Repeat("a", 256),
			},
			wantErr: false,
		},
		{
			name: "description too long",
			metadata: &ValidatorMetadata{
				OperatorAddress: "cosmos1operator",
				Moniker:         "validator",
				Description:     strings.Repeat("a", 257),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParamsValidate(t *testing.T) {
	ac := address.NewBech32Codec("cosmos")

	tests := []struct {
		name    string
		params  *Params
		wantErr bool
	}{
		{
			name: "valid admin address",
			params: &Params{
				Admin: "cosmos1w3jhxarpv3j8yvg4ufs4x",
			},
			wantErr: false,
		},
		{
			name: "empty admin address",
			params: &Params{
				Admin: "",
			},
			wantErr: true,
		},
		{
			name: "invalid admin address",
			params: &Params{
				Admin: "invalid-address",
			},
			wantErr: true,
		},
		{
			name: "admin address with wrong prefix",
			params: &Params{
				Admin: "osmo1w3jhxarpv3j8yvg4ufs4x",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate(ac)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenesisStateValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		gs      *GenesisState
		wantErr bool
	}{
		{
			name: "valid genesis state",
			gs: func() *GenesisState {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &GenesisState{
					Params: Params{},
					Validators: []Validator{
						{
							PubKey: pubKeyAny,
							Power:  100,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "cosmos1operator",
								Moniker:         "test-validator",
								Description:     "Test validator description",
							},
						},
					},
				}
			}(),
			wantErr: false,
		},
		{
			name: "empty genesis state",
			gs: &GenesisState{
				Params:     Params{},
				Validators: []Validator{},
			},
			wantErr: true, // Must have at least one validator with non-zero power
		},
		{
			name: "invalid validator in genesis",
			gs: func() *GenesisState {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &GenesisState{
					Params: Params{},
					Validators: []Validator{
						{
							PubKey: pubKeyAny,
							Power:  -1,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "cosmos1operator",
								Moniker:         "test-validator",
								Description:     "Test validator description",
							},
						},
					},
				}
			}(),
			wantErr: true,
		},
		{
			name: "multiple validators with one invalid",
			gs: func() *GenesisState {
				pubKey1 := ed25519.GenPrivKey().PubKey()
				pubKeyAny1, _ := codectypes.NewAnyWithValue(pubKey1)
				pubKey2 := ed25519.GenPrivKey().PubKey()
				pubKeyAny2, _ := codectypes.NewAnyWithValue(pubKey2)
				return &GenesisState{
					Params: Params{},
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
							Power:  50,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "",
								Moniker:         "validator-2",
								Description:     "Validator 2 description",
							},
						},
					},
				}
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gs.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateParamsValidate(t *testing.T) {
	ac := address.NewBech32Codec("cosmos")

	tests := []struct {
		name    string
		msg     *MsgUpdateParams
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgUpdateParams{
				Admin:  "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Params: Params{Admin: "cosmos1w3jhxarpv3j8yvg4ufs4x"},
			},
			wantErr: false,
		},
		{
			name: "empty admin (signer)",
			msg: &MsgUpdateParams{
				Admin:  "",
				Params: Params{Admin: "cosmos1w3jhxarpv3j8yvg4ufs4x"},
			},
			wantErr: true,
		},
		{
			name: "invalid admin (signer) address",
			msg: &MsgUpdateParams{
				Admin:  "invalid-address",
				Params: Params{Admin: "cosmos1w3jhxarpv3j8yvg4ufs4x"},
			},
			wantErr: true,
		},
		{
			name: "empty params admin",
			msg: &MsgUpdateParams{
				Admin:  "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Params: Params{Admin: ""},
			},
			wantErr: true,
		},
		{
			name: "invalid params admin address",
			msg: &MsgUpdateParams{
				Admin:  "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Params: Params{Admin: "invalid-address"},
			},
			wantErr: true,
		},
		{
			name: "both admin addresses invalid",
			msg: &MsgUpdateParams{
				Admin:  "invalid1",
				Params: Params{Admin: "invalid2"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate(ac)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateValidatorsValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgUpdateValidators
		wantErr bool
	}{
		{
			name: "valid message",
			msg: func() *MsgUpdateValidators {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &MsgUpdateValidators{
					Validators: []Validator{
						{
							PubKey: pubKeyAny,
							Power:  100,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "cosmos1operator",
								Moniker:         "test-validator",
								Description:     "Test validator description",
							},
						},
					},
				}
			}(),
			wantErr: false,
		},
		{
			name: "empty validators list",
			msg: &MsgUpdateValidators{
				Validators: []Validator{},
			},
			wantErr: true,
		},
		{
			name: "all zero power validators",
			msg: func() *MsgUpdateValidators {
				pubKey1 := ed25519.GenPrivKey().PubKey()
				pubKeyAny1, _ := codectypes.NewAnyWithValue(pubKey1)
				pubKey2 := ed25519.GenPrivKey().PubKey()
				pubKeyAny2, _ := codectypes.NewAnyWithValue(pubKey2)
				return &MsgUpdateValidators{
					Validators: []Validator{
						{
							PubKey: pubKeyAny1,
							Power:  0,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "cosmos1operator1",
								Moniker:         "validator-1",
								Description:     "Validator 1 description",
							},
						},
						{
							PubKey: pubKeyAny2,
							Power:  0,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "cosmos1operator2",
								Moniker:         "validator-2",
								Description:     "Validator 2 description",
							},
						},
					},
				}
			}(),
			wantErr: true,
		},
		{
			name: "invalid validator",
			msg: func() *MsgUpdateValidators {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &MsgUpdateValidators{
					Validators: []Validator{
						{
							PubKey: pubKeyAny,
							Power:  -1,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "cosmos1operator",
								Moniker:         "test-validator",
								Description:     "Test validator description",
							},
						},
					},
				}
			}(),
			wantErr: true,
		},
		{
			name: "duplicate operator addresses",
			msg: func() *MsgUpdateValidators {
				pubKey1 := ed25519.GenPrivKey().PubKey()
				pubKeyAny1, _ := codectypes.NewAnyWithValue(pubKey1)
				pubKey2 := ed25519.GenPrivKey().PubKey()
				pubKeyAny2, _ := codectypes.NewAnyWithValue(pubKey2)
				duplicateOperatorAddr := "cosmos1duplicate"
				return &MsgUpdateValidators{
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
			}(),
			wantErr: true,
		},
		{
			name: "multiple validators with one invalid",
			msg: func() *MsgUpdateValidators {
				pubKey1 := ed25519.GenPrivKey().PubKey()
				pubKeyAny1, _ := codectypes.NewAnyWithValue(pubKey1)
				pubKey2 := ed25519.GenPrivKey().PubKey()
				pubKeyAny2, _ := codectypes.NewAnyWithValue(pubKey2)
				return &MsgUpdateValidators{
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
							Power:  50,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "",
								Moniker:         "validator-2",
								Description:     "Validator 2 description",
							},
						},
					},
				}
			}(),
			wantErr: true,
		},
		{
			name: "total power overflow",
			msg: func() *MsgUpdateValidators {
				pubKey1 := ed25519.GenPrivKey().PubKey()
				pubKeyAny1, _ := codectypes.NewAnyWithValue(pubKey1)
				pubKey2 := ed25519.GenPrivKey().PubKey()
				pubKeyAny2, _ := codectypes.NewAnyWithValue(pubKey2)
				// Two validators whose powers sum to > math.MaxInt64
				power1 := int64(math.MaxInt64/2 + 1)
				power2 := int64(math.MaxInt64/2 + 1)
				return &MsgUpdateValidators{
					Validators: []Validator{
						{
							PubKey: pubKeyAny1,
							Power:  power1,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "cosmos1operator1",
								Moniker:         "validator-1",
								Description:     "Validator 1 description",
							},
						},
						{
							PubKey: pubKeyAny2,
							Power:  power2,
							Metadata: &ValidatorMetadata{
								OperatorAddress: "cosmos1operator2",
								Moniker:         "validator-2",
								Description:     "Validator 2 description",
							},
						},
					},
				}
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateValidatorValidate(t *testing.T) {
	ac := address.NewBech32Codec("cosmos")
	adminAddr := "cosmos1w3jhxarpv3j8yvg4ufs4x"

	tests := []struct {
		name    string
		msg     *MsgCreateValidator
		wantErr bool
	}{
		{
			name: "valid message",
			msg: func() *MsgCreateValidator {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &MsgCreateValidator{
					PubKey:          pubKeyAny,
					OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
					Moniker:         "test-moniker",
					Description:     "test-description",
					Admin:           adminAddr,
				}
			}(),
			wantErr: false,
		},
		{
			name: "nil pubkey",
			msg: &MsgCreateValidator{
				PubKey:          nil,
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         "test-moniker",
				Description:     "test-description",
				Admin:           adminAddr,
			},
			wantErr: true,
		},
		{
			name: "missing operator address",
			msg: &MsgCreateValidator{
				OperatorAddress: "",
				Moniker:         "test-moniker",
				Description:     "test-description",
				Admin:           adminAddr,
			},
			wantErr: true,
		},
		{
			name: "empty moniker",
			msg: &MsgCreateValidator{
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         "",
				Description:     "test-description",
				Admin:           adminAddr,
			},
			wantErr: true,
		},
		{
			name: "moniker too long",
			msg: &MsgCreateValidator{
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         strings.Repeat("a", 257),
				Description:     "test-description",
				Admin:           adminAddr,
			},
			wantErr: true,
		},
		{
			name: "empty description",
			msg: func() *MsgCreateValidator {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &MsgCreateValidator{
					PubKey:          pubKeyAny,
					OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
					Moniker:         "test-moniker",
					Description:     "",
					Admin:           adminAddr,
				}
			}(),
			wantErr: false,
		},
		{
			name: "description too long",
			msg: &MsgCreateValidator{
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         "test-moniker",
				Description:     strings.Repeat("a", 257),
				Admin:           adminAddr,
			},
			wantErr: true,
		},
		{
			name: "moniker at max length (256)",
			msg: func() *MsgCreateValidator {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &MsgCreateValidator{
					PubKey:          pubKeyAny,
					OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
					Moniker:         strings.Repeat("a", 256),
					Description:     "test-description",
					Admin:           adminAddr,
				}
			}(),
			wantErr: false,
		},
		{
			name: "description at max length (256)",
			msg: func() *MsgCreateValidator {
				pubKey := ed25519.GenPrivKey().PubKey()
				pubKeyAny, _ := codectypes.NewAnyWithValue(pubKey)
				return &MsgCreateValidator{
					PubKey:          pubKeyAny,
					OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
					Moniker:         "test-moniker",
					Description:     strings.Repeat("a", 256),
					Admin:           adminAddr,
				}
			}(),
			wantErr: false,
		},
		{
			name: "pubkey too short",
			msg: &MsgCreateValidator{
				PubKey:          &codectypes.Any{TypeUrl: "/cosmos.crypto.ed25519.PubKey", Value: make([]byte, 10)},
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         "test-moniker",
				Description:     "test-description",
				Admin:           adminAddr,
			},
			wantErr: true,
		},
		{
			name: "pubkey too long",
			msg: &MsgCreateValidator{
				PubKey:          &codectypes.Any{TypeUrl: "/cosmos.crypto.ed25519.PubKey", Value: make([]byte, MaxPubKeyLength+1)},
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         "test-moniker",
				Description:     "test-description",
				Admin:           adminAddr,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate(ac)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatorFeesValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		fees    *ValidatorFees
		wantErr bool
	}{
		{
			name:    "empty fees",
			fees:    &ValidatorFees{},
			wantErr: false,
		},
		{
			name: "valid fees",
			fees: &ValidatorFees{
				Fees: sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))},
			},
			wantErr: false,
		},
		{
			name: "negative fee amount",
			fees: &ValidatorFees{
				Fees: sdk.DecCoins{sdk.DecCoin{Denom: "stake", Amount: sdkmath.LegacyNewDec(-1)}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fees.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidAllocatedFees)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenesisAllocatedFeesValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		entry   *GenesisAllocatedFees
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid entry",
			entry: &GenesisAllocatedFees{
				ConsensusAddress: "cosmosvalcons1abc",
				Fees:             sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))},
			},
			wantErr: false,
		},
		{
			name: "empty consensus address",
			entry: &GenesisAllocatedFees{
				ConsensusAddress: "",
				Fees:             sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))},
			},
			wantErr: true,
			errMsg:  "consensus address cannot be empty",
		},
		{
			name: "negative fee",
			entry: &GenesisAllocatedFees{
				ConsensusAddress: "cosmosvalcons1abc",
				Fees:             sdk.DecCoins{sdk.DecCoin{Denom: "stake", Amount: sdkmath.LegacyNewDec(-1)}},
			},
			wantErr: true,
			errMsg:  "negative fee amount",
		},
		{
			name: "empty fees is valid",
			entry: &GenesisAllocatedFees{
				ConsensusAddress: "cosmosvalcons1abc",
				Fees:             sdk.DecCoins{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidAllocatedFees)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenesisAllocatedFeesValidate(t *testing.T) {
	consAddr := sdk.ConsAddress(ed25519.GenPrivKey().PubKey().Address())

	tests := []struct {
		name    string
		prefix  string
		entry   *GenesisAllocatedFees
		wantErr string
	}{
		{
			name:   "valid with cosmosvalcons prefix",
			prefix: "cosmosvalcons",
			entry: &GenesisAllocatedFees{
				ConsensusAddress: consAddr.String(),
				Fees:             sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))},
			},
		},
		{
			name:   "valid with custom prefix",
			prefix: "mychain",
			entry: &GenesisAllocatedFees{
				ConsensusAddress: sdk.MustBech32ifyAddressBytes("mychain", consAddr),
				Fees:             sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))},
			},
		},
		{
			name:   "wrong prefix rejected",
			prefix: "otherchain",
			entry: &GenesisAllocatedFees{
				ConsensusAddress: consAddr.String(), // cosmosvalcons prefix
				Fees:             sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))},
			},
			wantErr: "invalid consensus address",
		},
		{
			name:   "invalid bech32 rejected",
			prefix: "cosmosvalcons",
			entry: &GenesisAllocatedFees{
				ConsensusAddress: "invalid-address",
				Fees:             sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))},
			},
			wantErr: "invalid consensus address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := address.NewBech32Codec(tt.prefix)
			err := tt.entry.Validate(ac)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAllocatedFees(t *testing.T) {
	tests := []struct {
		name    string
		fees    []GenesisAllocatedFees
		wantErr string
	}{
		{
			name: "nil list is valid",
			fees: nil,
		},
		{
			name: "empty list is valid",
			fees: []GenesisAllocatedFees{},
		},
		{
			name: "valid entries",
			fees: []GenesisAllocatedFees{
				{ConsensusAddress: "addr1", Fees: sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))}},
				{ConsensusAddress: "addr2", Fees: sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(200))}},
			},
		},
		{
			name: "duplicate consensus addresses",
			fees: []GenesisAllocatedFees{
				{ConsensusAddress: "addr1", Fees: sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))}},
				{ConsensusAddress: "addr1", Fees: sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(200))}},
			},
			wantErr: "duplicate consensus address",
		},
		{
			name: "empty consensus address",
			fees: []GenesisAllocatedFees{
				{ConsensusAddress: "addr1", Fees: sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdkmath.LegacyNewDec(100))}},
				{ConsensusAddress: "", Fees: sdk.DecCoins{}},
			},
			wantErr: "allocated fees at index 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAllocatedFees(tt.fees)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgWithdrawFeesValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgWithdrawFees
		wantErr error
	}{
		{
			name: "valid message",
			msg: &MsgWithdrawFees{
				Operator: "cosmos1operator",
			},
			wantErr: nil,
		},
		{
			name: "missing operator",
			msg: &MsgWithdrawFees{
				Operator: "",
			},
			wantErr: ErrMissingOperatorAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
