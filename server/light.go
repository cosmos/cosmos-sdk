package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/log"
	tmmath "github.com/tendermint/tendermint/libs/math"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/light"
	lproxy "github.com/tendermint/tendermint/light/proxy"
	lrpc "github.com/tendermint/tendermint/light/rpc"
	dbs "github.com/tendermint/tendermint/light/store/db"
	rpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"
	dbm "github.com/tendermint/tm-db"
)

// simd light cosmoshub-3 --primary-addr http://193.26.156.221:26657/ --witness-addr http://144.76.61.201:26657/ --trusted-height 5940895 --trusted-hash 8663FBD3FB9DCE3D8E461EA521C38256F6EAF85D4FA492BAE26D5863F53CA150

const (
	listenAddr         = "listening-address"
	primaryAddr        = "primary-addr"
	witnessAddrsJoined = "witness-addr"
	dir                = "dir"
	maxOpenConnections = "max-open-connections"

	sequential     = "sequential-verification"
	trustingPeriod = "trust-period"
	trustedHeight  = "trusted-height"
	trustedHash    = "trusted-hash"
	trustLevel     = "trust-level"

	logLevel = "log-level"
)

var (
	primaryKey   = []byte("primary")
	witnessesKey = []byte("witnesses")
)

func LightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light [chainID]",
		Short: "Run a light client proxy server, verifying Tendermint rpc",
		Long: `Run a light client proxy server, verifying Tendermint rpc.

All calls that can be tracked back to a block header by a proof
will be verified before passing them back to the caller. Other than
that, it will present the same interface as a full Tendermint node.

Furthermore to the chainID, a fresh instance of a light client will
need a primary RPC address, a trusted hash and height and witness RPC addresses
(if not using sequential verification). To restart the node, thereafter
only the chainID is required.
`,
		RunE: runProxy,
		Args: cobra.ExactArgs(1),
		Example: `light cosmoshub-4 -primary-addr http://52.57.29.196:26657 -witness-addr http://public-seed-node.cosmoshub.certus.one:26657
	--height 962118 --hash 28B97BE9F6DE51AC69F70E0B7BFD7E5C9CD1A595B7DC31AFF27C50D4948020CD`,
	}

	cmd.Flags().String(listenAddr, "tcp://localhost:8888", "serve the proxy on the given address")
	cmd.Flags().String(primaryAddr, "", "connect to a Tendermint node at this address")
	cmd.Flags().String(witnessAddrsJoined, "", "tendermint nodes to cross-check the primary node, comma-separated")
	cmd.Flags().String(dir, os.ExpandEnv(filepath.Join("$HOME", ".tendermint-light")), "specify the directory")
	cmd.Flags().Int(maxOpenConnections, 900, "maximum number of simultaneous connections (including WebSocket).")
	cmd.Flags().Duration(trustingPeriod, 168*time.Hour, "trusting period that headers can be verified within. Should be significantly less than the unbonding period")
	cmd.Flags().Int64(trustedHeight, 1, "Trusted header's height")
	cmd.Flags().BytesHex(trustedHash, []byte{}, "Trusted header's hash")
	cmd.Flags().String(logLevel, "info", "Log level, info or debug (Default: info) ")
	cmd.Flags().String(trustLevel, "1/3", "trust level. Must be between 1/3 and 3/3")
	cmd.Flags().Bool(sequential, false, "sequential verification. Verify all headers sequentially as opposed to using skipping verification")

	return cmd
}

func runProxy(cmd *cobra.Command, args []string) error {
	// Initialize logger.
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	var option log.Option
	if logLevel == "info" {
		option, _ = log.AllowLevel("info")
	} else {
		option, _ = log.AllowLevel("debug")
	}
	logger = log.NewFilter(logger, option)

	chainID := args[0]
	logger.Info("Creating client...", "chainID", chainID)

	witnessesAddrs := []string{}
	if witnessAddrsJoined != "" {
		witnessesAddrs = strings.Split(witnessAddrsJoined, ",")
	}

	lightDB, err := dbm.NewGoLevelDB("light-client-db", dir)
	if err != nil {
		return fmt.Errorf("can't create a db: %w", err)
	}
	// create a prefixed db on the chainID
	db := dbm.NewPrefixDB(lightDB, []byte(chainID))

	primaryAddress := ""

	if primaryAddr == "" { // check to see if we can start from an existing state
		var err error
		primaryAddress, witnessesAddrs, err = checkForExistingProviders(db)
		if err != nil {
			return fmt.Errorf("failed to retrieve primary or witness from db: %w", err)
		}
		if primaryAddress == "" {
			return errors.New("no primary address was provided nor found. Please provide a primary (using -p)." +
				" Run the command: tendermint light --help for more information")
		}
	} else {
		err := saveProviders(db, primaryAddr, witnessAddrsJoined)
		if err != nil {
			logger.Error("Unable to save primary and or witness addresses", "err", err)
		}
	}

	tl, err := cmd.Flags().GetString(trustLevel)
	if err != nil {
		return err
	}

	trustLevel, err := tmmath.ParseFraction(tl)
	if err != nil {
		return fmt.Errorf("can't parse trust level: %w", err)
	}

	options := []light.Option{
		light.Logger(logger),
		light.ConfirmationFunction(func(action string) bool {
			fmt.Println(action)
			scanner := bufio.NewScanner(os.Stdin)
			for {
				scanner.Scan()
				response := scanner.Text()
				switch response {
				case "y", "Y":
					return true
				case "n", "N":
					return false
				default:
					fmt.Println("please input 'Y' or 'n' and press ENTER")
				}
			}
		}),
	}

	sequential, err := cmd.Flags().GetBool(sequential)
	if err != nil {
		return err
	}

	if sequential {
		options = append(options, light.SequentialVerification())
	} else {
		options = append(options, light.SkippingVerification(trustLevel))
	}

	var c *light.Client
	trustedHeight, err := cmd.Flags().GetInt64(trustedHeight)
	if err != nil {
		return err
	}

	trustedHash, err := cmd.Flags().GetBytesHex(trustedHash)
	if err != nil {
		return err
	}
	trustingPeriod, err := cmd.Flags().GetDuration(trustingPeriod)
	if err != nil {
		return err
	}

	if trustedHeight > 0 && len(trustedHash) > 0 { // fresh installation
		c, err = light.NewHTTPClient(
			context.Background(),
			chainID,
			light.TrustOptions{
				Period: trustingPeriod,
				Height: trustedHeight,
				Hash:   trustedHash,
			},
			primaryAddr,
			witnessesAddrs,
			dbs.New(lightDB, chainID),
			options...,
		)
	} else { // continue from latest state
		c, err = light.NewHTTPClientFromTrustedStore(
			chainID,
			trustingPeriod,
			primaryAddr,
			witnessesAddrs,
			dbs.New(lightDB, chainID),
			options...,
		)
	}
	if err != nil {
		return err
	}

	config := tmcfg.DefaultConfig()

	cfg := rpcserver.DefaultConfig()
	cfg.MaxBodyBytes = config.RPC.MaxBodyBytes
	cfg.MaxHeaderBytes = config.RPC.MaxHeaderBytes
	maxOpenConnections, err := cmd.Flags().GetInt(maxOpenConnections)
	if err != nil {
		return err
	}

	cfg.MaxOpenConnections = maxOpenConnections
	// If necessary adjust global WriteTimeout to ensure it's greater than
	// TimeoutBroadcastTxCommit.
	// See https://github.com/tendermint/tendermint/issues/3435
	if cfg.WriteTimeout <= config.RPC.TimeoutBroadcastTxCommit {
		cfg.WriteTimeout = config.RPC.TimeoutBroadcastTxCommit + 1*time.Second
	}

	p, err := lproxy.NewProxy(c, listenAddr, primaryAddr, cfg, logger, lrpc.KeyPathFn(MerkleKeyPathFn()))
	if err != nil {
		return err
	}

	// Stop upon receiving SIGTERM or CTRL-C.
	tmos.TrapSignal(logger, func() {
		p.Listener.Close()
	})

	logger.Info("Starting proxy...", "laddr", listenAddr)
	if err := p.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		logger.Error("proxy ListenAndServe", "err", err)
	}

	return nil
}

func checkForExistingProviders(db dbm.DB) (string, []string, error) {
	primaryBytes, err := db.Get(primaryKey)
	if err != nil {
		return "", []string{""}, err
	}
	witnessesBytes, err := db.Get(witnessesKey)
	if err != nil {
		return "", []string{""}, err
	}
	witnessesAddrs := strings.Split(string(witnessesBytes), ",")
	return string(primaryBytes), witnessesAddrs, nil
}

func saveProviders(db dbm.DB, primaryAddr, witnessesAddrs string) error {
	err := db.Set(primaryKey, []byte(primaryAddr))
	if err != nil {
		return fmt.Errorf("failed to save primary provider: %w", err)
	}
	err = db.Set(witnessesKey, []byte(witnessesAddrs))
	if err != nil {
		return fmt.Errorf("failed to save witness providers: %w", err)
	}
	return nil
}

// DefaultMerkleKeyPathFn creates a function used to generate merkle key paths
// from a path string and a key. This is the default used by the cosmos SDK.
// This merkle key paths are required when verifying /abci_query calls
func MerkleKeyPathFn() lrpc.KeyPathFunc {
	// regexp for extracting store name from /abci_query path
	storeNameRegexp := regexp.MustCompile(`\/store\/(.+)\/key`)

	return func(path string, key []byte) (merkle.KeyPath, error) {
		matches := storeNameRegexp.FindStringSubmatch(path)
		if len(matches) != 2 {
			return nil, fmt.Errorf("can't find store name in %s using %s", path, storeNameRegexp)
		}
		storeName := matches[1]

		kp := merkle.KeyPath{}
		kp = kp.AppendKey([]byte(storeName), merkle.KeyEncodingURL)
		kp = kp.AppendKey(key, merkle.KeyEncodingURL)
		return kp, nil
	}
}
