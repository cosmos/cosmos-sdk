package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/spf13/cobra"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"cosmossdk.io/tools/cosmovisor/v2/internal/watchers"

	"github.com/cosmos/cosmos-sdk/server"
)

func main() {
	cmd := &cobra.Command{
		Use:   "mock_node",
		Short: "A mock node for testing cosmovisor.",
		Long: `The --halt-interval flag is required and must be specified in order to halt the node.
The --upgrade-plan and --halt-height flags are mutually exclusive. It is an error to specify both.
Based on which flag is specified the node will either exhibit --halt-height before or
x/upgrade upgrade-info.json behavior.`,
	}
	var blockTime time.Duration
	var upgradePlan string
	var haltHeight uint64
	var homePath string
	var httpAddr string
	var blockUrl string
	var shutdownDelay time.Duration
	var shutdownOnUpgrade bool
	var upgradeInfoEncodingJson bool
	cmd.Flags().DurationVar(&blockTime, "block-time", 0, "Duration of time between blocks. This is required to simulate a progression of blocks over time.")
	cmd.Flags().StringVar(&upgradePlan, "upgrade-plan", "", "upgrade-info.json to create after the halt duration is reached. Either this flag or --halt-height must be specified but not both.")
	cmd.Flags().Uint64Var(&haltHeight, server.FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node. E")
	cmd.Flags().StringVar(&homePath, "home", "", "Home directory for the mock node. upgrade-info.json will be written to the data sub-directory of this directory. Defaults to the current directory.")
	cmd.Flags().StringVar(&httpAddr, "http-addr", ":26657", "HTTP server address to serve block information. Defaults to :26657.")
	cmd.Flags().StringVar(&blockUrl, "block-url", "/block", "URL at which the latest block information is served. Defaults to /block.")
	cmd.Flags().DurationVar(&shutdownDelay, "shutdown-delay", 0, "Duration to wait before shutting down the node upon receiving a shutdown signal. Defaults to 0 (no delay).")
	cmd.Flags().BoolVar(&shutdownOnUpgrade, "shutdown-on-upgrade", false, "If true, the node will shutdown immediately after reaching the upgrade height. If false, it will continue running until a shutdown signal is received. Defaults to false.")
	cmd.Flags().BoolVar(&upgradeInfoEncodingJson, "upgrade-info-encoding-json", false, "If true, the upgrade-info.json will be encoded using encoding/json instead of jsonpb. This is useful for testing compatibility with different JSON decoders. Defaults to false (uses jsonpb).")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if upgradePlan == "" && haltHeight == 0 {
			return fmt.Errorf("must specify either --upgrade-plan or --halt-height")
		}
		if blockTime == 0 {
			return fmt.Errorf("must specify --block-time")
		}
		if homePath == "" {
			var err error
			homePath, err = os.Getwd() // Default to current working directory if not specified
			if err != nil {
				return fmt.Errorf("unable to determine current working directory: %w", err)
			}
		}
		node := &MockNode{
			height:                  0,
			blockTime:               blockTime,
			haltHeight:              haltHeight,
			homePath:                homePath,
			httpAddr:                httpAddr,
			blockUrl:                blockUrl,
			shutdownDelay:           shutdownDelay,
			shutdownOnUpgrade:       shutdownOnUpgrade,
			upgradeInfoEncodingJson: upgradeInfoEncodingJson,
			logger:                  log.NewLogger(os.Stdout),
		}
		if upgradePlan != "" {
			node.upgradePlan = &upgradetypes.Plan{}
			err := jsonpb.Unmarshal(bytes.NewBufferString(upgradePlan), node.upgradePlan)
			if err != nil {
				return fmt.Errorf("unable to parse upgrade plan: %w", err)
			}
			if err := node.upgradePlan.ValidateBasic(); err != nil {
				return fmt.Errorf("invalid upgrade plan: %w", err)
			}
		}
		return node.Run(cmd.Context())
	}
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

type MockNode struct {
	height                  uint64
	blockTime               time.Duration
	upgradePlan             *upgradetypes.Plan
	haltHeight              uint64
	homePath                string
	httpAddr                string
	blockUrl                string
	logger                  log.Logger
	shutdownDelay           time.Duration
	shutdownOnUpgrade       bool
	upgradeInfoEncodingJson bool
}

func (n *MockNode) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	upgradeHeight := n.haltHeight
	if n.upgradePlan != nil {
		upgradePlanHeight := uint64(n.upgradePlan.Height)
		if upgradeHeight == 0 || upgradePlanHeight < upgradeHeight {
			upgradeHeight = upgradePlanHeight
		}
	}

	actualHeightFile := path.Join(n.homePath, "data", "actual-height")
	// try to read the actual-height file if it exists
	if bz, err := os.ReadFile(actualHeightFile); err == nil {
		n.logger.Info("Reading existing height", "height", string(bz))
		n.height, err = strconv.ParseUint(string(bz), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse actual height from file: %w", err)
		}
	}

	n.logger.Info("Starting mock node", "start_height", n.height, "block_time", n.blockTime, "upgrade_plan", n.upgradePlan, "halt_height", n.haltHeight)
	srv := n.startHTTPServer()
	ticker := time.NewTicker(n.blockTime)
	defer ticker.Stop()
	for n.height < upgradeHeight {
		n.logger.Info("Processed mock block", "height", n.height)
		select {
		case <-ctx.Done():
			n.logger.Info("Received shutdown signal, stopping node")
			if err := srv.Shutdown(ctx); err != nil {
				n.logger.Error("Error shutting down HTTP server", "err", err)
			}
			if n.shutdownDelay > 0 {
				n.logger.Info("Waiting for shutdown delay", "delay", n.shutdownDelay)
				time.Sleep(n.shutdownDelay)
			}
			return nil
		case <-ticker.C:
			n.height++
			// Write the current height to the actual-height file
			err := os.WriteFile(actualHeightFile, []byte(fmt.Sprintf("%d", n.height)), 0o644)
			if err != nil {
				return fmt.Errorf("failed to write actual height to file: %w", err)
			}
		}
	}
	if n.haltHeight == upgradeHeight { // if we have a halt height and we've reached it - there could be an earlier gov upgrade
		// this log line matches what BaseApp does when it reaches the halt height
		n.logger.Error(fmt.Sprintf("halt per configuration height %d", n.height))
	} else if n.upgradePlan != nil {
		n.logger.Info("Mock node reached upgrade height, writing upgrade-info.json", "upgrade_plan", n.upgradePlan)
		upgradeInfoPath := path.Join(n.homePath, "data", upgradetypes.UpgradeInfoFilename)
		var out string
		var err error
		if n.upgradeInfoEncodingJson {
			var bz []byte
			bz, err = json.Marshal(n.upgradePlan)
			out = string(bz)
		} else {
			out, err = (&jsonpb.Marshaler{
				EmitDefaults: false,
			}).MarshalToString(n.upgradePlan)
		}
		if err != nil {
			return fmt.Errorf("failed to marshal upgrade plan: %w", err)
		}
		err = os.MkdirAll(path.Dir(upgradeInfoPath), 0o755)
		if err != nil {
			return fmt.Errorf("failed to create directory for upgrade-info.json: %w", err)
		}
		err = os.WriteFile(upgradeInfoPath, []byte(out), 0o644)
		if err != nil {
			return fmt.Errorf("failed to write upgrade-info.json: %w", err)
		}
	}
	if n.shutdownOnUpgrade {
		n.logger.Info("Mock node reached upgrade height, configured to shut down immediately")
		return nil
	}
	// Don't exit until we receive a shutdown signal
	n.logger.Info("Mock node reached upgrade height, waiting for shutdown signal")
	<-ctx.Done()
	return nil
}

func (n *MockNode) startHTTPServer() *http.Server {
	http.HandleFunc(n.blockUrl, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(watchers.Response{
			Result: watchers.Result{
				Block: watchers.Block{
					Header: watchers.Header{
						Height: fmt.Sprintf("%d", n.height),
					},
				},
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	srv := &http.Server{
		Addr: n.httpAddr,
	}
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			n.logger.Error("HTTP server error", "err", err)
		}
	}()
	return srv
}
