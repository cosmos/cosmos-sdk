package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/plugins/ibc"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
	tmtypes "github.com/tendermint/tendermint/types"
)

func cmdIBCRegisterTx(c *cli.Context) error {
	chainID := c.String("chain_id")
	genesisFile := c.String("genesis")
	parent := c.Parent()

	genesisBytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return errors.New(cmn.Fmt("Error reading genesis file %v: %v", genesisFile, err))
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

	return appTx(parent, name, data)
}

func cmdIBCUpdateTx(c *cli.Context) error {
	parent := c.Parent()

	headerBytes, err := hex.DecodeString(stripHex(c.String("header")))
	if err != nil {
		return errors.New(cmn.Fmt("Header (%v) is invalid hex: %v", c.String("header"), err))
	}
	commitBytes, err := hex.DecodeString(stripHex(c.String("commit")))
	if err != nil {
		return errors.New(cmn.Fmt("Commit (%v) is invalid hex: %v", c.String("commit"), err))
	}

	header := new(tmtypes.Header)
	commit := new(tmtypes.Commit)

	if err := wire.ReadBinaryBytes(headerBytes, &header); err != nil {
		return errors.New(cmn.Fmt("Error unmarshalling header: %v", err))
	}
	if err := wire.ReadBinaryBytes(commitBytes, &commit); err != nil {
		return errors.New(cmn.Fmt("Error unmarshalling commit: %v", err))
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

	return appTx(parent, name, data)
}

func cmdIBCPacketCreateTx(c *cli.Context) error {
	fromChain, toChain := c.String("from"), c.String("to")
	packetType := c.String("type")

	payloadBytes, err := hex.DecodeString(stripHex(c.String("payload")))
	if err != nil {
		return errors.New(cmn.Fmt("Payload (%v) is invalid hex: %v", c.String("payload"), err))
	}

	sequence, err := getIBCSequence(c)
	if err != nil {
		return err
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

	return appTx(c.Parent(), "IBC", data)
}

func cmdIBCPacketPostTx(c *cli.Context) error {
	fromChain, fromHeight := c.String("from"), c.Int("height")

	packetBytes, err := hex.DecodeString(stripHex(c.String("packet")))
	if err != nil {
		return errors.New(cmn.Fmt("Packet (%v) is invalid hex: %v", c.String("packet"), err))
	}
	proofBytes, err := hex.DecodeString(stripHex(c.String("proof")))
	if err != nil {
		return errors.New(cmn.Fmt("Proof (%v) is invalid hex: %v", c.String("proof"), err))
	}

	var packet ibc.Packet
	var proof merkle.IAVLProof

	if err := wire.ReadBinaryBytes(packetBytes, &packet); err != nil {
		return errors.New(cmn.Fmt("Error unmarshalling packet: %v", err))
	}
	if err := wire.ReadBinaryBytes(proofBytes, &proof); err != nil {
		return errors.New(cmn.Fmt("Error unmarshalling proof: %v", err))
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

	return appTx(c.Parent(), "IBC", data)
}

func getIBCSequence(c *cli.Context) (uint64, error) {
	if c.IsSet("sequence") {
		return uint64(c.Int("sequence")), nil
	}

	// TODO: get sequence
	return 0, nil
}
