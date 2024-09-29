package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
	"cosmossdk.io/x/upgrade/plan"
)

func NewPrepareUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-upgrade",
		Short: "Prepare for the next upgrade",
		Long: `Prepare for the next upgrade by downloading and verifying the upgrade binary.
This command will download the binary specified in the upgrade-info.json file,
verify its checksum, and place it in the appropriate directory for Cosmovisor to use.`,
		RunE:         prepareUpgradeHandler,
		SilenceUsage: false,
		Args:         cobra.NoArgs,
	}
	return cmd
}

func prepareUpgradeHandler(cmd *cobra.Command, _ []string) error {
	configPath, err := cmd.Flags().GetString(cosmovisor.FlagCosmovisorConfig)
	if err != nil {
		return fmt.Errorf("failed to get config flag: %w", err)
	}

	cfg, err := cosmovisor.GetConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	logger := cfg.Logger(cmd.OutOrStdout())
	upgradeInfo, err := cfg.UpgradeInfo()
	if err != nil {
		return fmt.Errorf("failed to get upgrade info: %w", err)
	}

	logger.Info("Preparing for upgrade", "name", upgradeInfo.Name, "height", upgradeInfo.Height)

	upgradeInfoParsed, err := plan.ParseInfo(upgradeInfo.Info, plan.ParseOptionEnforceChecksum(cfg.DownloadMustHaveChecksum))
	if err != nil {
		return fmt.Errorf("failed to parse upgrade info: %w", err)
	}

	binaryURL, err := cosmovisor.GetBinaryURL(upgradeInfoParsed.Binaries)
	if err != nil {
		return fmt.Errorf("failed to get binary URL: %w", err)
	}

	logger.Info("Downloading upgrade binary", "url", binaryURL)

	upgradeBin := filepath.Join(cfg.UpgradeBin(upgradeInfo.Name), cfg.Name)
	if err := plan.DownloadUpgrade(filepath.Dir(upgradeBin), binaryURL, cfg.Name); err != nil {
		return fmt.Errorf("failed to download and verify binary: %w", err)
	}

	logger.Info("Upgrade preparation complete", "name", upgradeInfo.Name, "height", upgradeInfo.Height)

	return nil
}
