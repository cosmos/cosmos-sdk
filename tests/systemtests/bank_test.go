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
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	require.NotEqual(t, account1Addr, account2Addr)
	denom := "stake"
	initialAmount := 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, initialBalance},
		[]string{"genesis", "add-genesis-account", account2Addr, initialBalance},
	)
	sut.StartChain(t)

	// query accounts balances
	balance1 := cli.QueryBalance(account1Addr, denom)
	assert.Equal(t, int64(initialAmount), balance1)
	balance2 := cli.QueryBalance(account2Addr, denom)
	assert.Equal(t, int64(initialAmount), balance2)

	bankSendCmdArgs := []string{"tx", "bank", "send", account1Addr, account2Addr, "1000stake"}

	testCases := []struct {
		name         string
		extraArgs    []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			"valid transaction",
			[]string{"--fees=1stake"},
			false,
			0,
		},
		{
			"not enough fees",
			[]string{"--fees=0stake"},
			true,
			sdkerrors.ErrInsufficientFee.ABCICode(),
		},
		{
			"not enough gas",
			[]string{"--fees=1stake", "--gas=10"},
			true,
			sdkerrors.ErrOutOfGas.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmdArgs := append(bankSendCmdArgs, tc.extraArgs...)

			if tc.expectErr {
				assertErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					assert.Len(t, gotOutputs, 1)
					code := gjson.Get(gotOutputs[0].(string), "code")
					assert.True(t, code.Exists())
					assert.Equal(t, int64(tc.expectedCode), code.Int())
					return false // always abort
				}
				rsp := cli.WithRunErrorMatcher(assertErr).Run(cmdArgs...)
				RequireTxFailure(t, rsp)
			} else {
				rsp := cli.Run(cmdArgs...)
				txResult, found := cli.AwaitTxCommitted(rsp)
				assert.True(t, found)
				RequireTxSuccess(t, txResult)
			}
		})
	}

	// test tx bank send with insufficient funds
	insufficientCmdArgs := bankSendCmdArgs[0 : len(bankSendCmdArgs)-1]
	insufficientCmdArgs = append(insufficientCmdArgs, initialBalance, "--fees=10stake")
	rsp := cli.Run(insufficientCmdArgs...)
	RequireTxFailure(t, rsp)
	assert.Contains(t, rsp, sdkerrors.ErrInsufficientFunds.Error())

	// test tx bank send with unauthorized signature
	assertUnauthorizedErr := func(t assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		assert.Len(t, gotOutputs, 1)
		code := gjson.Get(gotOutputs[0].(string), "code")
		assert.True(t, code.Exists())
		assert.Equal(t, int64(sdkerrors.ErrUnauthorized.ABCICode()), code.Int())
		return false
	}
	invalidCli := cli
	invalidCli.chainID = cli.chainID + "a" // set invalid chain-id
	rsp = invalidCli.WithRunErrorMatcher(assertUnauthorizedErr).Run(bankSendCmdArgs...)
	RequireTxFailure(t, rsp)

	// test tx bank send generate only
	assertGenOnlyOutput := func(t assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		assert.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// get msg from output
		msgs := gjson.Get(rsp, "body.messages").Array()
		assert.Len(t, msgs, 1)
		// check from address is equal to account1 address
		fromAddr := gjson.Get(msgs[0].String(), "from_address").String()
		assert.Equal(t, account1Addr, fromAddr)
		// check to address is equal to account2 address
		toAddr := gjson.Get(msgs[0].String(), "to_address").String()
		assert.Equal(t, account2Addr, toAddr)
		return false
	}
	genCmdArgs := append(bankSendCmdArgs, "--generate-only")
	_ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(genCmdArgs...)

	// test tx bank send with dry-run flag
	assertDryRunOutput := func(t assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		assert.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// check gas estimate value found in output
		assert.Contains(t, rsp, "gas estimate")
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
	require.NotEqual(t, account1Addr, account2Addr, account3Addr)
	denom := "stake"
	initialAmount := 10000000
	initialBalance := fmt.Sprintf("%d%s", initialAmount, denom)
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, initialBalance},
		[]string{"genesis", "add-genesis-account", account2Addr, initialBalance},
		[]string{"genesis", "add-genesis-account", account3Addr, initialBalance},
	)
	sut.StartChain(t)

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
			"not enough fees",
			append(multiSendCmdArgs, "--fees=0stake"),
			true,
			sdkerrors.ErrInsufficientFee.ABCICode(),
			"insufficient fee",
		},
		{
			"not enough gas",
			append(multiSendCmdArgs, "--fees=1stake", "--gas=10"),
			true,
			sdkerrors.ErrOutOfGas.ABCICode(),
			"out of gas",
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
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if tc.expectErr {
				assertErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					assert.Len(t, gotOutputs, 1)
					output := gotOutputs[0].(string)
					assert.Contains(t, output, tc.expErrMsg)
					if tc.expectedCode != 0 {
						code := gjson.Get(output, "code")
						assert.True(t, code.Exists())
						assert.Equal(t, int64(tc.expectedCode), code.Int())
					}
					return false // always abort
				}
				_ = cli.WithRunErrorMatcher(assertErr).Run(tc.cmdArgs...)
			} else {
				rsp := cli.Run(tc.cmdArgs...)
				txResult, found := cli.AwaitTxCommitted(rsp)
				assert.True(t, found)
				RequireTxSuccess(t, txResult)
			}
		})
	}
}

func TestBankGRPCQueries(t *testing.T) {
	// scenario: test bank grpc gateway queries
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// update denom metadata in bank genesis
	atomDenomMetadata := `{"description":"The native staking token of the Cosmos Hub.","denom_units":[{"denom":"uatom","exponent":0,"aliases":["microatom"]},{"denom":"atom","exponent":6,"aliases":["ATOM"]}],"base":"uatom","display":"atom","name":"Cosmos Hub Atom","symbol":"ATOM","uri":"","uri_hash":""}`
	ethDenomMetadata := `{"description":"Ethereum mainnet token","denom_units":[{"denom":"wei","exponent":0,"aliases":[]},{"denom":"eth","exponent":6,"aliases":["ETH"]}],"base":"wei","display":"eth","name":"Ethereum","symbol":"ETH","uri":"","uri_hash":""}`

	bankDenomMetadata := fmt.Sprintf("[%s,%s]", atomDenomMetadata, ethDenomMetadata)

	sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		state, err := sjson.SetRawBytes(genesis, "app_state.bank.denom_metadata", []byte(bankDenomMetadata))
		assert.NoError(t, err)
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

	defaultExpSupplyOutput := `{"supply":[{"denom":"newdenom","amount":"10000000"},{"denom":"stake","amount":"2010000191"},{"denom":"testtoken","amount":"4000000000"}],"pagination":{"next_key":null,"total":"3"}}`
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
			defaultExpSupplyOutput,
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
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			assert.NoError(t, err)
			assert.Contains(t, string(resp), tc.expOut)
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
			`{"code":5,"message":"client metadata for denom foobar","details":[]}`,
		},
	}

	for _, tc := range dmTestCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequest(tc.url)
			assert.NoError(t, err)
			assert.Contains(t, string(resp), tc.expOut)
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
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequest(tc.url)
			assert.NoError(t, err)
			assert.Contains(t, string(resp), tc.expOut)
		})
	}
}
