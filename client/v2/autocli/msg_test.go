package autocli

//func TestBuildMsgMethodCommand(t *testing.T) {
//	b := &Builder{}
//	options := []*autocliv1.RpcCommandOptions{
//		{
//			RpcMethod:  "Echo",
//			Use:        "echo [pos1] [pos2] [pos3...]",
//			Version:    "1.0",
//			Alias:      []string{"e"},
//			SuggestFor: []string{"eco"},
//			Example:    "echo 1 abc {}",
//			Short:      "echo echos the value provided by the user",
//			Long:       "echo echos the value provided by the user as a proto JSON object with populated with the provided fields and positional arguments",
//			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
//				{
//					ProtoField: "positional1",
//				},
//				{
//					ProtoField: "positional2",
//				},
//				{
//					ProtoField: "positional3_varargs",
//					Varargs:    true,
//				},
//			},
//			FlagOptions: map[string]*autocliv1.FlagOptions{
//				"u32": {
//					Name:      "uint32",
//					Shorthand: "u",
//					Usage:     "some random uint32",
//				},
//				"i32": {
//					Usage:        "some random int32",
//					DefaultValue: "3",
//				},
//				"u64": {
//					Usage:             "some random uint64",
//					NoOptDefaultValue: "5",
//				},
//				"deprecated_field": {
//					Deprecated: "don't use this",
//				},
//				"shorthand_deprecated_field": {
//					Shorthand:  "s",
//					Deprecated: "bad idea",
//				},
//				"hidden_bool": {
//					Hidden: true,
//				},
//			},
//		},
//	}
//	//
//	//serviceDEscriptor := autocliv1.ServiceCommandDescriptor{
//	//	Service: testpb
//	//}
//}
