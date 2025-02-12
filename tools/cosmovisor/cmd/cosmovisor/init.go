package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"cosmossdk.io/log"
	"cosmossdk.io/tools/cosmovisor"
	"cosmossdk.io/x/upgrade/plan"
)

func NewInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init <path to executable>",
		Short: "Initialize a cosmovisor daemon home directory.",
		Long: `Initialize a cosmovisor daemon home directory with the provided executable.
Configuration file is initialized at the default path (<-home->/cosmovisor/config.toml).`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitializeCosmovisor(nil, args)
		},
	}

	return initCmd
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

	// skipping validation to not check if directories exist
	cfg, err := cosmovisor.GetConfigFromEnv(true)
	if err != nil {
		return err
	}

	// process to minimal validation
	if err := minConfigValidate(cfg); err != nil {
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

	// set current working directory to $DAEMON_NAME/cosmosvisor
	// to allow current symlink to be relative
	if err = os.Chdir(cfg.Root()); err != nil {
		return fmt.Errorf("failed to change directory to %s: %w", cfg.Root(), err)
	}

	logger.Info("checking on the current symlink and creating it if needed")
	cur, curErr := cfg.CurrentBin()
	if curErr != nil {
		return curErr
	}
	logger.Info(fmt.Sprintf("the current symlink points to: %q", cur))

	filePath, err := cfg.Export()
	if err != nil {
		return fmt.Errorf("failed to export configuration: %w", err)
	}
	logger.Info(fmt.Sprintf("cosmovisor config.toml created at: %s", filePath))

	return nil
}

func minConfigValidate(cfg *cosmovisor.Config) error {
	var errs []error
	if len(cfg.Name) == 0 {
		errs = append(errs, fmt.Errorf("%s is not set", cosmovisor.EnvName))
	}

	switch {
	case len(cfg.Home) == 0:
		errs = append(errs, fmt.Errorf("%s is not set", cosmovisor.EnvHome))
	case !filepath.IsAbs(cfg.Home):
		errs = append(errs, fmt.Errorf("%s must be an absolute path", cosmovisor.EnvHome))
	}

	return errors.Join(errs...)
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
