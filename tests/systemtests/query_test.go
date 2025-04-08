//go:build system_test

package systemtests

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/systemtests"
)

func TestRestQueries(t *testing.T) {
	// Scenario:
	// delegate tokens to validator
	// undelegate some tokens
	sut := systemtests.Sut
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, systemtests.Verbose)
	sut.StartChain(t)
	startHeight := sut.CurrentHeight()
	sut.AwaitNextBlock(t)

	queryUrl := fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/withdraw_address", sut.APIAddress(), cli.GetKeyAddr("node0"))
	_, headers, err := doRequest(queryUrl, map[string]string{})
	// then
	require.NoError(t, err)
	t.Logf("headers: %v", headers)

	const heightHeaderKey = "X-Cosmos-Block-Height"
	require.Contains(t, headers, heightHeaderKey)
	gotHeight, err := strconv.Atoi(headers[heightHeaderKey][0])
	require.NoError(t, err)
	assert.GreaterOrEqual(t, gotHeight, int(startHeight+1))

	// and when called with height header
	_, headers, err = doRequest(queryUrl, map[string]string{"X-Cosmos-Block-Height": strconv.Itoa(int(startHeight))})
	// then
	require.NoError(t, err)
	t.Logf("headers: %v", headers)
	require.Contains(t, headers, heightHeaderKey)

	gotHeight, err = strconv.Atoi(headers[heightHeaderKey][0])
	require.NoError(t, err)
	assert.Equal(t, int(startHeight), gotHeight)
}

func doRequest(url string, headers map[string]string) ([]byte, http.Header, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	if err = res.Body.Close(); err != nil {
		return nil, nil, err
	}
	fmt.Printf("headers: %v\n", res.Header)
	return body, res.Header, nil
}
