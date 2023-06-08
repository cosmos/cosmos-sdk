package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/cometbft/cometbft/libs/cli"
)

const (
	EnvVariable = "ENVIRONMENT"
	EnvMainnet  = "mainnet"
	EnvLocalnet = "localnet"
)

// ExportAirdropSnapshotCmd generates a snapshot.json from a provided exported genesis.json.
func ChangeEnvironmentCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-env [new env]",
		Short: "Set home environment variables for commands",
		Long: `Set home environment variables for commands
Example:
	osmosisd set-env mainnet
	osmosisd set-env localnet
	osmosisd set-env $HOME/.custom-dir
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			newEnv := args[0]

			currentEnvironment := GetHomeEnvironment(defaultNodeHome)
			fmt.Println("Current environment: ", currentEnvironment)

			if _, err := EnvironmentNameToPath(newEnv, defaultNodeHome); err != nil {
				return err
			}

			fmt.Println("New environment: ", newEnv)

			envMap := make(map[string]string)
			envMap[EnvVariable] = newEnv
			err := godotenv.Write(envMap, filepath.Join(defaultNodeHome, ".env"))
			if err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

// PrintEnvironmentCmd prints the current environment.
func PrintEnvironmentCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-env",
		Short: "Prints the current environment",
		Long: `Prints the current environment
Example:
	osmosisd get-env'
	Returns one of:
	- mainnet implying $HOME/.osmosisd
	- localosmosis implying $HOME/.osmosisd-local
	- localosmosis
	- custom path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			environment := GetHomeEnvironment(defaultNodeHome)
			path, err := EnvironmentNameToPath(environment, defaultNodeHome)
			if err != nil {
				return err
			}

			fmt.Println("Environment name: ", environment)
			fmt.Println("Environment path: ", path)
			return nil
		},
	}
	return cmd
}

func CreateEnvFile(cmd *cobra.Command, defaultNodeHome string) error {
	// Check if .env file was created in /.osmosisd
	envPath := filepath.Join(defaultNodeHome, ".env")
	if _, err := os.Stat(envPath); err != nil {
		// If not exist, we create a new .env file with node dir passed
		if os.IsNotExist(err) {
			// Create ./osmosisd if not exist
			if _, err = os.Stat(defaultNodeHome); err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(defaultNodeHome, 0777)
					if err != nil {
						return err
					}
				}
			}

			// Create environment file
			envFile, err := os.Create(envPath)
			if err != nil {
				return err
			}

			// In case the user wants to init in a specific dir, save it to .env
			nodeHome, err := cmd.Flags().GetString(cli.HomeFlag)
			if err != nil {
				fmt.Println("using mainnet environment")
				nodeHome = EnvMainnet
			}
			_, err = envFile.WriteString(fmt.Sprintf("OSMOSISD_ENVIRONMENT=%s", nodeHome))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func EnvironmentNameToPath(environmentName string, defaultNodeHome string) (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch environmentName {
	case EnvMainnet:
		return defaultNodeHome, nil
	case EnvLocalnet:
		return filepath.Join(userHomeDir, defaultNodeHome+"-local/"), nil
	default:
		osmosisdPath := filepath.Join(userHomeDir, environmentName)
		_, err := os.Stat(osmosisdPath)
		if os.IsNotExist(err) {
			// Creating new environment directory
			if err := os.Mkdir(osmosisdPath, os.ModePerm); err != nil {
				return "", err
			}
		}
		return osmosisdPath, nil
	}
}

func GetHomeEnvironment(defaultNodeHome string) string {
	envPath := filepath.Join(defaultNodeHome, ".env")

	// Use default node home if can't get environment
	err := godotenv.Load(envPath)
	if err != nil {
		// Failed to load, using default home directory
		return EnvMainnet
	}
	val := os.Getenv(EnvVariable)
	return val
}