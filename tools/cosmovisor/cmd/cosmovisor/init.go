package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"cosmossdk.io/log"
	"cosmossdk.io/tools/cosmovisor"
	"cosmossdk.io/x/upgrade/plan"
)

var initCmd = &cobra.Command{
	Use:          "init <path to executable>",
	Short:        "Initialize a cosmovisor daemon home directory.",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return InitializeCosmovisor(nil, args)
	},
}

// InitializeCosmovisor initializes the cosmovisor directories, current link, and initial executable.
func InitializeCosmovisor(logger log.Logger, args []string) error {
	if len(args) < 1 || len(args[0]) == 0 {
		return errors.New("no <path to executable> provided")
	}
	pathToExe := args[0]
	switch exeInfo, err := os.Stat(pathToExe); {
	case os.IsNotExist(err):
		return fmt.Errorf("executable file not found: %w", err)
	case err != nil:
		return fmt.Errorf("could not stat executable: %w", err)
	case exeInfo.IsDir():
		return errors.New("invalid path to executable: must not be a directory")
	}
	cfg, err := getConfigForInitCmd()
	if err != nil {
		return err
	}

	if logger == nil {
		logger = cfg.Logger(os.Stdout)
	}

	logger.Info("checking on the genesis/bin directory")
	genBinExe := cfg.GenesisBin()
	genBinDir, _ := filepath.Split(genBinExe)
	genBinDir = filepath.Clean(genBinDir)
	switch genBinDirInfo, genBinDirErr := os.Stat(genBinDir); {
	case os.IsNotExist(genBinDirErr):
		logger.Info(fmt.Sprintf("creating directory (and any parents): %q", genBinDir))
		mkdirErr := os.MkdirAll(genBinDir, 0o750)
		if mkdirErr != nil {
			return mkdirErr
		}
	case genBinDirErr != nil:
		return fmt.Errorf("error getting info on genesis/bin directory: %w", genBinDirErr)
	case !genBinDirInfo.IsDir():
		return fmt.Errorf("the path %q already exists but is not a directory", genBinDir)
	default:
		logger.Info(fmt.Sprintf("the %q directory already exists", genBinDir))
	}

	logger.Info("checking on the genesis/bin executable")
	if _, err = os.Stat(genBinExe); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("copying executable into place: %q", genBinExe))
		if cpErr := copyFile(pathToExe, genBinExe); cpErr != nil {
			return cpErr
		}
	} else {
		logger.Info(fmt.Sprintf("the %q file already exists", genBinExe))
	}
	logger.Info(fmt.Sprintf("making sure %q is executable", genBinExe))
	if err = plan.EnsureBinary(genBinExe); err != nil {
		return err
	}

	logger.Info("checking on the current symlink and creating it if needed")
	cur, curErr := cfg.CurrentBin()
	if curErr != nil {
		return curErr
	}
	logger.Info(fmt.Sprintf("the current symlink points to: %q", cur))

	return nil
}

// getConfigForInitCmd gets just the configuration elements needed to initialize cosmovisor.
func getConfigForInitCmd() (*cosmovisor.Config, error) {
	var errs []error
	// Note: Not using GetConfigFromEnv here because that checks that the directories already exist.
	// We also don't care about the rest of the configuration stuff in here.
	cfg := &cosmovisor.Config{
		Home: os.Getenv(cosmovisor.EnvHome),
		Name: os.Getenv(cosmovisor.EnvName),
	}

	var err error
	if cfg.ColorLogs, err = cosmovisor.BooleanOption(cosmovisor.EnvColorLogs, true); err != nil {
		errs = append(errs, err)
	}
	if cfg.TimeFormatLogs, err = cosmovisor.TimeFormatOptionFromEnv(cosmovisor.EnvTimeFormatLogs, time.Kitchen); err != nil {
		errs = append(errs, err)
	}

	if len(cfg.Name) == 0 {
		errs = append(errs, fmt.Errorf("%s is not set", cosmovisor.EnvName))
	}
	switch {
	case len(cfg.Home) == 0:
		errs = append(errs, fmt.Errorf("%s is not set", cosmovisor.EnvHome))
	case !filepath.IsAbs(cfg.Home):
		errs = append(errs, fmt.Errorf("%s must be an absolute path", cosmovisor.EnvHome))
	}
	if len(errs) > 0 {
		return cfg, errors.Join(errs...)
	}
	return cfg, nil
}

// copyFile copies the file at the given source to the given destination.
func copyFile(source, destination string) error {
	// assume we already know that src exists and is a regular file.
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}
