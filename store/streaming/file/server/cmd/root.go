package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/streaming/file/server/config"
)

const (
	cliFlagConfigFile = "config-file"
	cliFlagLogFile    = "log-file"
	cliFlagLogLevel   = "log-level"

	tomlFlagLogFile  = "file-server.log-file"
	tomlFlagLogLevel = "file-server.log-level"
)

var (
	configPath        string
	serverCfg         *config.StateServerConfig
	interfaceRegistry = codecTypes.NewInterfaceRegistry()
	marshaller        = codec.NewProtoCodec(interfaceRegistry)
)

var rootCmd = &cobra.Command{
	Use:              "file-server",
	PersistentPreRun: initFuncs,
}

func Execute() {
	log.Print("----- Starting Cosmos state change file server -----")
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().StringVar(&configPath, cliFlagConfigFile, "", "path to file server config file")

	// cli flags for all config variables
	rootCmd.PersistentFlags().String(cliFlagLogFile, "", "File for writing log messages to")
	rootCmd.PersistentFlags().String(cliFlagLogLevel, "", "Log level: log messages above this level will be logged")

	// and their toml bindings
	viper.BindPFlag(tomlFlagLogFile, rootCmd.PersistentFlags().Lookup(cliFlagLogFile))
	viper.BindPFlag(tomlFlagLogLevel, rootCmd.PersistentFlags().Lookup(cliFlagLogLevel))
}

func initFuncs(cmd *cobra.Command, args []string) {
	logfilePath := viper.GetString(tomlFlagLogFile)
	if logfilePath != "" {
		file, err := os.OpenFile(logfilePath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.Info().Str("logfile", logfilePath)
			log.Output(file)
		} else {
			log.Output(os.Stdout)
			log.Print("Failed to log to file, using default stdout")
		}
	} else {
		log.Output(os.Stdout)
	}
	logLevel()
}

func initConfig() {
	serverCfg = config.DefaultStateServerConfig()
	if configPath != "" {
		switch _, err := os.Stat(configPath); {
		case os.IsNotExist(err):
			log.Warn().Msgf("No config file found at %s, using default configuration", configPath)
			config.WriteConfigFile(configPath, serverCfg)

		case err != nil:
			log.Fatal().AnErr("Config file is specified but could not be loaded", err)

		default:
			_, fileName := filepath.Split(configPath)
			viper.SetConfigType("toml")
			viper.SetConfigName(fileName)
			viper.SetConfigFile(configPath)

			if err := viper.ReadInConfig(); err != nil {
				log.Fatal().AnErr(fmt.Sprintf("failed to read in config file '%s'", fileName), err)
			}
		}
	} else {
		log.Warn().Msg("No config file passed with --config flag, using default configuration")
		config.WriteConfigFile(configPath, serverCfg)
	}
	if err := viper.Unmarshal(serverCfg); err != nil {
		log.Fatal().Err(err)
	}
}

func logLevel() {
	lvl := parseLevel(viper.GetString(tomlFlagLogLevel))
	zerolog.SetGlobalLevel(lvl)
	log.Printf("Log level set to ", lvl.String())
}

func parseLevel(lvlString string) zerolog.Level {
	switch strings.ToLower(lvlString) {
	case "panic", "p", "5":
		return zerolog.PanicLevel
	case "fatal", "f", "4":
		return zerolog.FatalLevel
	case "error", "err", "e", "3":
		return zerolog.ErrorLevel
	case "warn", "w", "2":
		return zerolog.WarnLevel
	case "info", "i", "1":
		return zerolog.InfoLevel
	case "debug", "d", "0":
		return zerolog.DebugLevel
	case "trace", "t", "-1":
		return zerolog.TraceLevel
	default:
		return zerolog.Disabled
	}
}
