package cosmovisor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cosmossdk.io/log"
	"github.com/otiai10/copy"

	"cosmossdk.io/tools/cosmovisor/internal/checkers"
	"cosmossdk.io/tools/cosmovisor/internal/watchers"
	"github.com/cosmos/cosmos-sdk/x/upgrade/plan"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type Launcher struct {
	logger log.Logger
	cfg    *Config
	ctx    context.Context
	cancel context.CancelFunc
	// upgradePlanWatcher watches for data in an upgrade-info.json created by the running node
	upgradePlanWatcher watchers.Watcher[upgradetypes.Plan]
	// manualUpgradesWatcher watchers for data in an upgrade-info.json.batch created by the node operator
	manualUpgradesWatcher watchers.Watcher[ManualUpgradeBatch]
	haltHeightWatcher     watchers.Watcher[uint64]
	actualHeightWatcher   watchers.Watcher[uint64]
	heightChecker         checkers.HeightChecker
	upgradePlan           *upgradetypes.Plan
	manualUpgrade         *ManualUpgradePlan
}

func NewLauncher(logger log.Logger, cfg *Config) (Launcher, error) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM)
	dirWatcher, err := watchers.NewFSNotifyWatcher(ctx, cfg.UpgradeInfoDir(), []string{
		cfg.UpgradeInfoFilePath(),
		cfg.UpgradeInfoBatchFilePath(),
	})
	if err != nil {
		logger.Warn("failed to intialize fsnotify, it's probably not available on this platform, using polling only", "error", err)
	}

	// TODO the watchers should do data validation in additional to json unmarshaling
	nodeUpgradeWatcher := initWatcher[upgradetypes.Plan](ctx, cfg, dirWatcher, cfg.UpgradeInfoFilePath(), cfg.ParseUpgradeInfo)
	manualUpgradesWatcher := initWatcher[ManualUpgradeBatch](ctx, cfg, dirWatcher, cfg.UpgradeInfoBatchFilePath(), cfg.ParseManualUpgrades)

	return Launcher{
		logger:                logger,
		cfg:                   cfg,
		ctx:                   ctx,
		cancel:                cancel,
		upgradePlanWatcher:    nodeUpgradeWatcher,
		manualUpgradesWatcher: manualUpgradesWatcher,
	}, nil
}

func (l *Launcher) Watch() {
	errChan := joinChannels(l.upgradePlanWatcher.Errors(),
		l.manualUpgradesWatcher.Errors(),
		l.haltHeightWatcher.Errors(),
		l.actualHeightWatcher.Errors())
	for {
		select {
		case <-l.ctx.Done():

			// TODO handle cosmovisor shutdown
			return
		case upgradePlan := <-l.upgradePlanWatcher.Updated():
			l.upgradePlan = &upgradePlan
			// TODO upgrade plan received, positive signal to perform upgrade, no additional checks needed
		case <-l.manualUpgradesWatcher.Updated():
			l.logger.Info("manual upgrades watcher updated")
			// TODO received new manual upgrades batch:
			// must establish current node height and select the first manual upgrade after current height, if any
			// if one is found, node must be restarted with --halt-height
		case <-l.haltHeightWatcher.Updated():
			// TODO check against manual upgrade height
		case <-l.actualHeightWatcher.Updated():
			// TODO check against manual upgrade height
		case err := <-errChan:
			// for now just log errors
			l.logger.Error("error in upgrade plan watcher", "error", err)
		}
	}
}

// TODO fix this with WaitGroup
func joinChannels[T any](ch ...<-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, c := range ch {
			for msg := range c {
				out <- msg
			}
		}
	}()
	return out
}

//// BatchUpgradeWatcher starts a watcher loop that swaps upgrade manifests at the correct
//// height, given the batch upgrade file. It watches the current state of the chain
//// via the websocket API.
//func BatchUpgradeWatcher(ctx context.Context, cfg *Config, logger log.Logger) {
//	// load batch file in memory
//	uInfos, err := ReadManualUpgrades(cfg)
//	if err != nil {
//		logger.Warn("failed to load batch upgrade file", "error", err)
//		uInfos = []upgradetypes.Plan{}
//	}
//
//	watcher, err := fsnotify.NewWatcher()
//	if err != nil {
//		logger.Warn("failed to init watcher", "error", err)
//		return
//	}
//	defer watcher.Close()
//	err = watcher.Add(filepath.Dir(cfg.UpgradeInfoBatchFilePath()))
//	if err != nil {
//		logger.Warn("watcher failed to add upgrade directory", "error", err)
//		return
//	}
//
//	var conn *grpc.ClientConn
//	var grpcErr error
//
//	defer func() {
//		if conn != nil {
//			if err := conn.Close(); err != nil {
//				logger.Warn("couldn't stop gRPC client", "error", err)
//			}
//		}
//	}()
//
//	// Wait for the chain process to be ready
//pollLoop:
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		default:
//			conn, grpcErr = grpc.NewClient(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
//			if grpcErr == nil {
//				break pollLoop
//			}
//			time.Sleep(time.Second)
//		}
//	}
//
//	client := cmtservice.NewServiceClient(conn)
//
//	var prevUpgradeHeight int64 = -1
//
//	logger.Info("starting the batch watcher loop")
//	for {
//		select {
//		case event := <-watcher.Events:
//			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
//				uInfos, err = loadBatchUpgradeFile(cfg)
//				if err != nil {
//					logger.Warn("failed to load batch upgrade file", "error", err)
//					continue
//				}
//			}
//		case <-ctx.Done():
//			return
//		default:
//			if len(uInfos) == 0 {
//				// prevent spending extra CPU cycles
//				time.Sleep(time.Second)
//				continue
//			}
//			resp, err := client.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
//			if err != nil {
//				logger.Warn("error getting latest block", "error", err)
//				time.Sleep(time.Second)
//				continue
//			}
//
//			h := resp.SdkBlock.Header.Height
//			upcomingUpgrade := uInfos[0].Height
//			// replace upgrade-info and upgrade-info batch file
//			if h > prevUpgradeHeight && h < upcomingUpgrade {
//				jsonBytes, err := json.Marshal(uInfos[0])
//				if err != nil {
//					logger.Warn("error marshaling JSON for upgrade-info.json", "error", err, "upgrade", uInfos[0])
//					continue
//				}
//				if err := os.WriteFile(cfg.UpgradeInfoFilePath(), jsonBytes, 0o600); err != nil {
//					logger.Warn("error writing upgrade-info.json", "error", err)
//					continue
//				}
//				uInfos = uInfos[1:]
//
//				jsonBytes, err = json.Marshal(uInfos)
//				if err != nil {
//					logger.Warn("error marshaling JSON for upgrade-info.json.batch", "error", err, "upgrades", uInfos)
//					continue
//				}
//				if err := os.WriteFile(cfg.UpgradeInfoBatchFilePath(), jsonBytes, 0o600); err != nil {
//					logger.Warn("error writing upgrade-info.json.batch", "error", err)
//					// remove the upgrade-info.json.batch file to avoid non-deterministic behavior
//					err := os.Remove(cfg.UpgradeInfoBatchFilePath())
//					if err != nil && !os.IsNotExist(err) {
//						logger.Warn("error removing upgrade-info.json.batch", "error", err)
//						return
//					}
//					continue
//				}
//				prevUpgradeHeight = upcomingUpgrade
//			}
//
//			// Add a small delay to avoid hammering the gRPC endpoint
//			time.Sleep(time.Second)
//		}
//	}
//}

// Run launches the app in a subprocess and returns when the subprocess (app)
// exits (either when it dies, or *after* a successful upgrade.) and upgrade finished.
// Returns true if the upgrade request was detected and the upgrade process started.
func (l Launcher) Run(args []string, stdin io.Reader, stdout, stderr io.Writer) (bool, error) {
	bin, err := l.cfg.CurrentBin()
	if err != nil {
		return false, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := plan.EnsureBinary(bin); err != nil {
		return false, fmt.Errorf("current binary is invalid: %w", err)
	}

	l.logger.Info("running app", "path", bin, "args", args)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("launching process %s %s failed: %w", bin, strings.Join(args, " "), err)
	}

	//var wg sync.WaitGroup
	//wg.Add(1)
	//// TODO: replace BatchUpgradeWatcher
	//go func() {
	//	defer wg.Done()
	//	BatchUpgradeWatcher(ctx, l.cfg, l.logger)
	//}()

	//sigs := make(chan os.Signal, 1)
	//signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGTERM)
	//go func() {
	//	sig := <-sigs
	//	cancel()
	//	wg.Wait()
	//	if err := cmd.Process.Signal(sig); err != nil {
	//		l.logger.Error("terminated", "error", err, "bin", bin)
	//		os.Exit(1)
	//	}
	//}()

	if needsUpdate, err := l.WaitForUpgradeOrExit(cmd); err != nil || !needsUpdate {
		return false, err
	}

	if !IsSkipUpgradeHeight(args, *l.upgradePlan) {
		l.cfg.WaitRestartDelay()

		if err := l.doBackup(); err != nil {
			return false, err
		}

		if err := l.doCustomPreUpgrade(); err != nil {
			return false, err
		}

		if err := UpgradeBinary(l.logger, l.cfg, *l.upgradePlan); err != nil {
			return false, err
		}

		if err = l.doPreUpgrade(); err != nil {
			return false, err
		}

		return true, nil
	}

	//cancel()
	//wg.Wait()

	return false, nil
}

// WaitForUpgradeOrExit checks upgrade plan file created by the app.
// When it returns, the process (app) is finished.
//
// It returns (true, nil) if an upgrade should be initiated (and we killed the process)
// It returns (false, err) if the process died by itself
// It returns (false, nil) if the process exited normally without triggering an upgrade. This is very unlikely
// to happen with "start" but may happen with short-lived commands like `simd genesis export ...`
func (l Launcher) WaitForUpgradeOrExit(cmd *exec.Cmd) (bool, error) {
	//// TODO we shouldn't be getting any current upgrade because we're only using upgrade-info.json to receive signals from the node
	//currentUpgrade, err := l.cfg.UpgradeInfo()
	//if err != nil {
	//	// upgrade info not found do nothing
	//	currentUpgrade = upgradetypes.Plan{}
	//}
	//
	cmdDone := make(chan error)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	// TODO add manual-upgrades watcher
	// TODO replace with upgrade-info.json watcher
	case upgradePlan := <-l.upgradePlanWatcher.Updated():
		l.upgradePlan = &upgradePlan
		// upgrade - kill the process and restart
		l.logger.Info("daemon shutting down in an attempt to restart")

		if l.cfg.ShutdownGrace > 0 {
			// Interrupt signal
			l.logger.Info("sent interrupt to app, waiting for exit")
			_ = cmd.Process.Signal(syscall.SIGTERM)

			// Wait app exit
			psChan := make(chan *os.ProcessState)
			go func() {
				pstate, _ := cmd.Process.Wait()
				psChan <- pstate
			}()

			// Timeout and kill
			select {
			case <-psChan:
				// Normal Exit
				l.logger.Info("app exited normally")
			case <-time.After(l.cfg.ShutdownGrace):
				l.logger.Info("DAEMON_SHUTDOWN_GRACE exceeded, killing app")
				// Kill after grace period
				_ = cmd.Process.Kill()
			}
		} else {
			// Default: Immediate app kill
			_ = cmd.Process.Kill()
		}
	case err := <-cmdDone:
		// no error -> command exits normally (eg. short command like `gaiad version`)
		if err == nil {
			return false, nil
		}
	}
	return true, nil
}

func (l Launcher) doBackup() error {
	// take backup if `UNSAFE_SKIP_BACKUP` is not set.
	if !l.cfg.UnsafeSkipBackup {
		// check if upgrade-info.json is not empty.
		var uInfo upgradetypes.Plan
		upgradeInfoFile, err := os.ReadFile(l.cfg.UpgradeInfoFilePath())
		if err != nil {
			return fmt.Errorf("error while reading upgrade-info.json: %w", err)
		}

		if err = json.Unmarshal(upgradeInfoFile, &uInfo); err != nil {
			return err
		}

		if uInfo.Name == "" {
			return errors.New("upgrade-info.json is empty")
		}

		// a destination directory, Format YYYY-MM-DD
		st := time.Now()
		ymd := fmt.Sprintf("%d-%d-%d", st.Year(), st.Month(), st.Day())
		dst := filepath.Join(l.cfg.DataBackupPath, fmt.Sprintf("data"+"-backup-%s", ymd))

		l.logger.Info("starting to take backup of data directory", "backup start time", st)

		// copy the $DAEMON_HOME/data to a backup dir
		if err = copy.Copy(filepath.Join(l.cfg.Home, "data"), dst); err != nil {
			return fmt.Errorf("error while taking data backup: %w", err)
		}

		// backup is done, lets check endtime to calculate total time taken for backup process
		et := time.Now()
		l.logger.Info("backup completed", "backup saved at", dst, "backup completion time", et, "time taken to complete backup", et.Sub(st))
	}

	return nil
}
