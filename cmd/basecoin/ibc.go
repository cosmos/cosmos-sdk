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

	data := wire.BinaryBytes(ibcTx)
	name := "ibc"

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

	var header tmtypes.Header
	var commit tmtypes.Commit

	if err := wire.ReadBinaryBytes(headerBytes, &header); err != nil {
		return errors.New(cmn.Fmt("Error unmarshalling header: %v", err))
	}
	if err := wire.ReadBinaryBytes(commitBytes, &commit); err != nil {
		return errors.New(cmn.Fmt("Error unmarshalling commit: %v", err))
	}

	ibcTx := ibc.IBCUpdateChainTx{
		Header: header,
		Commit: commit,
	}

	fmt.Println("IBCTx:", string(wire.JSONBytes(ibcTx)))

	data := wire.BinaryBytes(ibcTx)
	name := "ibc"

	return appTx(parent, name, data)
}

func cmdIBCPacketCreateTx(c *cli.Context) error {
	return nil
}

func cmdIBCPacketPostTx(c *cli.Context) error {
	parent := c.Parent()

	var fromChain string
	var fromHeight uint64
	var proof merkle.IAVLProof

	var srcChain, dstChain string
	var sequence uint64
	var packetType string
	var payload []byte

	ibcTx := ibc.IBCPacketTx{
		FromChainID:     fromChain,
		FromChainHeight: fromHeight,
		Packet: ibc.Packet{
			SrcChainID: srcChain,
			DstChainID: dstChain,
			Sequence:   sequence,
			Type:       packetType,
			Payload:    payload,
		},
		Proof: proof,
	}

	fmt.Println("IBCTx:", string(wire.JSONBytes(ibcTx)))

	data := wire.BinaryBytes(ibcTx)
	name := "ibc"

	return appTx(parent, name, data)
}
