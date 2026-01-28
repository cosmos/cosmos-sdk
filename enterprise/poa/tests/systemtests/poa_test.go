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

//go:build system_test

package systemtests

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"cosmossdk.io/systemtests"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
)

const (
	poaModule = "poa"
)

// getAdmin queries the POA params and returns the admin address and corresponding key name
func getAdmin(t *testing.T, cli *systemtests.CLIWrapper) (addr, keyName string) {
	t.Helper()
	rsp := cli.CustomQuery("q", poaModule, "params")
	addr = gjson.Get(rsp, "params.admin").String()
	require.NotEmpty(t, addr, "admin should be set in params")

	keyName = getKeyNameForAddress(t, cli, addr)
	require.NotEmpty(t, keyName, "should find key for admin")
	return addr, keyName
}

// getKeyNameForAddress finds the keyring key name for a given address
func getKeyNameForAddress(t *testing.T, cli *systemtests.CLIWrapper, addr string) string {
	t.Helper()
	keysOutput := cli.Keys("keys", "list")
	keys := gjson.Get(keysOutput, "@this").Array()
	for _, key := range keys {
		if gjson.Get(key.Raw, "address").String() == addr {
			return gjson.Get(key.Raw, "name").String()
		}
	}
	return ""
}

// requireTxFailed checks that a transaction failed, either via non-zero code or error in response
func requireTxFailed(t *testing.T, rsp string) {
	t.Helper()
	code := gjson.Get(rsp, "code").Int()
	rawLog := gjson.Get(rsp, "raw_log").String()
	// Transaction failed if code is non-zero OR response contains error indicators
	failed := code != 0 ||
		strings.Contains(rawLog, "error") ||
		strings.Contains(rawLog, "invalid") ||
		strings.Contains(rawLog, "failed") ||
		strings.Contains(rsp, "error") ||
		strings.Contains(rsp, "invalid")
	require.True(t, failed, "expected transaction to fail, got: %s", rsp)
}

func TestPOAQueries(t *testing.T) {
	// Scenario:
	// Test all POA query endpoints
	// - params
	// - validators
	// - validator (single)
	// - total-power
	// - withdrawable-fees

	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	sut.StartChain(t)

	t.Run("query params", func(t *testing.T) {
		rsp := cli.CustomQuery("q", poaModule, "params")
		admin := gjson.Get(rsp, "params.admin").String()
		require.NotEmpty(t, admin, "admin should be set in params")
	})

	t.Run("query validators", func(t *testing.T) {
		rsp := cli.CustomQuery("q", poaModule, "validators")
		validators := gjson.Get(rsp, "validators").Array()
		require.NotEmpty(t, validators, "should have at least one validator")

		// Check first validator has expected fields
		firstValidator := validators[0]
		require.NotEmpty(t, gjson.Get(firstValidator.Raw, "pub_key").String(), "validator should have pub_key")
		power := gjson.Get(firstValidator.Raw, "power").Int()
		require.Equal(t, power, int64(10_000), "validator should have positive power")
	})

	t.Run("query single validator", func(t *testing.T) {
		// First get validator address from the list
		rsp := cli.CustomQuery("q", poaModule, "validators")
		validators := gjson.Get(rsp, "validators").Array()
		require.NotEmpty(t, validators)

		operatorAddr := gjson.Get(validators[0].Raw, "metadata.operator_address").String()
		require.NotEmpty(t, operatorAddr)

		// Query single validator
		rsp = cli.CustomQuery("q", poaModule, "validator", operatorAddr)
		validator := gjson.Get(rsp, "validator")
		require.True(t, validator.Exists(), "validator should exist in response")
		assert.Equal(t, operatorAddr, gjson.Get(rsp, "validator.metadata.operator_address").String())
	})

	t.Run("query total power", func(t *testing.T) {
		rsp := cli.CustomQuery("q", poaModule, "total-power")
		totalPower := gjson.Get(rsp, "total_power").Int()
		require.Equal(t, totalPower, int64(10_000*sut.NodesCount()), "total power should be positive")
	})

	t.Run("query withdrawable fees", func(t *testing.T) {
		// Get validator operator address
		rsp := cli.CustomQuery("q", poaModule, "validators")
		validators := gjson.Get(rsp, "validators").Array()
		require.NotEmpty(t, validators)

		operatorAddr := gjson.Get(validators[0].Raw, "metadata.operator_address").String()

		// Query withdrawable fees - may be empty initially but should not error
		rsp = cli.CustomQuery("q", poaModule, "withdrawable-fees", operatorAddr)
		// Response should have fees field (may be empty array)
		require.Contains(t, rsp, "fees")
	})
}

func TestUpdateParams(t *testing.T) {
	// Scenario:
	// - Admin updates params to change admin address
	// - Verify params query reflects change
	// - Non-admin attempting update fails

	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	// Add a new account that will become the new admin
	newAdminAddr := cli.AddKey("newadmin")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", newAdminAddr, "10000000stake"},
	)

	sut.StartChain(t)

	currentAdminAddr, currentAdminKeyName := getAdmin(t, cli)
	t.Logf("Current admin: %s (%s)", currentAdminAddr, currentAdminKeyName)
	t.Logf("New admin address: %s", newAdminAddr)

	t.Run("admin successfully updates params", func(t *testing.T) {
		// Update admin to new address
		rsp := cli.Run(
			"tx", poaModule, "update-params",
			"--admin="+newAdminAddr,
			"--from="+currentAdminKeyName,
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)

		// Verify params changed
		rsp = cli.CustomQuery("q", poaModule, "params")
		admin := gjson.Get(rsp, "params.admin").String()
		t.Logf("Admin after update: %s", admin)
		assert.Equal(t, newAdminAddr, admin, "admin should be updated to new address")
	})

	t.Run("old admin can no longer update params", func(t *testing.T) {
		// Try to update params with old admin - should fail
		rsp := cli.Run(
			"tx", poaModule, "update-params",
			"--admin="+currentAdminAddr,
			"--from="+currentAdminKeyName,
			"--fees=1stake",
		)
		requireTxFailed(t, rsp)
	})

	t.Run("new admin can update params", func(t *testing.T) {
		// New admin updates params back to original admin
		rsp := cli.Run(
			"tx", poaModule, "update-params",
			"--admin="+currentAdminAddr,
			"--from="+newAdminAddr,
			"--fees=1stake",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)

		// Verify params changed back
		rsp = cli.CustomQuery("q", poaModule, "params")
		admin := gjson.Get(rsp, "params.admin").String()
		assert.Equal(t, currentAdminAddr, admin, "admin should be reverted to original")
	})
}

func TestUpdateValidators(t *testing.T) {
	// Scenario:
	// - Admin updates existing validator's power
	// - Admin updates existing validator's metadata
	// - Verify changes in queries
	// - Non-admin attempting update fails
	// - Attempting to add unknown validator fails

	sut := systemtests.Sut
	sut.ResetChain(t)
	sut.StartChain(t)
	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	_, adminKeyName := getAdmin(t, cli)

	// Get current validator info
	rsp := cli.CustomQuery("q", poaModule, "validators")
	validatorsJSON := gjson.Get(rsp, "validators").Array()

	t.Run("admin updates validator", func(t *testing.T) {
		validatorToEdit := validatorsJSON[0]
		operatorAddr := gjson.Get(validatorToEdit.Raw, "metadata.operator_address").String()
		pubKeyType := gjson.Get(validatorToEdit.Raw, `pub_key.type`).String()
		pubKeyKey := gjson.Get(validatorToEdit.Raw, "pub_key.value").String()
		moniker := gjson.Get(validatorToEdit.Raw, "metadata.moniker").String()
		currentPower := gjson.Get(validatorToEdit.Raw, "power").Int()

		// Set new power to something different
		newPower := currentPower + 1000000

		// Build JSON with updated power (manually construct pub_key to preserve @type)
		updatedValidatorJSON := fmt.Sprintf(`[{
			"pub_key": {
				"@type": "%s",
				"key": "%s"
			},
			"power": %d,
			"metadata": {
				"moniker": "%s",
				"operator_address": "%s"
			}
		}]`, pubKeyType, pubKeyKey, newPower, moniker, operatorAddr)

		validatorsFile := systemtests.StoreTempFile(t, []byte(updatedValidatorJSON))
		defer validatorsFile.Close()

		rsp := cli.Run(
			"tx", poaModule, "update-validators",
			validatorsFile.Name(),
			"--from="+adminKeyName,
			"--gas=auto",
		)
		systemtests.RequireTxSuccess(t, rsp)

		// Verify power changed
		rsp = cli.CustomQuery("q", poaModule, "validator", operatorAddr)
		updatedPower := gjson.Get(rsp, "validator.power").Int()
		assert.Equal(t, newPower, updatedPower, "validator power should be updated")
	})

	t.Run("non-admin cannot update validators", func(t *testing.T) {
		validatorToEdit := validatorsJSON[0]
		operatorAddr := gjson.Get(validatorToEdit.Raw, "metadata.operator_address").String()
		pubKeyType := gjson.Get(validatorToEdit.Raw, `pub_key.type`).String()
		pubKeyKey := gjson.Get(validatorToEdit.Raw, "pub_key.value").String()
		moniker := gjson.Get(validatorToEdit.Raw, "metadata.moniker").String()
		currentPower := gjson.Get(validatorToEdit.Raw, "power").Int()
		newPower := currentPower + 1

		// Build JSON with updated power (manually construct pub_key to preserve @type)
		updatedValidatorJSON := fmt.Sprintf(`[{
			"pub_key": {
				"@type": "%s",
				"key": "%s"
			},
			"power": %d,
			"metadata": {
				"moniker": "%s",
				"operator_address": "%s"
			}
		}]`, pubKeyType, pubKeyKey, newPower, moniker, operatorAddr)

		validatorsFile := systemtests.StoreTempFile(t, []byte(updatedValidatorJSON))
		defer validatorsFile.Close()

		rsp, _ := cli.WithRunErrorsIgnored().RunOnly(
			"tx", poaModule, "update-validators",
			validatorsFile.Name(),
			"--from=node2",
			"--fees=1stake",
			"--gas=auto",
		)
		require.Contains(t, rsp, "invalid authority")
	})
}

func TestCreateValidator(t *testing.T) {
	// Scenario:
	// - Test creating validators with both ed25519 and secp256k1 keys
	// - Verify validators appear in the validators list (pending)
	// - Admin approves validators by updating power
	// - Verify validators are now active

	sut := systemtests.Sut

	// Enable secp256k1 pubkey type for consensus
	sut.ModifyGenesisJSON(t,
		SetConsensusValidatorPubKeyTypes(t, []string{"ed25519", "secp256k1"}),
	)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	// Create accounts for both validator types AFTER genesis is configured
	ed25519ValKeyName := "ed25519val"
	ed25519ValAddr := cli.AddKey(ed25519ValKeyName)

	secp256k1ValKeyName := "secp256k1val"
	secp256k1ValAddr := cli.AddKey(secp256k1ValKeyName)

	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", ed25519ValAddr, "10000000stake"},
		[]string{"genesis", "add-genesis-account", secp256k1ValAddr, "10000000stake"},
	)

	sut.StartChain(t)

	_, adminKeyName := getAdmin(t, cli)

	// Generate keys for both types
	ed25519PrivKey := ed25519.GenPrivKey()
	ed25519PubKey := ed25519PrivKey.PubKey()
	ed25519PkString := base64.StdEncoding.EncodeToString(ed25519PubKey.Bytes())

	secp256k1PrivKey := secp256k1.GenPrivKey()
	secp256k1PubKey := secp256k1PrivKey.PubKey()
	secp256k1PkString := base64.StdEncoding.EncodeToString(secp256k1PubKey.Bytes())

	t.Run("create ed25519 validator", func(t *testing.T) {
		// Create validator with ed25519 key
		rsp := cli.Run(
			"tx", poaModule, "create-validator",
			"ed25519-validator",
			ed25519PkString,
			"ed25519",
			"--description=Ed25519 test validator",
			"--from="+ed25519ValKeyName,
			"--gas=auto",
		)
		systemtests.RequireTxSuccess(t, rsp)

		// Query to verify the validator was created
		rsp = cli.CustomQuery("q", poaModule, "validator", ed25519ValAddr)
		validator := gjson.Get(rsp, "validator")
		require.True(t, validator.Exists(), "validator should exist")

		moniker := gjson.Get(rsp, "validator.metadata.moniker").String()
		assert.Equal(t, "ed25519-validator", moniker)

		// New validator should have 0 power initially
		power := gjson.Get(rsp, "validator.power").Int()
		assert.Equal(t, int64(0), power, "new validator should have 0 power initially")
	})

	t.Run("create secp256k1 validator", func(t *testing.T) {
		// Create validator with secp256k1 key
		rsp := cli.Run(
			"tx", poaModule, "create-validator",
			"secp256k1-validator",
			secp256k1PkString,
			"secp256k1",
			"--description=Secp256k1 test validator",
			"--from="+secp256k1ValKeyName,
			"--gas=auto",
		)
		systemtests.RequireTxSuccess(t, rsp)

		// Query to verify the validator was created
		rsp = cli.CustomQuery("q", poaModule, "validator", secp256k1ValAddr)
		validator := gjson.Get(rsp, "validator")
		require.True(t, validator.Exists(), "validator should exist")

		moniker := gjson.Get(rsp, "validator.metadata.moniker").String()
		assert.Equal(t, "secp256k1-validator", moniker)

		// New validator should have 0 power initially
		power := gjson.Get(rsp, "validator.power").Int()
		assert.Equal(t, int64(0), power, "new validator should have 0 power initially")
	})

	t.Run("admin activates both validators", func(t *testing.T) {
		rsp := cli.CustomQuery("q", poaModule, "total-power")
		totalPowerBefore := gjson.Get(rsp, "total_power").Int()

		ed25519Power := 1000
		secp256k1Power := 2000

		// Build validators JSON with both validators
		validatorsJSON := fmt.Sprintf(`[
			{
				"pub_key": {
					"@type": "/cosmos.crypto.ed25519.PubKey",
					"key": "%s"
				},
				"power": %d,
				"metadata": {
					"moniker": "ed25519-validator",
					"description": "Ed25519 test validator",
					"operator_address": "%s"
				}
			},
			{
				"pub_key": {
					"@type": "/cosmos.crypto.secp256k1.PubKey",
					"key": "%s"
				},
				"power": %d,
				"metadata": {
					"moniker": "secp256k1-validator",
					"description": "Secp256k1 test validator",
					"operator_address": "%s"
				}
			}
		]`,
			ed25519PkString,
			ed25519Power,
			ed25519ValAddr,
			secp256k1PkString,
			secp256k1Power,
			secp256k1ValAddr,
		)

		validatorsFile := systemtests.StoreTempFile(t, []byte(validatorsJSON))
		defer validatorsFile.Close()

		rsp, ok := cli.RunOnly(
			"tx", poaModule, "update-validators",
			validatorsFile.Name(),
			"--from="+adminKeyName,
			"--fees=1stake",
			"--gas=auto",
		)
		require.True(t, ok)
		systemtests.RequireTxSuccess(t, rsp)

		sut.AwaitNBlocks(t, 3)

		// Verify ed25519 validator now has power
		rsp = cli.CustomQuery("q", poaModule, "validator", ed25519ValAddr)
		power := gjson.Get(rsp, "validator.power").Int()
		assert.Equal(t, int64(ed25519Power), power, "ed25519 validator should now have power")

		// Verify secp256k1 validator now has power
		rsp = cli.CustomQuery("q", poaModule, "validator", secp256k1ValAddr)
		power = gjson.Get(rsp, "validator.power").Int()
		assert.Equal(t, int64(secp256k1Power), power, "secp256k1 validator should now have power")

		// Verify total power increased by both validators
		rsp = cli.CustomQuery("q", poaModule, "total-power")
		totalPowerAfter := gjson.Get(rsp, "total_power").Int()
		expectedTotal := totalPowerBefore + int64(ed25519Power) + int64(secp256k1Power)
		assert.Equal(t, expectedTotal, totalPowerAfter, "total power should include both new validators")
	})
}

func TestWithdrawFees(t *testing.T) {
	// Scenario:
	// - Run chain for several blocks to accumulate fees
	// - Query withdrawable fees for validator
	// - Validator withdraws fees
	// - Verify balance increased
	// - Verify withdrawable fees decreased

	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	// Add an account to generate transactions (and thus fees)
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "100000000stake"},
	)

	sut.StartChain(t)

	// Get validator operator address
	rsp := cli.CustomQuery("q", poaModule, "validators")
	validators := gjson.Get(rsp, "validators").Array()
	require.NotEmpty(t, validators)

	operators := make([]string, 0, len(validators))
	for _, val := range validators {
		op := gjson.Get(val.Raw, "metadata.operator_address").String()
		operators = append(operators, op)
	}

	recipient := operators[0]
	// Generate some transactions to create fees
	t.Log("Generating transactions to accumulate fees...")
	for i := 0; i < 5; i++ {
		rsp = cli.Run(
			"tx", "bank", "send",
			account1Addr, recipient, "10stake",
			"--from="+account1Addr,
			"--fees=100stake",
			"--gas=auto",
		)
		systemtests.RequireTxSuccess(t, rsp)
	}

	// Wait a few blocks for fee distribution
	sut.AwaitNBlocks(t, 3)

	expectedFee, err := types.ParseDecCoin("125.0stake")
	require.NoError(t, err)

	t.Run("check all fees", func(t *testing.T) {
		for _, operatorAddr := range operators {
			rsp = cli.CustomQuery("q", poaModule, "withdrawable-fees", operatorAddr)
			t.Logf("Withdrawable fees response: %s", rsp)
			// Fees should have accumulated
			fees := gjson.Get(rsp, "fees.fees").Array()
			require.NotEmpty(t, fees)
			amount, err := types.ParseDecCoin(fees[0].Str)
			require.NoError(t, err)

			require.True(t, amount.Equal(expectedFee))
		}
	})

	t.Run("validator withdraws fees", func(t *testing.T) {
		operatorAddr := operators[0]

		operatorKeyname := getKeyNameForAddress(t, cli, operatorAddr)
		require.NotEmpty(t, operatorKeyname, "should find key for validator operator")

		balBeforeInt := cli.QueryBalance(operatorAddr, "stake")
		balanceBefore := types.NewInt64DecCoin("stake", balBeforeInt)
		t.Logf("Balance Before: %s", balanceBefore.String())
		expectedBalanceAfter := balanceBefore.Add(expectedFee).Sub(types.NewInt64DecCoin("stake", 1))
		rsp = cli.Run(
			"tx", poaModule, "withdraw-fees",
			"--from="+operatorKeyname,
			"--gas=auto",
		)
		systemtests.RequireTxSuccess(t, rsp)

		// Check balance increased (accounting for tx fee)
		balanceAfterInt := cli.QueryBalance(operatorAddr, "stake")
		balanceAfter := types.NewInt64DecCoin("stake", balanceAfterInt)
		t.Logf("Balance After: %s", balanceAfter.String())
		require.True(t, expectedBalanceAfter.Equal(balanceAfter))
	})
}
