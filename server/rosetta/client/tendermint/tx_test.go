package tendermint

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_Tx(t *testing.T) {
	fileData, err := ioutil.ReadFile("testdata/tx.json")
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err2 := writer.Write(fileData)
		require.NoError(t, err2)

	}))
	defer s.Close()

	client := NewClient(s.URL)

	resp, err := client.Tx("0xF38E833151DFD041BE7CD8906C2C46CAA2FF2945D766CEFEEB4BC3C43F0AFEDF")
	require.NoError(t, err)

	require.Equal(t, "F38E833151DFD041BE7CD8906C2C46CAA2FF2945D766CEFEEB4BC3C43F0AFEDF", resp.Hash)
}
