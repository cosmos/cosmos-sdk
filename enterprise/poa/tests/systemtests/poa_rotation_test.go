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

//go:build system_test

package systemtests

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/tools/systemtests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const rotateCmd = "rotate-cons-pub-key"

// consKey bundles the representations of a freshly generated ed25519 consensus
// key needed by the rotation tests: the CLI base64 arg, the raw bytes for the
// CometBFT validator set lookup, and the bech32 consensus address for queries.
type consKey struct {
	Base64   string
	Bytes    []byte
	ConsAddr string
	// OperatorFromKey is the account address derived from this consensus key,
	// used only to exercise the operator-equals-consensus rejection path.
	OperatorFromKey string
}

func genConsKey(t *testing.T) consKey {
	t.Helper()
	pk := ed25519.GenPrivKey().PubKey()
	return consKey{
		Base64:          base64.StdEncoding.EncodeToString(pk.Bytes()),
		Bytes:           pk.Bytes(),
		ConsAddr:        sdk.GetConsAddress(pk).String(),
		OperatorFromKey: sdk.AccAddress(pk.Address()).String(),
	}
}

// createPOAValidator has the admin create an ed25519 validator with the given power.
func createPOAValidator(t *testing.T, cli *systemtests.CLIWrapper, adminKeyName, moniker, pkBase64, operatorAddr string, power int64) {
	t.Helper()
	rsp := cli.Run("tx", poaModule, "create-validator",
		moniker, pkBase64, "ed25519",
		"--operator-address="+operatorAddr,
		fmt.Sprintf("--power=%d", power),
		"--from="+adminKeyName, "--gas=auto",
	)
	systemtests.RequireTxSuccess(t, rsp)
}

// validatorPubKey returns the validator's current consensus pubkey base64 value.
func validatorPubKey(t *testing.T, cli *systemtests.CLIWrapper, addr string) string {
	t.Helper()
	rsp := cli.CustomQuery("q", poaModule, "validator", addr)
	return gjson.Get(rsp, "validator.pub_key.value").String()
}

func TestRotateConsPubKey(t *testing.T) {
	// One chain start; every subtest creates its own validators so they are
	// order-independent. Rotated validators are backed by generated keys that no
	// node signs for, so the genesis validators keep the chain producing blocks.
	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)

	// Funded operator accounts for the self-rotation subtests (operators pay their own fees).
	selfOpAddr := cli.AddKey("rot-self-op")
	feesOpAddr := cli.AddKey("rot-fees-op")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", selfOpAddr, "100000000stake"},
		[]string{"genesis", "add-genesis-account", feesOpAddr, "100000000stake"},
	)

	sut.StartChain(t)

	_, adminKeyName := getAdmin(t, cli)

	t.Run("operator rotates own key, old key leaves set and new key joins", func(t *testing.T) {
		oldKey := genConsKey(t)
		newKey := genConsKey(t)

		createPOAValidator(t, cli, adminKeyName, "self-rotate", oldKey.Base64, selfOpAddr, 1)
		sut.AwaitNBlocks(t, 3)

		// Old key is in the CometBFT validator set at the validator's power.
		require.Equal(t, int64(1), systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldKey.Bytes),
			"old consensus key should be in the CometBFT set before rotation")

		rsp := cli.Run("tx", poaModule, rotateCmd,
			newKey.Base64, "ed25519",
			"--operator-address="+selfOpAddr,
			"--from="+selfOpAddr, "--gas=auto",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNBlocks(t, 3)

		// Validator resolves by the new consensus address and by the unchanged operator address.
		rsp = cli.CustomQuery("q", poaModule, "validator", newKey.ConsAddr)
		require.True(t, gjson.Get(rsp, "validator").Exists(), "validator should resolve by new cons-addr")
		assert.Equal(t, selfOpAddr, gjson.Get(rsp, "validator.metadata.operator_address").String())
		assert.Equal(t, int64(1), gjson.Get(rsp, "validator.power").Int(), "power preserved across rotation")

		assert.Equal(t, newKey.Base64, validatorPubKey(t, cli, selfOpAddr), "operator now carries the new key")

		// Old key left the CometBFT set, new key joined at the same power.
		assert.Equal(t, int64(0), systemtests.QueryCometValidatorPower(sut.RPCClient(t), oldKey.Bytes),
			"old consensus key should leave the CometBFT set")
		assert.Equal(t, int64(1), systemtests.QueryCometValidatorPower(sut.RPCClient(t), newKey.Bytes),
			"new consensus key should join the CometBFT set at the same power")

		// Chain keeps producing blocks after the rotation.
		sut.AwaitNBlocks(t, 2)
	})

	t.Run("admin rotates another validators key", func(t *testing.T) {
		opAddr := cli.AddKey("rot-admin-op")
		oldKey := genConsKey(t)
		newKey := genConsKey(t)

		createPOAValidator(t, cli, adminKeyName, "admin-rotate", oldKey.Base64, opAddr, 1)
		sut.AwaitNextBlock(t)

		rsp := cli.Run("tx", poaModule, rotateCmd,
			newKey.Base64, "ed25519",
			"--operator-address="+opAddr,
			"--from="+adminKeyName, "--gas=auto",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)

		rsp = cli.CustomQuery("q", poaModule, "validator", newKey.ConsAddr)
		require.True(t, gjson.Get(rsp, "validator").Exists(), "validator should resolve by new cons-addr")
		assert.Equal(t, opAddr, gjson.Get(rsp, "validator.metadata.operator_address").String())
	})

	t.Run("fees survive rotation", func(t *testing.T) {
		oldKey := genConsKey(t)
		newKey := genConsKey(t)

		// Meaningful power so the validator accrues a visible share of the fees.
		createPOAValidator(t, cli, adminKeyName, "fees-rotate", oldKey.Base64, feesOpAddr, 10000)
		sut.AwaitNBlocks(t, 2)

		// Generate fees paid into the distribution pool.
		for i := 0; i < 5; i++ {
			rsp := cli.Run("tx", "bank", "send",
				selfOpAddr, feesOpAddr, "10stake",
				"--from="+selfOpAddr, "--fees=100stake", "--gas=auto",
			)
			systemtests.RequireTxSuccess(t, rsp)
		}
		sut.AwaitNBlocks(t, 3)

		rsp := cli.CustomQuery("q", poaModule, "withdrawable-fees", feesOpAddr)
		require.NotEmpty(t, gjson.Get(rsp, "fees.fees").Array(), "validator should have accrued fees before rotation")

		// Operator self-rotates.
		rsp = cli.Run("tx", poaModule, rotateCmd,
			newKey.Base64, "ed25519",
			"--operator-address="+feesOpAddr,
			"--from="+feesOpAddr, "--gas=auto",
		)
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNBlocks(t, 2)

		// Fees migrated with the rotation and remain withdrawable under the unchanged operator address.
		rsp = cli.CustomQuery("q", poaModule, "withdrawable-fees", feesOpAddr)
		require.NotEmpty(t, gjson.Get(rsp, "fees.fees").Array(), "fees should survive the rotation")

		balanceBefore := cli.QueryBalance(feesOpAddr, "stake")
		rsp = cli.Run("tx", poaModule, "withdraw-fees", "--from="+feesOpAddr, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		balanceAfter := cli.QueryBalance(feesOpAddr, "stake")
		assert.Greater(t, balanceAfter, balanceBefore, "withdrawing migrated fees should increase the balance")
	})

	t.Run("independent rotations in different blocks", func(t *testing.T) {
		opA := cli.AddKey("rot-indep-a")
		opB := cli.AddKey("rot-indep-b")
		keyA, keyB := genConsKey(t), genConsKey(t)
		newA, newB := genConsKey(t), genConsKey(t)

		createPOAValidator(t, cli, adminKeyName, "indep-a", keyA.Base64, opA, 1)
		createPOAValidator(t, cli, adminKeyName, "indep-b", keyB.Base64, opB, 1)
		sut.AwaitNextBlock(t)

		rsp := cli.Run("tx", poaModule, rotateCmd, newA.Base64, "ed25519",
			"--operator-address="+opA, "--from="+adminKeyName, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)

		rsp = cli.Run("tx", poaModule, rotateCmd, newB.Base64, "ed25519",
			"--operator-address="+opB, "--from="+adminKeyName, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)

		assert.Equal(t, newA.Base64, validatorPubKey(t, cli, opA))
		assert.Equal(t, newB.Base64, validatorPubKey(t, cli, opB))
	})

	t.Run("sequential chained rotation", func(t *testing.T) {
		op := cli.AddKey("rot-chain-op")
		key0, key1, key2 := genConsKey(t), genConsKey(t), genConsKey(t)

		createPOAValidator(t, cli, adminKeyName, "chain-rotate", key0.Base64, op, 1)
		sut.AwaitNextBlock(t)

		rsp := cli.Run("tx", poaModule, rotateCmd, key1.Base64, "ed25519",
			"--operator-address="+op, "--from="+adminKeyName, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
		assert.Equal(t, key1.Base64, validatorPubKey(t, cli, op))

		rsp = cli.Run("tx", poaModule, rotateCmd, key2.Base64, "ed25519",
			"--operator-address="+op, "--from="+adminKeyName, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)
		assert.Equal(t, key2.Base64, validatorPubKey(t, cli, op))

		// The freed intermediate cons-addr no longer resolves to a validator.
		rsp, _ = cli.WithRunErrorsIgnored().RunOnly("q", poaModule, "validator", key1.ConsAddr)
		assert.NotEqual(t, op, gjson.Get(rsp, "validator.metadata.operator_address").String(),
			"old cons-addr should no longer resolve to the operator")
	})

	t.Run("reuse a freed key", func(t *testing.T) {
		opA := cli.AddKey("rot-reuse-a")
		opB := cli.AddKey("rot-reuse-b")
		keyA, keyB, newA := genConsKey(t), genConsKey(t), genConsKey(t)

		createPOAValidator(t, cli, adminKeyName, "reuse-a", keyA.Base64, opA, 1)
		createPOAValidator(t, cli, adminKeyName, "reuse-b", keyB.Base64, opB, 1)
		sut.AwaitNextBlock(t)

		// Rotate A off keyA, freeing it.
		rsp := cli.Run("tx", poaModule, rotateCmd, newA.Base64, "ed25519",
			"--operator-address="+opA, "--from="+adminKeyName, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)

		// Rotate B onto the now-unused keyA. POA keeps no rotation lock, so this is allowed.
		rsp = cli.Run("tx", poaModule, rotateCmd, keyA.Base64, "ed25519",
			"--operator-address="+opB, "--from="+adminKeyName, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNextBlock(t)

		rsp = cli.CustomQuery("q", poaModule, "validator", keyA.ConsAddr)
		assert.Equal(t, opB, gjson.Get(rsp, "validator.metadata.operator_address").String(),
			"freed key should now belong to validator B")
	})

	t.Run("inactive power-0 validator rotation", func(t *testing.T) {
		op := cli.AddKey("rot-inactive-op")
		oldKey := genConsKey(t)
		newKey := genConsKey(t)

		createPOAValidator(t, cli, adminKeyName, "inactive-rotate", oldKey.Base64, op, 1)
		sut.AwaitNextBlock(t)

		// Set power to 0 so the validator leaves the active set.
		zero := validatorInfo{
			PubKeyType: "/cosmos.crypto.ed25519.PubKey", PubKeyKey: oldKey.Base64,
			OperatorAddr: op, Moniker: "inactive-rotate",
		}
		zeroFile := systemtests.StoreTempFile(t, []byte("["+validatorUpdateJSON(zero, 0)+"]"))
		rsp := cli.Run("tx", poaModule, "update-validators", zeroFile.Name(),
			"--from="+adminKeyName, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNBlocks(t, 2)

		rsp = cli.Run("tx", poaModule, rotateCmd, newKey.Base64, "ed25519",
			"--operator-address="+op, "--from="+adminKeyName, "--gas=auto")
		systemtests.RequireTxSuccess(t, rsp)
		sut.AwaitNBlocks(t, 2)

		// Record re-keyed and still power 0; new key never enters the CometBFT set.
		rsp = cli.CustomQuery("q", poaModule, "validator", newKey.ConsAddr)
		require.True(t, gjson.Get(rsp, "validator").Exists())
		assert.Equal(t, int64(0), gjson.Get(rsp, "validator.power").Int())
		assert.Equal(t, int64(0), systemtests.QueryCometValidatorPower(sut.RPCClient(t), newKey.Bytes),
			"power-0 validator must not enter the CometBFT set")
	})

	t.Run("rejections", func(t *testing.T) {
		opA := cli.AddKey("rot-rej-a")
		opB := cli.AddKey("rot-rej-b")
		keyA, keyB := genConsKey(t), genConsKey(t)

		createPOAValidator(t, cli, adminKeyName, "rej-a", keyA.Base64, opA, 1)
		createPOAValidator(t, cli, adminKeyName, "rej-b", keyB.Base64, opB, 1)
		sut.AwaitNextBlock(t)

		t.Run("new key already in use by another validator", func(t *testing.T) {
			rsp, _ := cli.WithRunErrorsIgnored().RunOnly("tx", poaModule, rotateCmd,
				keyB.Base64, "ed25519",
				"--operator-address="+opA, "--from="+adminKeyName, "--gas=auto")
			require.Contains(t, rsp, "already in use")
		})

		t.Run("no-op rotation to current key", func(t *testing.T) {
			rsp, _ := cli.WithRunErrorsIgnored().RunOnly("tx", poaModule, rotateCmd,
				keyA.Base64, "ed25519",
				"--operator-address="+opA, "--from="+adminKeyName, "--gas=auto")
			require.Contains(t, rsp, "nothing to rotate")
		})

		t.Run("unauthorized sender", func(t *testing.T) {
			// selfOpAddr is funded but is neither opA nor the admin.
			rsp, _ := cli.WithRunErrorsIgnored().RunOnly("tx", poaModule, rotateCmd,
				genConsKey(t).Base64, "ed25519",
				"--operator-address="+opA, "--from="+selfOpAddr, "--gas=auto")
			require.Contains(t, rsp, "neither the validator operator nor the admin")
		})

		t.Run("unknown validator", func(t *testing.T) {
			unknownOp := cli.AddKey("rot-unknown-op")
			rsp, _ := cli.WithRunErrorsIgnored().RunOnly("tx", poaModule, rotateCmd,
				genConsKey(t).Base64, "ed25519",
				"--operator-address="+unknownOp, "--from="+adminKeyName, "--gas=auto")
			require.Contains(t, rsp, "unknown validator")
		})

		t.Run("pubkey type not in consensus params", func(t *testing.T) {
			secpPk := base64.StdEncoding.EncodeToString(secp256k1.GenPrivKey().PubKey().Bytes())
			rsp, _ := cli.WithRunErrorsIgnored().RunOnly("tx", poaModule, rotateCmd,
				secpPk, "secp256k1",
				"--operator-address="+opA, "--from="+adminKeyName, "--gas=auto")
			require.Contains(t, rsp, "consensus parameters")
		})

		t.Run("operator derives from new consensus pubkey", func(t *testing.T) {
			// Create a validator whose operator address derives from derivedKey,
			// then rotate it to derivedKey so operator and consensus keys collide.
			derivedKey := genConsKey(t)
			otherKey := genConsKey(t)
			createPOAValidator(t, cli, adminKeyName, "rej-derived", otherKey.Base64, derivedKey.OperatorFromKey, 1)
			sut.AwaitNextBlock(t)

			rsp, _ := cli.WithRunErrorsIgnored().RunOnly("tx", poaModule, rotateCmd,
				derivedKey.Base64, "ed25519",
				"--operator-address="+derivedKey.OperatorFromKey, "--from="+adminKeyName, "--gas=auto")
			require.Contains(t, rsp, "different keys")
		})
	})
}
