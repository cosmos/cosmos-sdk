package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"cosmossdk.io/log"
	"cosmossdk.io/tools/cosmovisor"
	cverrors "cosmossdk.io/tools/cosmovisor/errors"
	"cosmossdk.io/x/upgrade/plan"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init <path to executable>",
	Short: "Initializes a cosmovisor daemon home directory.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := cmd.Context().Value(log.ContextKey).(*zerolog.Logger)

		return InitializeCosmovisor(logger, args)
	},
}

// InitializeCosmovisor initializes the cosmovisor directories, current link, and initial executable.
func InitializeCosmovisor(logger *zerolog.Logger, args []string) error {
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

	logger.Info().Msg("checking on the genesis/bin directory")
	genBinExe := cfg.GenesisBin()
	genBinDir, _ := filepath.Split(genBinExe)
	genBinDir = filepath.Clean(genBinDir)
	switch genBinDirInfo, genBinDirErr := os.Stat(genBinDir); {
	case os.IsNotExist(genBinDirErr):
		logger.Info().Msgf("creating directory (and any parents): %q", genBinDir)
		mkdirErr := os.MkdirAll(genBinDir, 0o755)
		if mkdirErr != nil {
			return mkdirErr
		}
	case genBinDirErr != nil:
		return fmt.Errorf("error getting info on genesis/bin directory: %w", genBinDirErr)
	case !genBinDirInfo.IsDir():
		return fmt.Errorf("the path %q already exists but is not a directory", genBinDir)
	default:
		logger.Info().Msgf("the %q directory already exists", genBinDir)
	}

	logger.Info().Msg("checking on the genesis/bin executable")
	if _, err = os.Stat(genBinExe); os.IsNotExist(err) {
		logger.Info().Msgf("copying executable into place: %q", genBinExe)
		if cpErr := copyFile(pathToExe, genBinExe); cpErr != nil {
			return cpErr
		}
	} else {
		logger.Info().Msgf("the %q file already exists", genBinExe)
	}
	logger.Info().Msgf("making sure %q is executable", genBinExe)
	if err = plan.EnsureBinary(genBinExe); err != nil {
		return err
	}

	logger.Info().Msg("checking on the current symlink and creating it if needed")
	cur, curErr := cfg.CurrentBin()
	if curErr != nil {
		return curErr
	}
	logger.Info().Msgf("the current symlink points to: %q", cur)

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
		return nil, cverrors.FlattenErrors(errs...)
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
