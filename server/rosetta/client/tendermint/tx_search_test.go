package tendermint

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_TxSearch(t *testing.T) {
	fileData, err := ioutil.ReadFile("testdata/tx_search.json")
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err2 := writer.Write(fileData)
		require.NoError(t, err2)

	}))
	defer s.Close()

	client := NewClient(s.URL)

	resp, err := client.TxSearch(fmt.Sprintf(`tx.height=%s`, "19176"))
	require.NoError(t, err)

	require.Equal(t, "1", resp.TotalCount)
	require.Equal(t, "F38E833151DFD041BE7CD8906C2C46CAA2FF2945D766CEFEEB4BC3C43F0AFEDF", resp.Txs[0].Hash)
}
