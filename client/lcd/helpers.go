package lcd

// NOTE: COPIED VERBATIM FROM tendermint/tendermint/rpc/test/helpers.go

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cmn "github.com/tendermint/tmlibs/common"

	cfg "github.com/tendermint/tendermint/config"
)

var globalConfig *cfg.Config

// f**ing long, but unique for each test
func makePathname() string {
	// get path
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// fmt.Println(p)
	sep := string(filepath.Separator)
	return strings.Replace(p, sep, "_", -1)
}

func randPort() int {
	return int(cmn.RandUint16()/2 + 10000)
}

func makeAddrs() (string, string, string) {
	start := randPort()
	return fmt.Sprintf("tcp://0.0.0.0:%d", start),
		fmt.Sprintf("tcp://0.0.0.0:%d", start+1),
		fmt.Sprintf("tcp://0.0.0.0:%d", start+2)
}

// GetConfig returns a config for the test cases as a singleton
func GetConfig() *cfg.Config {
	if globalConfig == nil {
		pathname := makePathname()
		globalConfig = cfg.ResetTestRoot(pathname)

		// and we use random ports to run in parallel
		tm, rpc, _ := makeAddrs()
		globalConfig.P2P.ListenAddress = tm
		globalConfig.RPC.ListenAddress = rpc
		globalConfig.TxIndex.IndexTags = "app.creator" // see kvstore application
	}
	return globalConfig
}
