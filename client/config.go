package client

import (
	"github.com/spf13/cobra"
	"github.com/mitchellh/go-homedir"
	"bufio"
	"path"
	"os"
	"github.com/pkg/errors"
	"encoding/json"
	"io/ioutil"
	"github.com/pelletier/go-toml"
	"fmt"
)

type cliConfig struct {
	Home      string `toml:"home"`
	ChainID   string `toml:"chain_id"`
	TrustNode bool   `toml:"trust_node"`
	Encoding  string `toml:"encoding"`
	Output    string `toml:"output"`
	Node      string `toml:"node"`
	Trace     bool   `toml:"trace"`
}

// ConfigCmd returns a CLI command to interactively create a
// Gaia CLI config file.
func ConfigCmd() *cobra.Command {
	cfg := &cobra.Command{
		Use:   "config",
		Short: "Interactively creates a Gaia CLI config file",
		RunE:  runConfigCmd,
	}

	return cfg
}

func runConfigCmd(cmd *cobra.Command, args [] string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	stdin := BufferStdin()
	gaiaDHome, err := handleGaiaDHome(home, stdin)
	if err != nil {
		return err
	}
	gaiaCLIHome, err := handleGaiaCLIHome(home, stdin)
	if err != nil {
		return err
	}
	node, err := handleNode(stdin)
	if err != nil {
		return err
	}
	trustNode, err := handleTrustNode(stdin)
	if err != nil {
		return err
	}

	encoding := "btc"
	output := "text"
	var chainID string

	gaiaDCfgPath := path.Join(gaiaDHome, "config")
	if info, err := os.Stat(gaiaDCfgPath); err == nil && info.IsDir() {
		chainID, err = processGaiaDConfig(gaiaDCfgPath)
		if err != nil {
			fmt.Println("Couldn't get gaiad file. Using empty chainID.")
		}
	} else {
		fmt.Println("No gaiad config found. Using empty chainID.")
	}

	cfg := &cliConfig{
		Home:      gaiaCLIHome,
		ChainID:   chainID,
		TrustNode: trustNode,
		Encoding:  encoding,
		Output:    output,
		Node:      node,
		Trace:     false,
	}

	return createGaiaCLIConfig(cfg)
}

func handleGaiaDHome(dir string, stdin *bufio.Reader) (string, error) {
	home, err := GetString("Where is your gaiad home directory? (Default: ~/.gaiad)", stdin)
	if err != nil {
		return "", err
	}

	if home == "" {
		home = path.Join(dir, ".gaiad")
	}

	return home, nil
}

func handleGaiaCLIHome(dir string, stdin *bufio.Reader) (string, error) {
	dirName := ".gaiacli"
	home, err := GetString(fmt.Sprintf("Where is your gaiacli home directory? (Default: ~/%s)", dirName), stdin)
	if err != nil {
		return "", err
	}

	if home == "" {
		home = path.Join(dir, dirName)
	}

	return home, nil
}

func handleNode(stdin *bufio.Reader) (string, error) {
	defaultNode := "tcp://localhost:26657"
	node, err := GetString(fmt.Sprintf("Where is your validator node running? (Default: %s)", defaultNode), stdin)
	if err != nil {
		return "", err
	}

	if node == "" {
		node = defaultNode
	}

	return node, nil
}

func handleTrustNode(stdin *bufio.Reader) (bool, error) {
	return GetConfirmation("Do you trust this node?", stdin)
}

func processGaiaDConfig(cfgPath string) (string, error) {
	fp := path.Join(cfgPath, "genesis.json")
	info, err := os.Stat(fp)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return "", errors.New("is directory")
	}

	genesis, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	json.Unmarshal(genesis, &data)
	chainID, ok := data["chain_id"].(string)
	if !ok {
		return "", errors.New("chain_id is not a string")
	}

	return chainID, nil
}

func createGaiaCLIConfig(cfg *cliConfig) error {
	cfgPath := path.Join(cfg.Home, "config")
	err := os.MkdirAll(cfgPath, os.ModePerm)
	if err != nil {
		return err
	}

	data, err := toml.Marshal(*cfg)
	if err != nil {
		return err
	}

	cfgFile := path.Join(cfgPath, "config.toml")
	if info, err := os.Stat(cfgFile); err == nil && !info.IsDir() {
		err = os.Rename(cfgFile, path.Join(cfgPath, "config.toml-old"))
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(cfgFile, data, os.ModePerm)
}
