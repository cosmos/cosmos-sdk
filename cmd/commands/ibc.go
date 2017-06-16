package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/plugins/ibc"

	"github.com/tendermint/go-wire"
)

// returns a new IBC plugin to be registered with Basecoin
func NewIBCPlugin() *ibc.IBCPlugin {
	return ibc.New()
}

//commands
var (
	IBCTxCmd = &cobra.Command{
		Use:   "ibc",
		Short: "An IBC transaction, for InterBlockchain Communication",
	}

	IBCRegisterTxCmd = &cobra.Command{
		Use:   "register",
		Short: "Register a blockchain via IBC",
		RunE:  ibcRegisterTxCmd,
	}
)

//flags
var (
	ibcChainIDFlag string
	ibcGenesisFlag string
)

func init() {
	// register flags
	registerFlags := []Flag2Register{
		{&ibcChainIDFlag, "ibc_chain_id", "", "ChainID for the new blockchain"},
		{&ibcGenesisFlag, "genesis", "", "Genesis file for the new blockchain"},
	}

	RegisterFlags(IBCRegisterTxCmd, registerFlags)

	//register commands
	IBCTxCmd.AddCommand(IBCRegisterTxCmd)
	RegisterTxSubcommand(IBCTxCmd)
}

//---------------------------------------------------------------------
// ibc command implementations

func ibcRegisterTxCmd(cmd *cobra.Command, args []string) error {
	chainID := ibcChainIDFlag
	genesisFile := ibcGenesisFlag

	genesisBytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return errors.Errorf("Error reading genesis file %v: %v\n", genesisFile, err)
	}

	ibcTx := ibc.IBCRegisterChainTx{
		ibc.BlockchainGenesis{
			ChainID: chainID,
			Genesis: string(genesisBytes),
		},
	}

	out, err := json.Marshal(ibcTx)
	if err != nil {
		return err
	}

	fmt.Printf("IBCTx: %s\n", string(out))

	data := []byte(wire.BinaryBytes(struct {
		ibc.IBCTx `json:"unwrap"`
	}{ibcTx}))
	name := "IBC"

	return AppTx(name, data)
}
