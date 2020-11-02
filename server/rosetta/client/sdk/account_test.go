package sdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthAccountClient(t *testing.T) {
	client, err := NewClient("0.0.0.0:9090")
	require.NoError(t, err)

	account, err := client.GetAuthAccount(context.Background(), "cosmos1y7jgzf0q4vvun8tlgc7lf5jj35anytrjsy9hv8", 0)
	require.NoError(t, err)

	fmt.Printf("account %v", account)
	//bz, err := ioutil.ReadFile("testdata/account.json")
	//require.NoError(t, err)
	//
	//s := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	//	_, err = writer.Write(bz)
	//	require.NoError(t, err)
	//}))
	//defer s.Close()
	//
	//cdc, _ := simapp.MakeCodecs()
	//client := NewClient(s.URL, cdc)
	//
	//addr := "cosmos15lc6l4nm3s9ya5an5vnv9r6na437ajpznkplhx"
	//res, err := client.GetAuthAccount(context.Background(), addr, 0)
	//require.NoError(t, err)
	//require.NotNil(t, res)
	//
	//require.Equal(t, int64(9694), res.Height)
	//require.Equal(t, addr, res.Result.Value.Address)
	//require.Equal(t, "2", res.Result.Value.AccountNumber)
	//require.Equal(t, "4", res.Result.Value.Sequence)
	//require.Equal(t, int64(1000), res.Result.Value.Coins[0].Amount.Int64())
}
