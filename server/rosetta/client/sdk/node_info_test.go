package sdk

import (
	"testing"
)

func TestGetNodeInfo(t *testing.T) {
	//bz, err := ioutil.ReadFile("testdata/nodeinfo.json")
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
	//moniker := "mynode"
	//res, err := client.GetNodeInfo(context.Background())
	//t.Log(res)
	//
	//require.NoError(t, err)
	//require.NotNil(t, res)
	//require.Equal(t, moniker, res.Moniker)
	//require.Equal(t, "0.33.7", res.Version)
}
