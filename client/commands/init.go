package commands

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/certifiers"
	"github.com/tendermint/light-client/certifiers/files"
	"github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/tendermint/types"
)

var (
	dirPerm = os.FileMode(0700)
)

//nolint
const (
	SeedFlag      = "seed"
	HashFlag      = "valhash"
	GenesisFlag   = "genesis"
	FlagTrustNode = "trust-node"

	ConfigFile = "config.toml"
)

// InitCmd will initialize the basecli store
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the light client for a new chain",
	RunE:  runInit,
}

var ResetCmd = &cobra.Command{
	Use:   "reset_all",
	Short: "DANGEROUS: Wipe out all client data, including keys",
	RunE:  runResetAll,
}

func init() {
	InitCmd.Flags().Bool("force-reset", false, "Wipe clean an existing client store, except for keys")
	InitCmd.Flags().String(SeedFlag, "", "Seed file to import (optional)")
	InitCmd.Flags().String(HashFlag, "", "Trusted validator hash (must match to accept)")
	InitCmd.Flags().String(GenesisFlag, "", "Genesis file with chainid and validators (optional)")
}

func runInit(cmd *cobra.Command, args []string) error {
	root := viper.GetString(cli.HomeFlag)
	if viper.GetBool("force-reset") {
		resetRoot(root, true)
	}

	// make sure we don't have an existing client initialized
	inited, err := WasInited(root)
	if err != nil {
		return err
	}
	if inited {
		return errors.Errorf("%s already is initialized, --force-reset if you really want to wipe it out", root)
	}

	// clean up dir if init fails
	err = doInit(cmd, root)
	if err != nil {
		resetRoot(root, true)
	}
	return err
}

// doInit actually creates all the files, on error, we should revert it all
func doInit(cmd *cobra.Command, root string) error {
	// read the genesis file if present, and populate --chain-id and --valhash
	err := checkGenesis(cmd)
	if err != nil {
		return err
	}

	err = initConfigFile(cmd)
	if err != nil {
		return err
	}
	err = initSeed()
	return err
}

func runResetAll(cmd *cobra.Command, args []string) error {
	root := viper.GetString(cli.HomeFlag)
	resetRoot(root, false)
	return nil
}

func resetRoot(root string, saveKeys bool) {
	tmp := filepath.Join(os.TempDir(), cmn.RandStr(16))
	keys := filepath.Join(root, "keys")
	if saveKeys {
		os.Rename(keys, tmp)
	}
	os.RemoveAll(root)
	if saveKeys {
		os.Mkdir(root, 0700)
		os.Rename(tmp, keys)
	}
}

type Runable func(cmd *cobra.Command, args []string) error

// Any commands that require and init'ed basecoin directory
// should wrap their RunE command with RequireInit
// to make sure that the client is initialized.
//
// This cannot be called during PersistentPreRun,
// as they are called from the most specific command first, and root last,
// and the root command sets up viper, which is needed to find the home dir.
func RequireInit(run Runable) Runable {
	return func(cmd *cobra.Command, args []string) error {
		// otherwise, run the wrappped command
		if viper.GetBool(FlagTrustNode) {
			return run(cmd, args)
		}

		// first check if we were Init'ed and if not, return an error
		root := viper.GetString(cli.HomeFlag)
		init, err := WasInited(root)
		if err != nil {
			return err
		}
		if !init {
			return errors.Errorf("You must run '%s init' first", cmd.Root().Name())
		}

		// otherwise, run the wrappped command
		return run(cmd, args)
	}
}

// WasInited returns true if a basecoin was previously initialized
// in this directory.  Important to ensure proper behavior.
//
// Returns error if we have filesystem errors
func WasInited(root string) (bool, error) {
	// make sure there is a directory here in any case
	os.MkdirAll(root, dirPerm)

	// check if there is a config.toml file
	cfgFile := filepath.Join(root, "config.toml")
	_, err := os.Stat(cfgFile)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.WithStack(err)
	}

	// check if there are non-empty checkpoints and validators dirs
	dirs := []string{
		filepath.Join(root, files.CheckDir),
		filepath.Join(root, files.ValDir),
	}
	// if any of these dirs is empty, then we have no data
	for _, d := range dirs {
		empty, err := isEmpty(d)
		if err != nil {
			return false, err
		}
		if empty {
			return false, nil
		}
	}

	// looks like we have everything
	return true, nil
}

func checkGenesis(cmd *cobra.Command) error {
	genesis := viper.GetString(GenesisFlag)
	if genesis == "" {
		return nil
	}

	doc, err := types.GenesisDocFromFile(genesis)
	if err != nil {
		return err
	}

	flags := cmd.Flags()
	flags.Set(ChainFlag, doc.ChainID)
	hash := doc.ValidatorHash()
	hexHash := hex.EncodeToString(hash)
	flags.Set(HashFlag, hexHash)

	return nil
}

// isEmpty returns false if we can read files in this dir.
// if it doesn't exist, read issues, etc... return true
//
// TODO: should we handle errors otherwise?
func isEmpty(dir string) (bool, error) {
	// check if we can read the directory, missing is fine, other error is not
	d, err := os.Open(dir)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer d.Close()

	// read to see if any (at least one) files here...
	files, err := d.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	if err != nil {
		return false, errors.WithStack(err)
	}
	empty := len(files) == 0
	return empty, nil
}

type Config struct {
	Chain    string `toml:"chain-id,omitempty"`
	Node     string `toml:"node,omitempty"`
	Output   string `toml:"output,omitempty"`
	Encoding string `toml:"encoding,omitempty"`
}

func setConfig(flags *pflag.FlagSet, f string, v *string) {
	if flags.Changed(f) {
		*v = viper.GetString(f)
	}
}

func initConfigFile(cmd *cobra.Command) error {
	flags := cmd.Flags()
	var cfg Config

	required := []string{ChainFlag, NodeFlag}
	for _, f := range required {
		if !flags.Changed(f) {
			return errors.Errorf(`"--%s" required`, f)
		}
	}

	setConfig(flags, ChainFlag, &cfg.Chain)
	setConfig(flags, NodeFlag, &cfg.Node)
	setConfig(flags, cli.OutputFlag, &cfg.Output)
	setConfig(flags, cli.EncodingFlag, &cfg.Encoding)

	out, err := os.Create(filepath.Join(viper.GetString(cli.HomeFlag), ConfigFile))
	if err != nil {
		return errors.WithStack(err)
	}
	defer out.Close()

	// save the config file
	err = toml.NewEncoder(out).Encode(cfg)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func initSeed() (err error) {
	// create a provider....
	trust, source := GetProviders()

	// load a seed file, or get data from the provider
	var seed certifiers.Seed
	seedFile := viper.GetString(SeedFlag)
	if seedFile == "" {
		fmt.Println("Loading validator set from tendermint rpc...")
		seed, err = certifiers.LatestSeed(source)
	} else {
		fmt.Printf("Loading validators from file %s\n", seedFile)
		seed, err = certifiers.LoadSeed(seedFile)
	}
	// can't load the seed? abort!
	if err != nil {
		return err
	}

	// make sure it is a proper seed
	err = seed.ValidateBasic(viper.GetString(ChainFlag))
	if err != nil {
		return err
	}

	// validate hash interactively or not
	hash := viper.GetString(HashFlag)
	if hash != "" {
		var hashb []byte
		hashb, err = hex.DecodeString(hash)
		if err == nil && !bytes.Equal(hashb, seed.Hash()) {
			err = errors.Errorf("Seed hash doesn't match expectation: %X", seed.Hash())
		}
	} else {
		err = validateHash(seed)
	}

	if err != nil {
		return err
	}

	// if accepted, store seed as current state
	trust.StoreSeed(seed)
	return nil
}

func validateHash(seed certifiers.Seed) error {
	// ask the user to verify the validator hash
	fmt.Println("\nImportant: if this is incorrect, all interaction with the chain will be insecure!")
	fmt.Printf("  Given validator hash valid: %X\n", seed.Hash())
	fmt.Println("Is this valid (y/n)?")
	valid := askForConfirmation()
	if !valid {
		return errors.New("Invalid validator hash, try init with proper seed later")
	}
	return nil
}

func askForConfirmation() bool {
	var resp string
	_, err := fmt.Scanln(&resp)
	if err != nil {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
	resp = strings.ToLower(resp)
	if resp == "y" || resp == "yes" {
		return true
	} else if resp == "n" || resp == "no" {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}
