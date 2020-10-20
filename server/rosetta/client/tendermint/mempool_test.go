package tendermint

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_UnconfirmedTxs(t *testing.T) {
	fileData, err := ioutil.ReadFile("testdata/unconfirmed_txs.json")
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err2 := writer.Write(fileData)
		require.NoError(t, err2)
	}))
	defer s.Close()

	client := NewClient(s.URL)

	resp, err := client.UnconfirmedTxs()
	require.NoError(t, err)

	require.Equal(t, []string{
		"swEoKBapCjuoo2GaChQHHvELd6j+pCLvyEBKrpOmC1tAqBIUkx6VOgaY9RvG0j2wHdnQykFZbSAaCQoEYXRvbRIBMRIEEMCaDBpqCibrWumHIQLMAYSOINcDLMQWfMl/W3hM/gNWwsfgG8aEMkNaBA1VXxJA2PNOBc2szg27zyPFnZjOdLuAzaXGdY3kc4SwJNJgAAwRA6VVW5QUA8FZICcn6Gq/3iaTGKDnBHId/e8mg8a5FQ==",
	}, resp.Txs)
}
