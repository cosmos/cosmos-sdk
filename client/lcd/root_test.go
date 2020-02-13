package lcd

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwaggerAsRootRoute(t *testing.T) {
	cmd := newTestCmd()
	go cmd.Execute()

	// Wait until is up
	time.Sleep(time.Millisecond * 200)

	res, err := http.Get("http://127.0.0.1:1317/")
	require.NoError(t, err)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	err = res.Body.Close()
	require.NoError(t, err)

	assert.True(t, strings.Contains(string(body), "<title>Swagger UI</title>"))
}

func newTestCmd() *cobra.Command {
	cdc := codec.New()
	cmd := ServeCommand(cdc, func(server *RestServer) {})
	viper.Set(flags.FlagListenAddr, "tcp://localhost:1317")

	return cmd
}
