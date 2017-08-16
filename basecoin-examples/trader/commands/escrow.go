package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tendermint/basecoin-examples/trader/plugins/escrow"
	"github.com/tendermint/basecoin-examples/trader/types"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
	bc "github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"github.com/urfave/cli"
)

const EscrowName = "escrow"

var (
	EscrowRecvFlag = cli.StringFlag{
		Name:  "recv",
		Usage: "Who is the intended recipient of the escrow",
	}
	EscrowArbiterFlag = cli.StringFlag{
		Name:  "arbiter",
		Usage: "Who is the arbiter of the escrow",
	}
	EscrowAddrFlag = cli.StringFlag{
		Name:  "escrow",
		Usage: "The address of this escrow",
	}
	EscrowExpireFlag = cli.Uint64Flag{
		Name:  "expire",
		Value: 0,
		Usage: "The block height when the escrow expires",
	}
	EscrowPayoutFlag = cli.BoolTFlag{
		Name:  "abort-payout",
		Usage: "Set this flag if to return the money to the sender",
	}
)

var (
	EscrowTxCmd = cli.Command{
		Name:  "escrow",
		Usage: "Create and resolve escrows",
		Subcommands: []cli.Command{
			EscrowCreateTxCmd,
			EscrowResolveTxCmd,
			EscrowExpireTxCmd,
			EscrowQueryCmd,
		},
	}

	EscrowCreateTxCmd = cli.Command{
		Name:  "create",
		Usage: "Create a new escrow by sending money",
		Flags: append(bcmd.TxFlags,
			EscrowRecvFlag,
			EscrowArbiterFlag,
			EscrowExpireFlag,
		),
		Action: func(c *cli.Context) error {
			return cmdEscrowCreateTx(c)
		},
	}

	EscrowResolveTxCmd = cli.Command{
		Name:  "pay",
		Usage: "Resolve the escrow by paying out of returning the money",
		Flags: append(bcmd.TxFlags,
			EscrowAddrFlag,
			EscrowPayoutFlag,
		),
		Action: func(c *cli.Context) error {
			return cmdEscrowResolveTx(c)
		},
	}

	EscrowExpireTxCmd = cli.Command{
		Name:  "expire",
		Usage: "Call to expire the escrow if no action in a given time",
		Flags: append(bcmd.TxFlags,
			EscrowAddrFlag,
		),
		Action: func(c *cli.Context) error {
			return cmdEscrowExpireTx(c)
		},
	}

	EscrowQueryCmd = cli.Command{
		Name:      "query",
		Usage:     "Return the contents of the given escrow",
		ArgsUsage: "<address>",
		Action: func(c *cli.Context) error {
			return cmdEscrowQuery(c)
		},
		Flags: []cli.Flag{
			bcmd.NodeFlag,
		},
	}
)

func init() {
	bcmd.RegisterTxSubcommand(EscrowTxCmd)
	bcmd.RegisterStartPlugin(EscrowName,
		func() bc.Plugin { return escrow.New(EscrowName) })
}

func cmdEscrowCreateTx(c *cli.Context) error {
	recvHex := c.String(EscrowRecvFlag.Name)
	arbHex := c.String(EscrowArbiterFlag.Name)
	expire := c.Uint64(EscrowExpireFlag.Name)

	// convert destination address to bytes
	recv, err := hex.DecodeString(bcmd.StripHex(recvHex))
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	// convert destination address to bytes
	arb, err := hex.DecodeString(bcmd.StripHex(arbHex))
	if err != nil {
		return errors.New("Arbiter address is invalid hex: " + err.Error())
	}

	tx := types.CreateEscrowTx{
		Recipient:  recv,
		Arbiter:    arb,
		Expiration: expire,
	}
	data := types.EscrowTxBytes(tx)
	return bcmd.AppTx(c, EscrowName, data)
}

func cmdEscrowResolveTx(c *cli.Context) error {
	addrHex := c.String(EscrowAddrFlag.Name)
	payout := c.Bool(EscrowPayoutFlag.Name)

	// convert destination address to bytes
	addr, err := hex.DecodeString(bcmd.StripHex(addrHex))
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	tx := types.ResolveEscrowTx{
		Escrow: addr,
		Payout: payout,
	}
	data := types.EscrowTxBytes(tx)
	return bcmd.AppTx(c, EscrowName, data)
}

func cmdEscrowExpireTx(c *cli.Context) error {
	addrHex := c.String(EscrowAddrFlag.Name)

	// convert destination address to bytes
	addr, err := hex.DecodeString(bcmd.StripHex(addrHex))
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	tx := types.ExpireEscrowTx{
		Escrow: addr,
	}
	data := types.EscrowTxBytes(tx)
	return bcmd.AppTx(c, EscrowName, data)
}

func cmdEscrowQuery(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("account command requires an argument ([address])")
	}
	addrHex := bcmd.StripHex(c.Args()[0])

	// convert destination address to bytes
	addr, err := hex.DecodeString(addrHex)
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	esc, err := getEscrow(c.String("node"), addr)
	if err != nil {
		return err
	}
	fmt.Println(string(wire.JSONBytes(esc)))
	return nil
}

func getEscrow(tmAddr string, address []byte) (*types.EscrowData, error) {
	prefix := []byte(fmt.Sprintf("%s/", EscrowName))
	key := append(prefix, address...)
	response, err := bcmd.Query(tmAddr, key)
	if err != nil {
		return nil, err
	}

	escrowBytes := response.Value

	if len(escrowBytes) == 0 {
		return nil, fmt.Errorf("Escrow bytes are empty for address: %X ", address)
	}
	esc, err := types.ParseEscrow(escrowBytes)
	if err != nil {
		return nil, fmt.Errorf("Error reading account %X error: %v",
			escrowBytes, err.Error())
	}
	return &esc, nil

}
