package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/tools/cosmovisor"

	"github.com/cosmos/cosmos-sdk/x/upgrade/plan"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func NewPrepareUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-upgrade",
		Short: "Prepare for the next upgrade",
		Long: `Prepare for the next upgrade by downloading and verifying the upgrade binary.
This command will query the chain for the current upgrade plan and download the specified binary.
gRPC must be enabled on the node for this command to work.`,
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

	grpcAddress := cfg.GRPCAddress
	logger.Info("Using gRPC address", "address", grpcAddress)

	upgradeInfo, err := queryUpgradeInfoFromChain(grpcAddress)
	if err != nil {
		return fmt.Errorf("failed to query upgrade info: %w", err)
	}

	if upgradeInfo == nil {
		logger.Info("No active upgrade plan found")
		return nil
	}

	logger.Info("Preparing for upgrade", "name", upgradeInfo.Name, "height", upgradeInfo.Height)

	upgradeInfoParsed, err := plan.ParseInfo(upgradeInfo.Info, plan.ParseOptionEnforceChecksum(cfg.DownloadMustHaveChecksum))
	if err != nil {
		return fmt.Errorf("failed to parse upgrade info: %w", err)
	}

	binaryURL, err := cosmovisor.GetBinaryURL(upgradeInfoParsed.Binaries)
	if err != nil {
		return fmt.Errorf("binary URL not found in upgrade plan. Cannot prepare for upgrade: %w", err)
	}

	logger.Info("Downloading upgrade binary", "url", binaryURL)

	if err := plan.DownloadUpgrade(cfg.UpgradeDir(upgradeInfo.Name), binaryURL, cfg.Name); err != nil {
		return fmt.Errorf("failed to download and verify binary: %w", err)
	}

	logger.Info("Upgrade preparation complete", "name", upgradeInfo.Name, "height", upgradeInfo.Height)

	return nil
}

func queryUpgradeInfoFromChain(grpcAddress string) (*upgradetypes.Plan, error) {
	if grpcAddress == "" {
		return nil, fmt.Errorf("gRPC address is empty")
	}

	grpcConn, err := getClient(grpcAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to open gRPC client: %w", err)
	}
	defer grpcConn.Close()

	queryClient := upgradetypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := queryClient.CurrentPlan(ctx, &upgradetypes.QueryCurrentPlanRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to query current upgrade plan: %w", err)
	}

	return res.Plan, nil
}

func getClient(endpoint string) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if strings.HasPrefix(endpoint, "https://") {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	return grpc.NewClient(endpoint, opts...)
}
