package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
	"cosmossdk.io/x/upgrade/plan"
)

func NewPrepareUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-upgrade",
		Short: "Prepare for the next upgrade",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cosmovisor.GetConfigFromFile(cmd.Flag(cosmovisor.FlagCosmovisorConfig).Value.String())
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}

			logger := cfg.Logger(os.Stdout)

			upgradeInfo, err := cfg.UpgradeInfo()
			if err != nil {
				return fmt.Errorf("failed to get upgrade info: %w", err)
			}

			if upgradeInfo.Name == "" {
				logger.Info("No upgrade scheduled")
				return nil
			}

			logger.Info("Preparing for upgrade", "name", upgradeInfo.Name, "height", upgradeInfo.Height)

			upgradeInfoParsed, err := plan.ParseInfo(upgradeInfo.Info, plan.ParseOptionEnforceChecksum(cfg.DownloadMustHaveChecksum))
			if err != nil {
				return fmt.Errorf("failed to parse upgrade info: %w", err)
			}

			binaryURL, err := getBinaryURL(upgradeInfoParsed.Binaries)
			if err != nil {
				return fmt.Errorf("failed to get binary URL: %w", err)
			}

			binaryPath := filepath.Join(cfg.UpgradeDir(upgradeInfo.Name), "bin", cfg.Name)
			if err := downloadBinary(binaryURL, binaryPath); err != nil {
				return fmt.Errorf("failed to download binary: %w", err)
			}

			logger.Info("Binary downloaded", "path", binaryPath)

			if err := verifyChecksum(binaryPath, upgradeInfoParsed.Binaries[runtime.GOOS+"/"+runtime.GOARCH]); err != nil {
				return fmt.Errorf("checksum verification failed: %w", err)
			}

			logger.Info("Checksum verified successfully")

			return nil
		},
	}

	return cmd
}

func getBinaryURL(binaries plan.BinaryDownloadURLMap) (string, error) {
	osArch := runtime.GOOS + "/" + runtime.GOARCH
	url, ok := binaries[osArch]
	if !ok {
		return "", fmt.Errorf("binary not found for %s", osArch)
	}
	
	parts := strings.Split(url, "?checksum=")
	if len(parts) > 1 {
		return parts[0], nil
	}
	return url, nil
}

func downloadBinary(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	
	parts := strings.Split(expectedChecksum, "sha256:")
	if len(parts) != 2 {
		return fmt.Errorf("invalid checksum format: %s", expectedChecksum)
	}
	expectedChecksumOnly := strings.TrimSpace(parts[1])

	if actualChecksum != expectedChecksumOnly {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksumOnly, actualChecksum)
	}

	return nil
}
