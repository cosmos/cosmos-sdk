package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

const (
	FlagNode1 = "node1"
	FlagNode2 = "node2"
	FlagFrom1 = "from1"
	FlagFrom2 = "from2"
)

func handshake(cdc *codec.Codec, storeKey string, prefix []byte, portId, chanId, connId string) channel.HandshakeState {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	clientManager := client.NewManager(base)
	connectionManager := connection.NewManager(base, clientManager)
	man := channel.NewHandshaker(channel.NewManager(base, connectionManager))
	return man.CLIObject(portId, chanId, []string{connId})
}

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "IBC channel transaction subcommands",
	}

	cmd.AddCommand(
		GetCmdHandshake(storeKey, cdc),
	)

	return cmd
}

// TODO: move to 02/tendermint
func getHeader(ctx context.CLIContext) (res tendermint.Header, err error) {
	node, err := ctx.GetNode()
	if err != nil {
		return
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return
	}

	height := info.Response.LastBlockHeight
	prevHeight := height - 1

	commit, err := node.Commit(&height)
	if err != nil {
		return
	}

	validators, err := node.Validators(&prevHeight)
	if err != nil {
		return
	}

	nextValidators, err := node.Validators(&height)
	if err != nil {
		return
	}

	res = tendermint.Header{
		SignedHeader:     commit.SignedHeader,
		ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
		NextValidatorSet: tmtypes.NewValidatorSet(nextValidators.Validators),
	}

	return
}

func GetCmdHandshake(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handshake",
		Short: "initiate connection handshake between two chains",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom1)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1)).
				WithBroadcastMode(flags.BroadcastBlock)
			q1 := state.NewCLIQuerier(ctx1)

			ctx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2)).
				WithBroadcastMode(flags.BroadcastBlock)
			q2 := state.NewCLIQuerier(ctx2)

			portid1 := args[0]
			chanid1 := args[1]
			connid1 := args[2]
			portid2 := args[3]
			chanid2 := args[4]
			connid2 := args[5]

			chan1 := channel.Channel{
				Counterparty:     chanid2,
				CounterpartyPort: portid2,
				ConnectionHops:   []string{connid1},
			}

			chan2 := channel.Channel{
				Counterparty:     chanid1,
				CounterpartyPort: portid1,
				ConnectionHops:   []string{connid2},
			}

			obj1 := handshake(cdc, storeKey, version.DefaultPrefix(), portid1, chanid1, connid1)
			obj2 := handshake(cdc, storeKey, version.DefaultPrefix(), portid2, chanid2, connid2)

			conn1, _, err := obj1.OriginConnection().ConnectionCLI(q1)
			if err != nil {
				return err
			}

			clientId1 := conn1.Client
			conn2, _, err := obj2.OriginConnection().ConnectionCLI(q2)
			if err != nil {
				return err
			}
			clientId2 := conn2.Client

			// TODO: check state and if not Idle continue existing process
			msgInit := channel.MsgOpenInit{
				PortID:    portid1,
				ChannelID: chanid1,
				Channel:   chan1,
				Signer:    ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgInit})
			if err != nil {
				return err
			}

			// Another block has to be passed after msgInit is commited
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err := getHeader(ctx1)
			if err != nil {
				return err
			}

			msgUpdate := client.MsgUpdateClient{
				ClientID: clientId2,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgUpdate})

			fmt.Printf("updated apphash to %X\n", header.AppHash)

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			fmt.Printf("querying from %d\n", header.Height-1)

			_, pchan, err := obj1.ChannelCLI(q1)
			if err != nil {
				return err
			}
			_, pstate, err := obj1.StateCLI(q1)
			if err != nil {
				return err
			}

			msgTry := channel.MsgOpenTry{
				PortID:    portid2,
				ChannelID: chanid2,
				Channel:   chan2,
				Proofs:    []commitment.Proof{pchan, pstate},
				Height:    uint64(header.Height),
				Signer:    ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgTry})
			if err != nil {
				return err
			}

			// Another block has to be passed after msgInit is commited
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx2)
			if err != nil {
				return err
			}

			msgUpdate = client.MsgUpdateClient{
				ClientID: clientId1,
				Header:   header,
				Signer:   ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgUpdate})

			q2 = state.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))

			_, pchan, err = obj2.ChannelCLI(q2)
			if err != nil {
				return err
			}
			_, pstate, err = obj2.StateCLI(q2)
			if err != nil {
				return err
			}

			msgAck := channel.MsgOpenAck{
				PortID:    portid1,
				ChannelID: chanid1,
				Proofs:    []commitment.Proof{pchan, pstate},
				Height:    uint64(header.Height),
				Signer:    ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgAck})
			if err != nil {
				return err
			}

			// Another block has to be passed after msgInit is commited
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx1)
			if err != nil {
				return err
			}

			msgUpdate = client.MsgUpdateClient{
				ClientID: clientId2,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgUpdate})

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))

			_, pstate, err = obj1.StateCLI(q1)
			if err != nil {
				return err
			}

			msgConfirm := channel.MsgOpenConfirm{
				PortID:    portid2,
				ChannelID: chanid2,
				Proofs:    []commitment.Proof{pstate},
				Height:    uint64(header.Height),
				Signer:    ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgConfirm})
			if err != nil {
				return err
			}

			return nil
		},
	}

	// TODO: Create flag description
	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "")
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "")
	cmd.Flags().String(FlagFrom1, "", "")
	cmd.Flags().String(FlagFrom2, "", "")

	cmd.MarkFlagRequired(FlagFrom1)
	cmd.MarkFlagRequired(FlagFrom2)

	return cmd
}
