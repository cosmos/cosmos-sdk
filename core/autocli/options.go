package autocli

import (
	gogoproto "github.com/cosmos/gogoproto/proto"
)

func init() {
	gogoproto.RegisterType((*ModuleOptions)(nil), "cosmos.autocli.v1.ModuleOptions")
	gogoproto.RegisterType((*ServiceCommandDescriptor)(nil), "cosmos.autocli.v1.ServiceCommandDescriptor")
	gogoproto.RegisterType((*RpcCommandOptions)(nil), "cosmos.autocli.v1.RpcCommandOptions")
	gogoproto.RegisterType((*FlagOptions)(nil), "cosmos.autocli.v1.FlagOptions")
	gogoproto.RegisterType((*PositionalArgDescriptor)(nil), "cosmos.autocli.v1.PositionalArgDescriptor")
}

// ModuleOptions describes the CLI options for a Cosmos SDK module.
type ModuleOptions struct {
	Tx    *ServiceCommandDescriptor `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	Query *ServiceCommandDescriptor `protobuf:"bytes,2,opt,name=query,proto3" json:"query,omitempty"`
}

func (*ModuleOptions) Reset()         {}
func (*ModuleOptions) String() string { return "" }
func (*ModuleOptions) ProtoMessage()  {}

// ServiceCommandDescriptor describes a CLI command based on a protobuf service.
type ServiceCommandDescriptor struct {
	Service              string                               `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	RpcCommandOptions    []*RpcCommandOptions                 `protobuf:"bytes,2,rep,name=rpc_command_options,json=rpcCommandOptions,proto3" json:"rpc_command_options,omitempty"`
	SubCommands          map[string]*ServiceCommandDescriptor `protobuf:"bytes,3,rep,name=sub_commands,json=subCommands,proto3" json:"sub_commands,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	EnhanceCustomCommand bool                                 `protobuf:"varint,4,opt,name=enhance_custom_command,json=enhanceCustomCommand,proto3" json:"enhance_custom_command,omitempty"`
	Short                string                               `protobuf:"bytes,5,opt,name=short,proto3" json:"short,omitempty"`
}

func (*ServiceCommandDescriptor) Reset()         {}
func (*ServiceCommandDescriptor) String() string { return "" }
func (*ServiceCommandDescriptor) ProtoMessage()  {}

// RpcCommandOptions specifies options for commands generated from protobuf rpc methods.
type RpcCommandOptions struct {
	RpcMethod      string                     `protobuf:"bytes,1,opt,name=rpc_method,json=rpcMethod,proto3" json:"rpc_method,omitempty"`
	Use            string                     `protobuf:"bytes,2,opt,name=use,proto3" json:"use,omitempty"`
	Long           string                     `protobuf:"bytes,3,opt,name=long,proto3" json:"long,omitempty"`
	Short          string                     `protobuf:"bytes,4,opt,name=short,proto3" json:"short,omitempty"`
	Example        string                     `protobuf:"bytes,5,opt,name=example,proto3" json:"example,omitempty"`
	Alias          []string                   `protobuf:"bytes,6,rep,name=alias,proto3" json:"alias,omitempty"`
	SuggestFor     []string                   `protobuf:"bytes,7,rep,name=suggest_for,json=suggestFor,proto3" json:"suggest_for,omitempty"`
	Deprecated     string                     `protobuf:"bytes,8,opt,name=deprecated,proto3" json:"deprecated,omitempty"`
	Version        string                     `protobuf:"bytes,9,opt,name=version,proto3" json:"version,omitempty"`
	FlagOptions    map[string]*FlagOptions    `protobuf:"bytes,10,rep,name=flag_options,json=flagOptions,proto3" json:"flag_options,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	PositionalArgs []*PositionalArgDescriptor `protobuf:"bytes,11,rep,name=positional_args,json=positionalArgs,proto3" json:"positional_args,omitempty"`
	Skip           bool                       `protobuf:"varint,12,opt,name=skip,proto3" json:"skip,omitempty"`
	GovProposal    bool                       `protobuf:"varint,13,opt,name=gov_proposal,json=govProposal,proto3" json:"gov_proposal,omitempty"`
}

func (*RpcCommandOptions) Reset()         {}
func (*RpcCommandOptions) String() string { return "" }
func (*RpcCommandOptions) ProtoMessage()  {}

// FlagOptions are options for flags generated from rpc request fields.
type FlagOptions struct {
	Name                string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Shorthand           string `protobuf:"bytes,2,opt,name=shorthand,proto3" json:"shorthand,omitempty"`
	Usage               string `protobuf:"bytes,3,opt,name=usage,proto3" json:"usage,omitempty"`
	DefaultValue        string `protobuf:"bytes,4,opt,name=default_value,json=defaultValue,proto3" json:"default_value,omitempty"`
	Deprecated          string `protobuf:"bytes,6,opt,name=deprecated,proto3" json:"deprecated,omitempty"`
	ShorthandDeprecated string `protobuf:"bytes,7,opt,name=shorthand_deprecated,json=shorthandDeprecated,proto3" json:"shorthand_deprecated,omitempty"`
	Hidden              bool   `protobuf:"varint,8,opt,name=hidden,proto3" json:"hidden,omitempty"`
}

func (*FlagOptions) Reset()         {}
func (*FlagOptions) String() string { return "" }
func (*FlagOptions) ProtoMessage()  {}

// PositionalArgDescriptor describes a positional argument.
type PositionalArgDescriptor struct {
	ProtoField string `protobuf:"bytes,1,opt,name=proto_field,json=protoField,proto3" json:"proto_field,omitempty"`
	Varargs    bool   `protobuf:"varint,2,opt,name=varargs,proto3" json:"varargs,omitempty"`
	Optional   bool   `protobuf:"varint,3,opt,name=optional,proto3" json:"optional,omitempty"`
}

func (*PositionalArgDescriptor) Reset()         {}
func (*PositionalArgDescriptor) String() string { return "" }
func (*PositionalArgDescriptor) ProtoMessage()  {}
