package commands

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/plugins/ibc"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
	tmtypes "github.com/tendermint/tendermint/types"
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
		Run:   ibcRegisterTxCmd,
	}

	IBCUpdateTxCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the latest state of a blockchain via IBC",
		Run:   ibcUpdateTxCmd,
	}

	IBCPacketTxCmd = &cobra.Command{
		Use:   "packet",
		Short: "Send a new packet via IBC",
	}

	IBCPacketCreateTxCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an egress IBC packet",
		Run:   ibcPacketCreateTxCmd,
	}

	IBCPacketPostTxCmd = &cobra.Command{
		Use:   "post",
		Short: "Deliver an IBC packet to another chain",
		Run:   ibcPacketPostTxCmd,
	}
)

//flags
var (
	ibcChainIDFlag  string
	ibcGenesisFlag  string
	ibcHeaderFlag   string
	ibcCommitFlag   string
	ibcFromFlag     string
	ibcToFlag       string
	ibcTypeFlag     string
	ibcPayloadFlag  string
	ibcPacketFlag   string
	ibcProofFlag    string
	ibcSequenceFlag int
	ibcHeightFlag   int
)

func init() {

	// register flags
	registerFlags := []Flag2Register{
		{&ibcChainIDFlag, "ibc_chain_id", "", "ChainID for the new blockchain"},
		{&ibcGenesisFlag, "genesis", "", "Genesis file for the new blockchain"},
	}

	updateFlags := []Flag2Register{
		{&ibcHeaderFlag, "header", "", "Block header for an ibc update"},
		{&ibcCommitFlag, "commit", "", "Block commit for an ibc update"},
	}

	fromFlagReg := Flag2Register{&ibcFromFlag, "ibc_from", "", "Source ChainID"}

	packetCreateFlags := []Flag2Register{
		fromFlagReg,
		{&ibcToFlag, "to", "", "Destination ChainID"},
		{&ibcTypeFlag, "type", "", "IBC packet type (eg. coin},"},
		{&ibcPayloadFlag, "payload", "", "IBC packet payload"},
		{&ibcSequenceFlag, "ibc_sequence", -1, "sequence number for IBC packet"},
	}

	packetPostFlags := []Flag2Register{
		fromFlagReg,
		{&ibcHeightFlag, "height", 0, "Height the packet became egress in source chain"},
		{&ibcPacketFlag, "packet", "", "hex-encoded IBC packet"},
		{&ibcProofFlag, "proof", "", "hex-encoded proof of IBC packet from source chain"},
	}

	RegisterFlags(IBCRegisterTxCmd, registerFlags)
	RegisterFlags(IBCUpdateTxCmd, updateFlags)
	RegisterFlags(IBCPacketCreateTxCmd, packetCreateFlags)
	RegisterFlags(IBCPacketPostTxCmd, packetPostFlags)

	//register commands
	IBCTxCmd.AddCommand(IBCRegisterTxCmd, IBCUpdateTxCmd, IBCPacketTxCmd)
	IBCPacketTxCmd.AddCommand(IBCPacketCreateTxCmd, IBCPacketPostTxCmd)
	RegisterTxSubcommand(IBCTxCmd)
}

//---------------------------------------------------------------------
// ibc command implementations

func ibcRegisterTxCmd(cmd *cobra.Command, args []string) {
	chainID := ibcChainIDFlag
	genesisFile := ibcGenesisFlag

	genesisBytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error reading genesis file %v: %+v\n", genesisFile, err))
	}

	ibcTx := ibc.IBCRegisterChainTx{
		ibc.BlockchainGenesis{
			ChainID: chainID,
			Genesis: string(genesisBytes),
		},
	}

	fmt.Println("IBCTx:", string(wire.JSONBytes(ibcTx)))

	data := []byte(wire.BinaryBytes(struct {
		ibc.IBCTx `json:"unwrap"`
	}{ibcTx}))
	name := "IBC"

	AppTx(name, data)
}

func ibcUpdateTxCmd(cmd *cobra.Command, args []string) {
	headerBytes, err := hex.DecodeString(StripHex(ibcHeaderFlag))
	if err != nil {
		cmn.Exit(fmt.Sprintf("Header (%v) is invalid hex: %+v\n", ibcHeaderFlag, err))
	}

	commitBytes, err := hex.DecodeString(StripHex(ibcCommitFlag))
	if err != nil {
		cmn.Exit(fmt.Sprintf("Commit (%v) is invalid hex: %+v\n", ibcCommitFlag, err))
	}

	header := new(tmtypes.Header)
	commit := new(tmtypes.Commit)

	err = wire.ReadBinaryBytes(headerBytes, &header)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error unmarshalling header: %+v\n", err))
	}

	err = wire.ReadBinaryBytes(commitBytes, &commit)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error unmarshalling commit: %+v\n", err))
	}

	ibcTx := ibc.IBCUpdateChainTx{
		Header: *header,
		Commit: *commit,
	}

	fmt.Println("IBCTx:", string(wire.JSONBytes(ibcTx)))

	data := []byte(wire.BinaryBytes(struct {
		ibc.IBCTx `json:"unwrap"`
	}{ibcTx}))
	name := "IBC"

	AppTx(name, data)
}

func ibcPacketCreateTxCmd(cmd *cobra.Command, args []string) {
	fromChain, toChain := ibcFromFlag, ibcToFlag
	packetType := ibcTypeFlag

	payloadBytes, err := hex.DecodeString(StripHex(ibcPayloadFlag))
	if err != nil {
		cmn.Exit(fmt.Sprintf("Payload (%v) is invalid hex: %+v\n", ibcPayloadFlag, err))
	}

	sequence, err := ibcSequenceCmd()
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	ibcTx := ibc.IBCPacketCreateTx{
		Packet: ibc.Packet{
			SrcChainID: fromChain,
			DstChainID: toChain,
			Sequence:   sequence,
			Type:       packetType,
			Payload:    payloadBytes,
		},
	}

	fmt.Println("IBCTx:", string(wire.JSONBytes(ibcTx)))

	data := []byte(wire.BinaryBytes(struct {
		ibc.IBCTx `json:"unwrap"`
	}{ibcTx}))

	AppTx("IBC", data)
}

func ibcPacketPostTxCmd(cmd *cobra.Command, args []string) {
	fromChain, fromHeight := ibcFromFlag, ibcHeightFlag

	packetBytes, err := hex.DecodeString(StripHex(ibcPacketFlag))
	if err != nil {
		cmn.Exit(fmt.Sprintf("Packet (%v) is invalid hex: %+v\n", ibcPacketFlag, err))
	}

	proofBytes, err := hex.DecodeString(StripHex(ibcProofFlag))
	if err != nil {
		cmn.Exit(fmt.Sprintf("Proof (%v) is invalid hex: %+v\n", ibcProofFlag, err))
	}

	var packet ibc.Packet
	proof := new(merkle.IAVLProof)

	err = wire.ReadBinaryBytes(packetBytes, &packet)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error unmarshalling packet: %+v\n", err))
	}

	err = wire.ReadBinaryBytes(proofBytes, &proof)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error unmarshalling proof: %+v\n", err))
	}

	ibcTx := ibc.IBCPacketPostTx{
		FromChainID:     fromChain,
		FromChainHeight: uint64(fromHeight),
		Packet:          packet,
		Proof:           proof,
	}

	fmt.Println("IBCTx:", string(wire.JSONBytes(ibcTx)))

	data := []byte(wire.BinaryBytes(struct {
		ibc.IBCTx `json:"unwrap"`
	}{ibcTx}))

	AppTx("IBC", data)
}

func ibcSequenceCmd() (uint64, error) {
	if ibcSequenceFlag >= 0 {
		return uint64(ibcSequenceFlag), nil
	}

	// TODO: get sequence
	return 0, nil
}
