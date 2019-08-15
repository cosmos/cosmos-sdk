package cli

/*
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {

}
*/
const (
	FlagNode1 = "node1"
	FlagNode2 = "node2"
	FlagFrom1 = "from1"
	FlagFrom2 = "from2"
)

// TODO
/*
func handshake(ctx context.CLIContext, cdc *codec.Codec, storeKey string, prefix []byte, portid, chanid string) (channel.HandshakeObject, error) {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewHandshaker(channel.NewManager(base, connman))
	return man.CLIQuery(ctx, portid, chanid)
}

func lastheight(ctx context.CLIContext) (uint64, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return 0, err
	}

	return uint64(info.Response.LastBlockHeight), nil
}

func GetCmdChannelHandshake(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handshake",
		Short: "initiate channel handshake between two chains",
		Args:  cobra.ExactArgs(4),
		// Args: []string{connid1, chanid1, chanfilepath1, connid2, chanid2, chanfilepath2}
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContext().
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1)).
				WithFrom(viper.GetString(FlagFrom1))

			ctx2 := context.NewCLIContext().
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2)).
				WithFrom(viper.GetString(FlagFrom2))

			conn1id := args[0]
			chan1id := args[1]
			conn1bz, err := ioutil.ReadFile(args[2])
			if err != nil {
				return err
			}
			var conn1 channel.Channel
			if err := cdc.UnmarshalJSON(conn1bz, &conn1); err != nil {
				return err
			}

			obj1, err := handshake(ctx1, cdc, storeKey, ibc.Version, conn1id, chan1id)
			if err != nil {
				return err
			}

			conn2id := args[3]
			chan2id := args[4]
			conn2bz, err := ioutil.ReadFile(args[5])
			if err != nil {
				return err
			}
			var conn2 channel.Channel
			if err := cdc.UnmarshalJSON(conn2bz, &conn2); err != nil {
				return err
			}

			obj2, err := handshake(ctx2, cdc, storeKey, ibc.Version, conn1id, chan1id)
			if err != nil {
				return err
			}

			// TODO: check state and if not Idle continue existing process
			height, err := lastheight(ctx2)
			if err != nil {
				return err
			}
			nextTimeout := height + 1000 // TODO: parameterize
			msginit := channel.MsgOpenInit{
				ConnectionID: conn1id,
				ChannelID:    chan1id,
				Channel:      conn1,
				NextTimeout:  nextTimeout,
				Signer:       ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msginit})
			if err != nil {
				return err
			}

			timeout := nextTimeout
			height, err = lastheight(ctx1)
			if err != nil {
				return err
			}
			nextTimeout = height + 1000
			_, pconn, err := obj1.ChannelCLI(ctx1)
			if err != nil {
				return err
			}
			_, pstate, err := obj1.StateCLI(ctx1)
			if err != nil {
				return err
			}
			_, ptimeout, err := obj1.NextTimeoutCLI(ctx1)
			if err != nil {
				return err
			}

			msgtry := channel.MsgOpenTry{
				ConnectionID: conn2id,
				ChannelID:    chan2id,
				Channel:      conn2,
				Timeout:      timeout,
				NextTimeout:  nextTimeout,
				Proofs:       []commitment.Proof{pconn, pstate, ptimeout},
				Signer:       ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgtry})
			if err != nil {
				return err
			}

			timeout = nextTimeout
			height, err = lastheight(ctx2)
			if err != nil {
				return err
			}
			nextTimeout = height + 1000
			_, pconn, err = obj2.Channel(ctx2)
			if err != nil {
				return err
			}
			_, pstate, err = obj2.State(ctx2)
			if err != nil {
				return err
			}
			_, ptimeout, err = obj2.NextTimeout(ctx2)
			if err != nil {
				return err
			}

			msgack := channel.MsgOpenAck{
				ConnectionID: conn1id,
				ChannelID:    chan1id,
				Timeout:      timeout,
				Proofs:       []commitment.Proof{pconn, pstate, ptimeout},
				Signer:       ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgack})
			if err != nil {
				return err
			}

			timeout = nextTimeout
			_, pstate, err = obj1.State(ctx1)
			if err != nil {
				return err
			}
			_, ptimeout, err = obj1.NextTimeout(ctx1)
			if err != nil {
				return err
			}

			msgconfirm := channel.MsgOpenConfirm{
				ConnectionID: conn2id,
				ChannelID:    chan2id,
				Timeout:      timeout,
				Proofs:       []commitment.Proof{pstate, ptimeout},
				Signer:       ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgconfirm})
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func GetCmdRelay(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay",
		Short: "relay pakcets between two channels",
		Args:  cobra.ExactArgs(4),
		// Args: []string{connid1, chanid1, chanfilepath1, connid2, chanid2, chanfilepath2}
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx1 := context.NewCLIContext().
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1)).
				WithFrom(viper.GetString(FlagFrom1))

			ctx2 := context.NewCLIContext().
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2)).
				WithFrom(viper.GetString(FlagFrom2))

			conn1id, chan1id, conn2id, chan2id := args[0], args[1], args[2], args[3]

			obj1 := object(ctx1, cdc, storeKey, ibc.Version, conn1id, chan1id)
			obj2 := object(ctx2, cdc, storeKey, ibc.Version, conn2id, chan2id)

			return relayLoop(cdc, ctx1, ctx2, obj1, obj2, conn1id, chan1id, conn2id, chan2id)
		},
	}

	return cmd
}

func relayLoop(cdc *codec.Codec,
	ctx1, ctx2 context.CLIContext,
	obj1, obj2 channel.CLIObject,
	conn1id, chan1id, conn2id, chan2id string,
) error {
	for {
		// TODO: relay() should be goroutine and return error by channel
		err := relay(cdc, ctx1, ctx2, obj1, obj2, conn2id, chan2id)
		// TODO: relayBetween() should retry several times before halt
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
}

func relay(cdc *codec.Codec, ctxFrom, ctxTo context.CLIContext, objFrom, objTo channel.CLIObject, connidTo, chanidTo string) error {
	txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

	seq, _, err := objTo.SeqRecv(ctxTo)
	if err != nil {
		return err
	}

	sent, _, err := objFrom.SeqSend(ctxFrom)
	if err != nil {
		return err
	}

	for i := seq; i <= sent; i++ {
		packet, proof, err := objFrom.Packet(ctxFrom, seq)
		if err != nil {
			return err
		}

		msg := channel.MsgReceive{
			ConnectionID: connidTo,
			ChannelID:    chanidTo,
			Packet:       packet,
			Proofs:       []commitment.Proof{proof},
			Signer:       ctxTo.GetFromAddress(),
		}

		err = utils.GenerateOrBroadcastMsgs(ctxTo, txBldr, []sdk.Msg{msg})
		if err != nil {
			return err
		}
	}

	return nil
}
*/
