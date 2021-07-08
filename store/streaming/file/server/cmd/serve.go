package cmd

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store/streaming/file"
	"github.com/cosmos/cosmos-sdk/store/streaming/file/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/store/streaming/file/server/grpc"
)

const (
	cliFlagRemoveAfter    = "remove-after"
	cliFlagFilePrefix     = "file-prefix"
	cliFlagReadDir        = "read-dir"
	cliFlagChainID        = "chain-id"
	cliFlagGRPCAddress    = "grpc-address"
	cliFlagGRPCWebEnable  = "grpc-web-enable"
	cliFlagGRPCWebAddress = "grpc-web-address"

	tomlFlagRemoveAfter    = "file-server.remove-after"
	tomlFlagFilePrefix     = "file-server.file-prefix"
	tomlFlagReadDir        = "file-server.read-dir"
	tomlFlagChainID        = "file-server.chain-id"
	tomlFlagGRPCAddress    = "file-server.grpc-address"
	tomlFlagGRPCWebEnable  = "file-server.grpc-web-enable"
	tomlFlagGRPCWebAddress = "file-server.grpc-web-address"
)

var (
	subLogger zerolog.Logger
)

// grpcCmd represents the serve command
var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "serve state change data over gRPC",
	Long: `This command configures a gRPC server to serve data
from state change files.

`,
	Run: func(cmd *cobra.Command, args []string) {
		subLogger = log.With().Str("subcommand", cmd.CalledAs()).Logger()
		serve()
	},
}

func serve() {
	subLogger.Print("starting file server")
	handler, err := servergrpc.NewHandler(serverCfg, marshaller, server.ZeroLogWrapper{Logger: subLogger})
	if err != nil {
		subLogger.Fatal().AnErr("Failed to create a new file server gRPC handler", err)
	}
	grpcSrv, err := servergrpc.StartGRPCServer(handler, serverCfg.GRPCAddress)
	if err != nil {
		subLogger.Fatal().AnErr("Failed to start gRPC server", err)
	}
	var grpcWebSrv *http.Server
	if serverCfg.GRPCWebEnabled {
		grpcWebSrv, err = servergrpc.StartGRPCWeb(grpcSrv, serverCfg.GRPCWebAddress)
		if err != nil {
			subLogger.Fatal().AnErr("Failed to start gRPC web server", err)
		}
	}

	defer func() {
		handler.Stop()
		grpcSrv.Stop()
		if grpcWebSrv != nil {
			grpcWebSrv.Close()
		}
		subLogger.Info().Msg("exiting...")
	}()

	// wait for quit signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	subLogger.Info().Msgf("received %d signal, quiting file-server", int(sig.(syscall.Signal))+128)
}

func init() {
	rootCmd.AddCommand(grpcCmd)

	// cli flags for all config variables
	grpcCmd.PersistentFlags().Bool(cliFlagRemoveAfter, false, "Remove files after reading from them")
	grpcCmd.PersistentFlags().String(cliFlagFilePrefix, "", "Prefix for the files we are reading from")
	grpcCmd.PersistentFlags().String(cliFlagReadDir, file.DefaultWriteDir, "Directory we are reading from")
	grpcCmd.PersistentFlags().String(cliFlagChainID, "", "ChainID for the data we are reading and serving")

	grpcCmd.PersistentFlags().String(cliFlagGRPCAddress, config.DefaultGRPCAddress, "the gRPC server address to listen on")
	grpcCmd.PersistentFlags().Bool(cliFlagGRPCWebEnable, true, "Define if the gRPC-Web server should be enabled.")
	grpcCmd.PersistentFlags().String(cliFlagGRPCWebAddress, config.DefaultGRPCWebAddress, "The gRPC-Web server address to listen on")

	// and their toml bindings
	viper.BindPFlag(tomlFlagRemoveAfter, grpcCmd.PersistentFlags().Lookup(cliFlagRemoveAfter))
	viper.BindPFlag(tomlFlagFilePrefix, grpcCmd.PersistentFlags().Lookup(cliFlagFilePrefix))
	viper.BindPFlag(tomlFlagReadDir, grpcCmd.PersistentFlags().Lookup(cliFlagReadDir))
	viper.BindPFlag(tomlFlagChainID, grpcCmd.PersistentFlags().Lookup(cliFlagChainID))

	viper.BindPFlag(tomlFlagGRPCAddress, grpcCmd.PersistentFlags().Lookup(cliFlagGRPCAddress))
	viper.BindPFlag(tomlFlagGRPCWebEnable, grpcCmd.PersistentFlags().Lookup(cliFlagGRPCWebEnable))
	viper.BindPFlag(tomlFlagGRPCWebAddress, grpcCmd.PersistentFlags().Lookup(cliFlagGRPCWebAddress))
}
