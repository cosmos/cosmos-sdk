package cli

const (
	FlagProve = "prove"
)

// TODO
/*
func object(cdc *codec.Codec, storeKey string, prefix []byte, portid, chanid string, connids []string) channel.Object {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewManager(base, connman)
	return man.CLIObject(portid, chanid, connids)
}

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcQueryCmd := &cobra.Command{
		Use:                        "connection",
		Short:                      "Channel query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcQueryCmd.AddCommand(cli.GetCommands(
	//		GetCmdQueryChannel(storeKey, cdc),
	)...)
	return ibcQueryCmd
}

func QueryChannel(ctx context.CLIContext, obj channel.Object, prove bool) (res utils.JSONObject, err error) {
	conn, connp, err := obj.ChannelCLI(ctx)
	if err != nil {
		return
	}
	avail, availp, err := obj.AvailableCLI(ctx)
	if err != nil {
		return
	}
	/*
		kind, kindp, err := obj.Kind(ctx)
		if err != nil {
			return
		}

	seqsend, seqsendp, err := obj.SeqSendCLI(ctx)
	if err != nil {
		return
	}

	seqrecv, seqrecvp, err := obj.SeqRecvCLI(ctx)
	if err != nil {
		return
	}

	if prove {
		return utils.NewJSONObject(
			conn, connp,
			avail, availp,
			//			kind, kindp,
			seqsend, seqsendp,
			seqrecv, seqrecvp,
		), nil
	}

	return utils.NewJSONObject(
		conn, nil,
		avail, nil,
		seqsend, nil,
		seqrecv, nil,
	), nil
}


func GetCmdQueryChannel(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Query stored connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			obj := object(ctx, cdc, storeKey, ibc.Version, args[0], args[1])
			jsonobj, err := QueryChannel(ctx, obj, viper.GetBool(FlagProve))
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, jsonobj))

			return nil
		},
	}

	cmd.Flags().Bool(FlagProve, false, "(optional) show proofs for the query results")

	return cmd
}
*/
