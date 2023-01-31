package autocli

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

var testCmdMsgDesc = &autocliv1.ServiceCommandDescriptor{
	Service: "testpb.Msg",
	RpcCommandOptions: []*autocliv1.RpcCommandOptions{
		{
			RpcMethod:  "Msg",
			Use:        "msg [pos1] [pos2] [pos3...]",
			Version:    "1.0",
			Alias:      []string{"m"},
			SuggestFor: []string{"msg"},
			Example:    "msg 1 abc {}",
			Short:      "msg msg the value provided by the user",
			Long:       "msg msg the value provided by the user as a proto JSON object with populated with the provided fields and positional arguments",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{
					ProtoField: "positional1",
				},
				{
					ProtoField: "positional2",
				},
			},
			FlagOptions: map[string]*autocliv1.FlagOptions{
				"u32": {
					Name:      "uint32",
					Shorthand: "u",
					Usage:     "some random uint32",
				},
				"i32": {
					Usage:        "some random int32",
					DefaultValue: "3",
				},
				"u64": {
					Usage:             "some random uint64",
					NoOptDefaultValue: "5",
				},
				"deprecated_field": {
					Deprecated: "don't use this",
				},
				"shorthand_deprecated_field": {
					Shorthand:  "s",
					Deprecated: "bad idea",
				},
				"hidden_bool": {
					Hidden: true,
				},
			},
		},
	},
	SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
		// we test the sub-command functionality using the same service with different options
		"deprecatedmsg": {
			Service: "testpb.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:  "Msg",
					Deprecated: "dont use this",
				},
			},
		},
	},
}
