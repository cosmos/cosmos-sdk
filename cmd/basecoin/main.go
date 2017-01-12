package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strings"

	"github.com/tendermint/basecoin/app"
	. "github.com/tendermint/go-common"
	cfg "github.com/tendermint/go-config"
	"github.com/tendermint/go-logger"
	"github.com/tendermint/go-p2p"
	rpcserver "github.com/tendermint/go-rpc/server"
	eyes "github.com/tendermint/merkleeyes/client"
	tmcfg "github.com/tendermint/tendermint/config/tendermint"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	rpccore "github.com/tendermint/tendermint/rpc/core"
	tmtypes "github.com/tendermint/tendermint/types"
)

var safeRoutePoints = []string{
	"status",
	"genesis",
	"block",
	"blockchain",
	"validators",
	"dump_consensus_state",
	"broadcast_tx_sync",
	"num_unconfirmed_txs",
}

func main() {

	////////////////////////////////////
	// Load Basecoin

	eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
	genFilePath := flag.String("genesis", "", "Genesis file, if any")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr, "socket")
	if err != nil {
		Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin app
	app := app.NewBasecoin(eyesCli)

	// If genesis file was specified, set key-value options
	if *genFilePath != "" {
		kvz := loadGenesis(*genFilePath)
		for _, kv := range kvz {
			log := app.SetOption(kv.Key, kv.Value)
			fmt.Println(Fmt("Set %v=%v. Log: %v", kv.Key, kv.Value, log))
		}
	}

	////////////////////////////////////////////
	// Create & start tendermint node

	var config cfg.Config
	config = tmcfg.GetConfig("")

	privValidatorFile := config.GetString("priv_validator_file")
	privValidator := tmtypes.LoadOrGenPrivValidator(privValidatorFile)
	n := node.NewNode(config, privValidator, proxy.NewLocalClientCreator(app))

	protocol, address := node.ProtocolAndAddress(config.GetString("node_laddr"))
	l := p2p.NewDefaultListener(protocol, address, config.GetBool("skip_upnp"))
	n.AddListener(l)
	if err := n.Start(); err != nil {
		Exit(Fmt("Failed to start node: %v", err))
	}

	var log = logger.New("module", "main")
	log.Notice("Started node", "nodeInfo", n.NodeInfo())

	// If seedNode is provided by config, dial out.
	if config.GetString("seeds") != "" {
		seeds := strings.Split(config.GetString("seeds"), ",")
		n.DialSeeds(seeds)
	}

	// Run the tendermint RPC server.
	if config.GetString("rpc_laddr") != "" {
		_, err := StartRPC(n, config)
		if err != nil {
			PanicCrisis(err)
		}
	}

	// Start the listener
	//svr, err := server.NewServer(*addrPtr, "socket", app)
	//if err != nil {
	//	Exit("create listener: " + err.Error())
	//}

	// Wait forever
	TrapSignal(func() {
		// Cleanup
		n.Stop()
	})

}

// Only expose safe components to the internet
// Expose the rest to localhost
func StartRPC(n *node.Node, config cfg.Config) ([]net.Listener, error) {
	rpccore.SetConfig(config)
	rpccore.SetEventSwitch(n.EventSwitch())
	rpccore.SetBlockStore(n.BlockStore())
	rpccore.SetConsensusState(n.ConsensusState())
	rpccore.SetMempool(n.MempoolReactor().Mempool)
	rpccore.SetSwitch(n.Switch())
	//rpccore.SetPrivValidator(n.PrivValidator())
	rpccore.SetGenesisDoc(n.GenesisDoc())
	rpccore.SetProxyAppQuery(n.ProxyApp().Query())

	safeRoutes := make(map[string]*rpcserver.RPCFunc)
	for _, k := range safeRoutePoints {
		route, ok := rpccore.Routes[k]
		if !ok {
			PanicSanity(k)
		}
		safeRoutes[k] = route
	}

	var listeners []net.Listener

	listenAddrs := strings.Split(config.GetString("rpc_laddr"), ",")
	listenAddrSafe := listenAddrs[0]

	// the first listener is the public safe rpc
	mux := http.NewServeMux()
	rpcserver.RegisterRPCFuncs(mux, safeRoutes)
	listener, err := rpcserver.StartHTTPServer(listenAddrSafe, mux)
	if err != nil {
		return nil, err
	}
	listeners = append(listeners, listener)

	if len(listenAddrs) > 1 {
		listenAddrUnsafe := listenAddrs[1]
		// expose the full rpc
		mux := http.NewServeMux()
		wm := rpcserver.NewWebsocketManager(rpccore.Routes, n.EventSwitch())
		mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, rpccore.Routes)
		listener, err := rpcserver.StartHTTPServer(listenAddrUnsafe, mux)
		if err != nil {
			return nil, err
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}

//----------------------------------------

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func loadGenesis(filePath string) (kvz []KeyValue) {
	kvz_ := []interface{}{}
	bytes, err := ReadFile(filePath)
	if err != nil {
		Exit("loading genesis file: " + err.Error())
	}
	err = json.Unmarshal(bytes, &kvz_)
	if err != nil {
		Exit("parsing genesis file: " + err.Error())
	}
	if len(kvz_)%2 != 0 {
		Exit("genesis cannot have an odd number of items.  Format = [key1, value1, key2, value2, ...]")
	}
	for i := 0; i < len(kvz_); i += 2 {
		keyIfc := kvz_[i]
		valueIfc := kvz_[i+1]
		var key, value string
		key, ok := keyIfc.(string)
		if !ok {
			Exit(Fmt("genesis had invalid key %v of type %v", keyIfc, reflect.TypeOf(keyIfc)))
		}
		if value_, ok := valueIfc.(string); ok {
			value = value_
		} else {
			valueBytes, err := json.Marshal(valueIfc)
			if err != nil {
				Exit(Fmt("genesis had invalid value %v: %v", value_, err.Error()))
			}
			value = string(valueBytes)
		}
		kvz = append(kvz, KeyValue{key, value})
	}
	return kvz
}
