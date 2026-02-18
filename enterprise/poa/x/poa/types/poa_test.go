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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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
			wantErr: false,
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
			},
			wantErr: true,
		},
		{
			name: "missing operator address",
			msg: &MsgCreateValidator{
				OperatorAddress: "",
				Moniker:         "test-moniker",
				Description:     "test-description",
			},
			wantErr: true,
		},
		{
			name: "empty moniker",
			msg: &MsgCreateValidator{
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         "",
				Description:     "test-description",
			},
			wantErr: true,
		},
		{
			name: "moniker too long",
			msg: &MsgCreateValidator{
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         strings.Repeat("a", 257),
				Description:     "test-description",
			},
			wantErr: true,
		},
		{
			name: "empty description",
			msg: &MsgCreateValidator{
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         "test-moniker",
				Description:     "",
			},
			wantErr: true,
		},
		{
			name: "description too long",
			msg: &MsgCreateValidator{
				OperatorAddress: "cosmos1w3jhxarpv3j8yvg4ufs4x",
				Moniker:         "test-moniker",
				Description:     strings.Repeat("a", 257),
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
				}
			}(),
			wantErr: false,
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
