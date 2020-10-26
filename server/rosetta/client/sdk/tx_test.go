package sdk

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/cosmos/cosmos-sdk/x/auth"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/go-amino"

	"github.com/stretchr/testify/require"
)

func TestGetTx(t *testing.T) {
	bz, err := ioutil.ReadFile("testdata/validtx.json")
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err = writer.Write(bz)
		require.NoError(t, err)
	}))
	defer s.Close()

	cdc := amino.NewCodec()
	sdk.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	client := NewClient(s.URL, cdc)

	hash := "CFFE3295A82BC0104F1175C26384235B6B3DA80306597F8590684282E195EF1C"
	res, err := client.GetTx(context.Background(), hash)
	t.Log(res)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, hash, res.TxHash)
}
