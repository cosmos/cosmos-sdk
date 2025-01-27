package systemtests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
)

func TestAccountCreation(t *testing.T) {
	// scenario: test account creation
	// given a running chain
	// when accountA is sending funds to accountB,
	// AccountB should not be created
	// when accountB is sending funds to accountA,
	// AccountB should be created

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	require.NotEqual(t, account1Addr, account2Addr)
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	systest.Sut.StartChain(t)

	// query account1
	rsp := cli.CustomQuery("q", "auth", "account", account1Addr)
	assert.Equal(t, account1Addr, gjson.Get(rsp, "account.value.address").String(), rsp)

	rsp1 := cli.RunAndWait("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake")
	systest.RequireTxSuccess(t, rsp1)

	// query account2
	assertNotFound := func(t assert.TestingT, err error, msgAndArgs ...interface{}) (ok bool) {
		return strings.Contains(err.Error(), "not found: key not found")
	}
	_ = cli.WithRunErrorMatcher(assertNotFound).CustomQuery("q", "auth", "account", account2Addr)

	rsp3 := cli.RunAndWait("tx", "bank", "send", account2Addr, account1Addr, "1000stake", "--from="+account2Addr, "--fees=1stake")
	systest.RequireTxSuccess(t, rsp3)

	// query account2 to make sure its created
	rsp4 := cli.CustomQuery("q", "auth", "account", account2Addr)
	assert.Equal(t, "1", gjson.Get(rsp4, "account.value.sequence").String(), rsp4)
	rsp5 := cli.CustomQuery("q", "auth", "account", account1Addr)
	assert.Equal(t, "1", gjson.Get(rsp5, "account.value.sequence").String(), rsp5)
}

func TestAccountsMigration(t *testing.T) {
	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	legacyAddress := cli.GetKeyAddr("node0")
	// Create a receiver account
	receiverName := "receiver-account"
	receiverAddress := cli.AddKey(receiverName)
	require.NotEmpty(t, receiverAddress)
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", receiverAddress, "1000000stake"},
	)

	systest.Sut.StartChain(t)

	// Get pubkey
	pubKeyValue := cli.GetPubKeyByCustomField(legacyAddress, "address")
	require.NotEmpty(t, pubKeyValue, "Public key for legacy-account not found")

	// 1. Verify the account
	rsp := cli.CustomQuery("q", "auth", "account", legacyAddress)
	require.NotEmpty(t, rsp)
	accountInfo := gjson.Parse(rsp)
	addressFromResponse := accountInfo.Get("account.value.address").String()
	require.Equal(t, legacyAddress, addressFromResponse, "The address in the response should match the legacy address")

	// 2. Migrate this account from x/auth to x/accounts

	// Verify the account not exist in account
	rsp = cli.WithRunErrorsIgnored().CustomQuery("q", "accounts", "query", legacyAddress, "cosmos.accounts.defaults.base.v1.QuerySequence", "{}")
	require.Contains(t, rsp, "not found: key")

	accountInitMsg := fmt.Sprintf(`{
	   "@type": "/cosmos.accounts.defaults.base.v1.MsgInit",
	   "pub_key": {
	       "@type": "/cosmos.crypto.secp256k1.PubKey",
	       "key": "%s"
	   },
	   "init_sequence": "0"
	}`, pubKeyValue)

	rsp = cli.RunAndWait("tx", "auth", "migrate-account",
		"--account-type=base",
		fmt.Sprintf("--account-init-msg=%s", accountInitMsg),
		fmt.Sprintf("--from=%s", legacyAddress),
		"--fees=1stake")
	systest.RequireTxSuccess(t, rsp)

	// 3. Now the account should be existed, query the account Sequence
	rsp = cli.CustomQuery("q", "accounts", "query", legacyAddress, "cosmos.accounts.defaults.base.v1.QuerySequence", "{}")
	sequence := gjson.Get(rsp, "response.sequence").Exists()
	require.True(t, sequence, "Sequence field should exist")

	// 4. Execute a transaction using the bank module

	// Check initial balances
	legacyBalance := cli.QueryBalance(legacyAddress, "stake")
	receiverBalance := cli.QueryBalance(receiverAddress, "stake")
	require.Equal(t, int64(399999999), legacyBalance) // 20000000 - 1 (fee for migration)
	require.Equal(t, int64(1000000), receiverBalance)

	transferAmount := "1000000"
	rsp = cli.RunAndWait("tx", "bank", "send",
		legacyAddress,
		receiverAddress,
		transferAmount+"stake",
		fmt.Sprintf("--from=%s", legacyAddress),
		"--fees=1stake")
	systest.RequireTxSuccess(t, rsp)

	// Verify the balances after the transaction
	newLegacyBalance := cli.QueryBalance(legacyAddress, "stake")
	newReceiverBalance := cli.QueryBalance(receiverAddress, "stake")

	expectedLegacyBalance := legacyBalance - 1000000 - 1 // Initial balance - transferred amount - fee
	expectedReceiverBalance := receiverBalance + 1000000 // Initial balance + transferred amount

	require.Equal(t, expectedLegacyBalance, newLegacyBalance, "Legacy account balance is incorrect after transfer")
	require.Equal(t, expectedReceiverBalance, newReceiverBalance, "Receiver account balance is incorrect after transfer")

	// 5. Test swapKey functionality
	newKeyName := "new-key"
	newKeyAddress := cli.AddKey(newKeyName)
	require.NotEmpty(t, newKeyAddress)

	newPubKey := cli.GetPubKeyByCustomField(newKeyAddress, "address")
	require.NotEmpty(t, newPubKey, "Public key for new-key not found")

	swapKeyMsg := fmt.Sprintf(`{
		"new_pub_key": {
			"@type": "/cosmos.crypto.secp256k1.PubKey",
			"key": "%s"
		}
	}`, newPubKey)

	rsp = cli.RunAndWait("tx", "accounts", "execute", legacyAddress, "cosmos.accounts.defaults.base.v1.MsgSwapPubKey", swapKeyMsg,
		fmt.Sprintf("--from=%s", legacyAddress),
		"--fees=1stake")
	systest.RequireTxSuccess(t, rsp)
}
