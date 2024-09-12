//go:build system_test

package systemtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func TestBankSendTxCmd(t *testing.T) {
	// scenario: test bank send command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	sut.StartChain(t)

	// query validator balance and make sure it has enough balance
	var transferAmount int64 = 1000
	valBalance := cli.QueryBalance(valAddr, denom)
	require.Greater(t, valBalance, transferAmount, "not enough balance found with validator")

	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom)}

	// test valid transaction
	rsp := cli.Run(append(bankSendCmdArgs, "--fees=1stake")...)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)
	// check valaddr balance equals to valBalance-(transferedAmount+feeAmount)
	require.Equal(t, valBalance-(transferAmount+1), cli.QueryBalance(valAddr, denom))
	// check receiver balance equals to transferAmount
	require.Equal(t, transferAmount, cli.QueryBalance(receiverAddr, denom))

	// test tx bank send with insufficient funds
	insufficientCmdArgs := bankSendCmdArgs[0 : len(bankSendCmdArgs)-1]
	insufficientCmdArgs = append(insufficientCmdArgs, fmt.Sprintf("%d%s", valBalance, denom), "--fees=10stake")
	rsp = cli.Run(insufficientCmdArgs...)
	RequireTxFailure(t, rsp)
	require.Contains(t, rsp, sdkerrors.ErrInsufficientFunds.Error())

	// test tx bank send with unauthorized signature
	assertUnauthorizedErr := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		code := gjson.Get(gotOutputs[0].(string), "code")
		require.True(t, code.Exists())
		require.Equal(t, int64(sdkerrors.ErrUnauthorized.ABCICode()), code.Int())
		return false
	}
	invalidCli := cli
	invalidCli.chainID = cli.chainID + "a" // set invalid chain-id
	rsp = invalidCli.WithRunErrorMatcher(assertUnauthorizedErr).Run(bankSendCmdArgs...)
	RequireTxFailure(t, rsp)

	// test tx bank send generate only
	assertGenOnlyOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// get msg from output
		msgs := gjson.Get(rsp, "body.messages").Array()
		require.Len(t, msgs, 1)
		// check from address is equal to account1 address
		fromAddr := gjson.Get(msgs[0].String(), "from_address").String()
		require.Equal(t, valAddr, fromAddr)
		// check to address is equal to account2 address
		toAddr := gjson.Get(msgs[0].String(), "to_address").String()
		require.Equal(t, receiverAddr, toAddr)
		return false
	}
	genCmdArgs := append(bankSendCmdArgs, "--generate-only")
	_ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(genCmdArgs...)

	// test tx bank send with dry-run flag
	assertDryRunOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// check gas estimate value found in output
		require.Contains(t, rsp, "gas estimate")
		return false
	}
	dryRunCmdArgs := append(bankSendCmdArgs, "--dry-run")
	_ = cli.WithRunErrorMatcher(assertDryRunOutput).Run(dryRunCmdArgs...)
}

func TestBankMultiSendTxCmd(t *testing.T) {
	// scenario: test bank multi-send command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	account3Addr := cli.AddKey("account3")
	require.NotEqual(t, account1Addr, account2Addr)
	require.NotEqual(t, account1Addr, account3Addr)
	denom := "stake"
	var initialAmount int64 = 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, initialBalance},
		[]string{"genesis", "add-genesis-account", account2Addr, initialBalance},
	)
	sut.StartChain(t)

	// query accounts balances
	account1Bal := cli.QueryBalance(account1Addr, denom)
	require.Equal(t, initialAmount, account1Bal)
	account2Bal := cli.QueryBalance(account2Addr, denom)
	require.Equal(t, initialAmount, account2Bal)
	var account3Bal int64 = 0

	multiSendCmdArgs := []string{"tx", "bank", "multi-send", account1Addr, account2Addr, account3Addr, "1000stake", "--from=" + account1Addr}

	testCases := []struct {
		name         string
		cmdArgs      []string
		expectErr    bool
		expectedCode uint32
		expErrMsg    string
	}{
		{
			"valid transaction",
			append(multiSendCmdArgs, "--fees=1stake"),
			false,
			0,
			"",
		},
		{
			"not enough arguments",
			[]string{"tx", "bank", "multi-send", account1Addr, account2Addr, "1000stake", "--from=" + account1Addr},
			true,
			0,
			"only received 3",
		},
		{
			"chain-id shouldn't be used with offline and generate-only flags",
			append(multiSendCmdArgs, "--generate-only", "--offline", "-a=0", "-s=4"),
			true,
			0,
			"chain ID cannot be used",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectErr {
				assertErr := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					require.Len(t, gotOutputs, 1)
					output := gotOutputs[0].(string)
					require.Contains(t, output, tc.expErrMsg)
					if tc.expectedCode != 0 {
						code := gjson.Get(output, "code")
						require.True(t, code.Exists())
						require.Equal(t, int64(tc.expectedCode), code.Int())
					}
					return false // always abort
				}
				_ = cli.WithRunErrorMatcher(assertErr).Run(tc.cmdArgs...)
			} else {
				rsp := cli.Run(tc.cmdArgs...)
				txResult, found := cli.AwaitTxCommitted(rsp)
				require.True(t, found)
				RequireTxSuccess(t, txResult)
				// check account1 balance equals to account1Bal - transferredAmount*no_of_accounts - fees
				expAcc1Balance := account1Bal - (1000 * 2) - 1
				require.Equal(t, expAcc1Balance, cli.QueryBalance(account1Addr, denom))
				account1Bal = expAcc1Balance
				// check account2 balance equals to account2Bal + transferredAmount
				expAcc2Balance := account2Bal + 1000
				require.Equal(t, expAcc2Balance, cli.QueryBalance(account2Addr, denom))
				account2Bal = expAcc2Balance
				// check account3 balance equals to account3Bal + transferredAmount
				expAcc3Balance := account3Bal + 1000
				require.Equal(t, expAcc3Balance, cli.QueryBalance(account3Addr, denom))
				account3Bal = expAcc3Balance
			}
		})
	}
}

func TestBankGRPCQueries(t *testing.T) {
	// scenario: test bank grpc gateway queries
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// update bank denom metadata in genesis
	atomDenomMetadata := `{"description":"The native staking token of the Cosmos Hub.","denom_units":[{"denom":"uatom","exponent":0,"aliases":["microatom"]},{"denom":"atom","exponent":6,"aliases":["ATOM"]}],"base":"uatom","display":"atom","name":"Cosmos Hub Atom","symbol":"ATOM","uri":"","uri_hash":""}`
	ethDenomMetadata := `{"description":"Ethereum mainnet token","denom_units":[{"denom":"wei","exponent":0,"aliases":[]},{"denom":"eth","exponent":6,"aliases":["ETH"]}],"base":"wei","display":"eth","name":"Ethereum","symbol":"ETH","uri":"","uri_hash":""}`

	bankDenomMetadata := fmt.Sprintf("[%s,%s]", atomDenomMetadata, ethDenomMetadata)

	sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		state, err := sjson.SetRawBytes(genesis, "app_state.bank.denom_metadata", []byte(bankDenomMetadata))
		require.NoError(t, err)
		return state
	})

	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	newDenom := "newdenom"
	initialAmount := "10000000"
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake," + initialAmount + newDenom},
	)

	// start chain
	sut.StartChain(t)
	baseurl := fmt.Sprintf("http://localhost:%d", apiPortStart)

	// test supply grpc endpoint
	supplyUrl := baseurl + "/cosmos/bank/v1beta1/supply"

	// as supply might change for each block, can't set complete expected output
	expTotalSupplyOutput := `{"supply":[{"denom":"newdenom","amount":"10000000"},{"denom":"stake","amount"`
	specificDenomOutput := fmt.Sprintf(`{"denom":"%s","amount":"%s"}`, newDenom, initialAmount)
	bogusDenomOutput := `{"denom":"foobar","amount":"0"}`

	blockHeightHeader := "x-cosmos-block-height"
	blockHeight := sut.CurrentHeight()

	supplyTestCases := []struct {
		name    string
		url     string
		headers map[string]string
		expOut  string
	}{
		{
			"test GRPC total supply",
			supplyUrl,
			map[string]string{
				blockHeightHeader: fmt.Sprintf("%d", blockHeight),
			},
			expTotalSupplyOutput,
		},
		{
			"test GRPC total supply of a specific denom",
			supplyUrl + "/by_denom?denom=" + newDenom,
			map[string]string{},
			specificDenomOutput,
		},
		{
			"error when querying supply with height greater than block height",
			supplyUrl,
			map[string]string{
				blockHeightHeader: fmt.Sprintf("%d", blockHeight+5),
			},
			"invalid height",
		},
		{
			"test GRPC total supply of a bogus denom",
			supplyUrl + "/by_denom?denom=foobar",
			map[string]string{},
			bogusDenomOutput,
		},
	}

	for _, tc := range supplyTestCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			require.NoError(t, err)
			require.Contains(t, string(resp), tc.expOut)
		})
	}

	// test denom metadata endpoint
	denomMetadataUrl := baseurl + "/cosmos/bank/v1beta1/denoms_metadata"
	dmTestCases := []struct {
		name   string
		url    string
		expOut string
	}{
		{
			"test GRPC client metadata",
			denomMetadataUrl,
			bankDenomMetadata,
		},
		{
			"test GRPC client metadata of a specific denom",
			denomMetadataUrl + "/uatom",
			atomDenomMetadata,
		},
		{
			"test GRPC client metadata of a bogus denom",
			denomMetadataUrl + "/foobar",
			`"details":[]`,
		},
	}

	for _, tc := range dmTestCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequest(tc.url)
			require.NoError(t, err)
			require.Contains(t, string(resp), tc.expOut)
		})
	}

	// test bank balances endpoint
	balanceUrl := baseurl + "/cosmos/bank/v1beta1/balances/"
	allBalancesOutput := `{"balances":[` + specificDenomOutput + `,{"denom":"stake","amount":"10000000"}],"pagination":{"next_key":null,"total":"2"}}`

	balanceTestCases := []struct {
		name   string
		url    string
		expOut string
	}{
		{
			"test GRPC total account balance",
			balanceUrl + account1Addr,
			allBalancesOutput,
		},
		{
			"test GRPC account balance of a specific denom",
			fmt.Sprintf("%s%s/by_denom?denom=%s", balanceUrl, account1Addr, newDenom),
			specificDenomOutput,
		},
		{
			"test GRPC account balance of a bogus denom",
			fmt.Sprintf("%s%s/by_denom?denom=foobar", balanceUrl, account1Addr),
			bogusDenomOutput,
		},
	}

	for _, tc := range balanceTestCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequest(tc.url)
			require.NoError(t, err)
			require.Contains(t, string(resp), tc.expOut)
		})
	}
}
