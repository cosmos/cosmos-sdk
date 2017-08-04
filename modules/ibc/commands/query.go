package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/query"
	"github.com/tendermint/basecoin/modules/ibc"
	"github.com/tendermint/basecoin/stack"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
)

// TODO: query seeds (register/update)

// IBCQueryCmd - parent command to query ibc info
var IBCQueryCmd = &cobra.Command{
	Use:   "ibc",
	Short: "Get information about IBC",
	RunE:  commands.RequireInit(ibcQueryCmd),
	// HandlerInfo
}

// ChainsQueryCmd - get a list of all registered chains
var ChainsQueryCmd = &cobra.Command{
	Use:   "chains",
	Short: "Get a list of all registered chains",
	RunE:  commands.RequireInit(chainsQueryCmd),
	// ChainSet ([]string)
}

// ChainQueryCmd - get details on one registered chain
var ChainQueryCmd = &cobra.Command{
	Use:   "chain [id]",
	Short: "Get details on one registered chain",
	RunE:  commands.RequireInit(chainQueryCmd),
	// ChainInfo
}

// PacketsQueryCmd - get latest packet in a queue
var PacketsQueryCmd = &cobra.Command{
	Use:   "packets",
	Short: "Get latest packet in a queue",
	RunE:  commands.RequireInit(packetsQueryCmd),
	// uint64
}

// PacketQueryCmd - get the names packet (by queue and sequence)
var PacketQueryCmd = &cobra.Command{
	Use:   "packet",
	Short: "Get packet with given sequence from the named queue",
	RunE:  commands.RequireInit(packetQueryCmd),
	// Packet
}

//nolint
const (
	FlagFromChain = "from"
	FlagToChain   = "to"
	FlagSequence  = "sequence"
)

func init() {
	IBCQueryCmd.AddCommand(
		ChainQueryCmd,
		ChainsQueryCmd,
		PacketQueryCmd,
		PacketsQueryCmd,
	)

	fs1 := PacketsQueryCmd.Flags()
	fs1.String(FlagFromChain, "", "Name of the input chain (where packets came from)")
	fs1.String(FlagToChain, "", "Name of the output chain (where packets go to)")

	fs2 := PacketQueryCmd.Flags()
	fs2.String(FlagFromChain, "", "Name of the input chain (where packets came from)")
	fs2.String(FlagToChain, "", "Name of the output chain (where packets go to)")
	fs2.Int(FlagSequence, -1, "Index of the packet in the queue (starts with 0)")
}

func ibcQueryCmd(cmd *cobra.Command, args []string) error {
	var res ibc.HandlerInfo
	key := stack.PrefixedKey(ibc.NameIBC, ibc.HandlerKey())
	prove := !viper.GetBool(commands.FlagTrustNode)
	h, err := query.GetParsed(key, &res, prove)
	if err != nil {
		return err
	}
	return query.OutputProof(res, h)
}

func chainsQueryCmd(cmd *cobra.Command, args []string) error {
	list := [][]byte{}
	key := stack.PrefixedKey(ibc.NameIBC, ibc.ChainsKey())
	prove := !viper.GetBool(commands.FlagTrustNode)
	h, err := query.GetParsed(key, &list, prove)
	if err != nil {
		return err
	}

	// convert these names to strings for better output
	res := make([]string, len(list))
	for i := range list {
		res[i] = string(list[i])
	}

	return query.OutputProof(res, h)
}

func chainQueryCmd(cmd *cobra.Command, args []string) error {
	arg, err := commands.GetOneArg(args, "id")
	if err != nil {
		return err
	}

	var res ibc.ChainInfo
	key := stack.PrefixedKey(ibc.NameIBC, ibc.ChainKey(arg))
	prove := !viper.GetBool(commands.FlagTrustNode)
	h, err := query.GetParsed(key, &res, prove)
	if err != nil {
		return err
	}

	return query.OutputProof(res, h)
}

func assertOne(from, to string) error {
	if from == "" && to == "" {
		return errors.Errorf("You must specify either --%s or --%s",
			FlagFromChain, FlagToChain)
	}
	if from != "" && to != "" {
		return errors.Errorf("You can only specify one of --%s or --%s",
			FlagFromChain, FlagToChain)
	}
	return nil
}

func packetsQueryCmd(cmd *cobra.Command, args []string) error {
	from := viper.GetString(FlagFromChain)
	to := viper.GetString(FlagToChain)
	err := assertOne(from, to)
	if err != nil {
		return err
	}

	var key []byte
	if from != "" {
		key = stack.PrefixedKey(ibc.NameIBC, ibc.QueueInKey(from))
	} else {
		key = stack.PrefixedKey(ibc.NameIBC, ibc.QueueOutKey(to))
	}

	var res uint64
	prove := !viper.GetBool(commands.FlagTrustNode)
	h, err := query.GetParsed(key, &res, prove)
	if err != nil {
		return err
	}

	return query.OutputProof(res, h)
}

func packetQueryCmd(cmd *cobra.Command, args []string) error {
	from := viper.GetString(FlagFromChain)
	to := viper.GetString(FlagToChain)
	err := assertOne(from, to)
	if err != nil {
		return err
	}

	seq := viper.GetInt(FlagSequence)
	if seq < 0 {
		return errors.Errorf("--%s must be a non-negative number", FlagSequence)
	}
	prove := !viper.GetBool(commands.FlagTrustNode)

	var key []byte
	if from != "" {
		key = stack.PrefixedKey(ibc.NameIBC, ibc.QueueInPacketKey(from, uint64(seq)))
	} else {
		key = stack.PrefixedKey(ibc.NameIBC, ibc.QueueOutPacketKey(to, uint64(seq)))
	}

	// Input queue just display the results
	var packet ibc.Packet
	if from != "" {
		h, err := query.GetParsed(key, &packet, prove)
		if err != nil {
			return err
		}
		return query.OutputProof(packet, h)
	}

	// output queue, create a post packet
	bs, height, proof, err := query.GetWithProof(key)
	if err != nil {
		return err
	}
	err = wire.ReadBinaryBytes(bs, &packet)
	if err != nil {
		return err
	}

	// create the post packet here.
	post := ibc.PostPacketTx{
		FromChainID:     commands.GetChainID(),
		FromChainHeight: height,
		Key:             key,
		Packet:          packet,
		Proof:           proof,
	}

	// print json direct, as we don't need to wrap with the height
	res, err := data.ToJSON(post)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}
