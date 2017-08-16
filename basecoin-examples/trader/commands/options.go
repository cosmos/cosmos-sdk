package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tendermint/basecoin-examples/trader/plugins/options"
	"github.com/tendermint/basecoin-examples/trader/types"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
	bc "github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"github.com/urfave/cli"
)

const OptionName = "options"

var (
	OptionAddrFlag = cli.StringFlag{
		Name:  "option",
		Usage: "The address of this option",
	}
	OptionExpireFlag = cli.Uint64Flag{
		Name:  "expire",
		Value: 0,
		Usage: "The block height when the option expires",
	}
	OptionSellToFlag = cli.StringFlag{
		Name:  "sellto",
		Usage: "Who to sell the options to (optional)",
	}
	OptionTradeAmountFlag = cli.StringFlag{
		Name:  "trade",
		Usage: "Amount of coins to trade in format <amt><coin>,<amt2><coin2>,...",
	}
	OptionPriceAmountFlag = cli.StringFlag{
		Name:  "price",
		Usage: "Price to buy option in format <amt><coin>,<amt2><coin2>,...",
	}
)

var (
	OptionsTxCmd = cli.Command{
		Name:  "options",
		Usage: "Create, trade, and exercise currency options",
		Subcommands: []cli.Command{
			OptionsCreateTxCmd,
			OptionsSellTxCmd,
			OptionsBuyTxCmd,
			OptionsExerciseTxCmd,
			OptionsDisolveTxCmd,
			OptionsQueryCmd,
		},
	}

	OptionsCreateTxCmd = cli.Command{
		Name:  "create",
		Usage: "Create a new option by sending money",
		Flags: append(bcmd.TxFlags,
			OptionExpireFlag,
			OptionTradeAmountFlag,
		),
		Action: func(c *cli.Context) error {
			return cmdOptionCreateTx(c)
		},
	}

	OptionsSellTxCmd = cli.Command{
		Name:  "sell",
		Usage: "Offer to sell this option",
		Flags: append(bcmd.TxFlags,
			OptionAddrFlag,
			OptionSellToFlag,
			OptionPriceAmountFlag,
		),
		Action: func(c *cli.Context) error {
			return cmdOptionSellTx(c)
		},
	}

	OptionsBuyTxCmd = cli.Command{
		Name:  "buy",
		Usage: "Attempt to buy this option",
		Flags: append(bcmd.TxFlags,
			OptionAddrFlag,
		),
		Action: func(c *cli.Context) error {
			return cmdOptionBuyTx(c)
		},
	}

	OptionsExerciseTxCmd = cli.Command{
		Name:  "exercise",
		Usage: "Exercise this option to trade currency at the given rate",
		Flags: append(bcmd.TxFlags,
			OptionAddrFlag,
		),
		Action: func(c *cli.Context) error {
			return cmdOptionExerciseTx(c)
		},
	}

	OptionsDisolveTxCmd = cli.Command{
		Name:  "disolve",
		Usage: "Attempt to disolve this option (if never sold, or already expired)",
		Flags: append(bcmd.TxFlags,
			OptionAddrFlag,
		),
		Action: func(c *cli.Context) error {
			return cmdOptionDisolveTx(c)
		},
	}

	OptionsQueryCmd = cli.Command{
		Name:      "query",
		Usage:     "Return the contents of the given option",
		ArgsUsage: "<address>",
		Action: func(c *cli.Context) error {
			return cmdOptionQuery(c)
		},
		Flags: []cli.Flag{
			bcmd.NodeFlag,
		},
	}
)

func init() {
	bcmd.RegisterTxSubcommand(OptionsTxCmd)
	bcmd.RegisterStartPlugin(OptionName,
		func() bc.Plugin { return options.New(OptionName) })
}

func cmdOptionCreateTx(c *cli.Context) error {
	tradeAmount := c.String(OptionTradeAmountFlag.Name)
	expire := c.Uint64(EscrowExpireFlag.Name)

	tradeCoins, err := bcmd.ParseCoins(tradeAmount)
	if err != nil {
		return err
	}

	tx := types.CreateOptionTx{
		Expiration: expire,
		Trade:      tradeCoins,
	}
	data := types.OptionsTxBytes(tx)
	return bcmd.AppTx(c, OptionName, data)
}

func cmdOptionSellTx(c *cli.Context) error {
	addrHex := c.String(OptionAddrFlag.Name)
	buyerHex := c.String(OptionSellToFlag.Name)
	price := c.String(OptionPriceAmountFlag.Name)

	// convert destination address to bytes
	addr, err := hex.DecodeString(bcmd.StripHex(addrHex))
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	buyer, err := hex.DecodeString(bcmd.StripHex(buyerHex))
	if err != nil { // this is optional, we can ignore it
		buyer = nil
	}

	priceCoins, err := bcmd.ParseCoins(price)
	if err != nil {
		return err
	}

	tx := types.SellOptionTx{
		Addr:      addr,
		NewHolder: buyer,
		Price:     priceCoins,
	}
	data := types.OptionsTxBytes(tx)
	return bcmd.AppTx(c, OptionName, data)
}

func cmdOptionBuyTx(c *cli.Context) error {
	addrHex := c.String(OptionAddrFlag.Name)

	// convert destination address to bytes
	addr, err := hex.DecodeString(bcmd.StripHex(addrHex))
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	tx := types.BuyOptionTx{
		Addr: addr,
	}
	data := types.OptionsTxBytes(tx)
	return bcmd.AppTx(c, OptionName, data)
}

func cmdOptionExerciseTx(c *cli.Context) error {
	addrHex := c.String(OptionAddrFlag.Name)

	// convert destination address to bytes
	addr, err := hex.DecodeString(bcmd.StripHex(addrHex))
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	tx := types.ExerciseOptionTx{
		Addr: addr,
	}
	data := types.OptionsTxBytes(tx)
	return bcmd.AppTx(c, OptionName, data)
}

func cmdOptionDisolveTx(c *cli.Context) error {
	addrHex := c.String(OptionAddrFlag.Name)

	// convert destination address to bytes
	addr, err := hex.DecodeString(bcmd.StripHex(addrHex))
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	tx := types.DisolveOptionTx{
		Addr: addr,
	}
	data := types.OptionsTxBytes(tx)
	return bcmd.AppTx(c, OptionName, data)
}

func cmdOptionQuery(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("account command requires an argument ([address])")
	}
	addrHex := bcmd.StripHex(c.Args()[0])

	// convert destination address to bytes
	addr, err := hex.DecodeString(addrHex)
	if err != nil {
		return errors.New("Recv address is invalid hex: " + err.Error())
	}

	opt, err := getOption(c.String("node"), addr)
	if err != nil {
		return err
	}
	fmt.Println(string(wire.JSONBytes(opt)))
	return nil
}

func getOption(tmAddr string, address []byte) (*types.OptionData, error) {
	prefix := []byte(fmt.Sprintf("%s/", OptionName))
	key := append(prefix, address...)
	response, err := bcmd.Query(tmAddr, key)
	if err != nil {
		return nil, err
	}

	optionBytes := response.Value

	if len(optionBytes) == 0 {
		return nil, fmt.Errorf("Option bytes are empty for address: %X ", address)
	}
	opt, err := types.ParseOptionData(optionBytes)
	if err != nil {
		return nil, fmt.Errorf("Error reading option %X error: %v",
			optionBytes, err.Error())
	}
	return &opt, nil
}
