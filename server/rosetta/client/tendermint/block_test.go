package tendermint

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_Block(t *testing.T) {
	fileData, err := ioutil.ReadFile("testdata/block.json")
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err2 := writer.Write(fileData)
		require.NoError(t, err2)
	}))
	defer s.Close()

	client := NewClient(s.URL)

	resp, err := client.Block(1)
	require.NoError(t, err)

	require.Equal(t, "2020-09-21T15:59:29.392015Z", resp.Block.Header.Time)
	require.Equal(t, "2939", resp.Block.Header.Height)
	require.Equal(t, "4D15DD40F8A29D5F9509C0DDCB12AE1AA6E99C290B0813DB19898334302A9EE0", resp.BlockID.Hash)
}

func TestClient_BlockByHash(t *testing.T) {
	fileData, err := ioutil.ReadFile("testdata/block_by_hash.json")
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err2 := writer.Write(fileData)
		require.NoError(t, err2)
	}))
	defer s.Close()

	client := NewClient(s.URL)

	resp, err := client.BlockByHash("0xA0590A598798973CA3224C089AFE4D7897AD150590F5E835A714BA58AA92D526")
	require.NoError(t, err)

	require.Equal(t, "2020-09-21T11:47:45.419592Z", resp.Block.Header.Time)
	require.Equal(t, "1", resp.Block.Header.Height)
	require.Equal(t, "A0590A598798973CA3224C089AFE4D7897AD150590F5E835A714BA58AA92D526", resp.BlockID.Hash)
}
