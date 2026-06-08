package autocli

// ModuleOptions describes the CLI options for a Cosmos SDK module.
type ModuleOptions struct {
	// Tx describes the tx commands for the module.
	Tx *ServiceCommandDescriptor

	// Query describes the query commands for the module.
	Query *ServiceCommandDescriptor
}

// ServiceCommandDescriptor describes a CLI command based on a protobuf service.
type ServiceCommandDescriptor struct {
	// Service is the fully qualified name of the protobuf service to build the
	// command from. It can be left empty if SubCommands are used instead.
	Service string

	// RpcCommandOptions are options for commands generated from rpc methods.
	// If no options are specified for a given rpc method, a command will be
	// generated for that method with the default options.
	RpcCommandOptions []*RpcCommandOptions

	// SubCommands is a map of optional sub-commands for this command based on
	// different protobuf services. The map key is used as the name of the
	// sub-command.
	SubCommands map[string]*ServiceCommandDescriptor

	// EnhanceCustomCommand specifies whether to enhance an existing custom
	// command with the services from gRPC. If false and a custom command
	// already exists, no commands will be generated for the service.
	EnhanceCustomCommand bool

	// Short is an optional override for the short description of the auto
	// generated command.
	Short string
}

// RpcCommandOptions specifies options for commands generated from protobuf rpc
// methods.
type RpcCommandOptions struct {
	// RpcMethod is the short name of the protobuf rpc method this command is
	// generated from.
	RpcMethod string

	// Use is the one-line usage method. It also allows specifying an alternate
	// name for the command as the first word of the usage text.
	Use string

	// Long is the long message shown in the 'help <this-command>' output.
	Long string

	// Short is the short description shown in the 'help' output.
	Short string

	// Example is examples of how to use the command.
	Example string

	// Alias is an array of aliases that can be used instead of the first word
	// in Use.
	Alias []string

	// SuggestFor is an array of command names for which this command will be
	// suggested — similar to aliases but only suggests.
	SuggestFor []string

	// Deprecated defines, if this command is deprecated and should print this
	// string when used.
	Deprecated string

	// Version defines the version for this command.
	Version string

	// FlagOptions are options for flags generated from rpc request fields.
	// By default all request fields are configured as flags. They can also be
	// configured as positional args instead using PositionalArgs.
	FlagOptions map[string]*FlagOptions

	// PositionalArgs specifies positional arguments for the command.
	PositionalArgs []*PositionalArgDescriptor

	// Skip specifies whether to skip this rpc method when generating commands.
	Skip bool

	// GovProposal specifies whether autocli should generate a gov proposal
	// transaction for this rpc method instead of a plain transaction.
	GovProposal bool
}

// FlagOptions are options for flags generated from rpc request fields.
type FlagOptions struct {
	// Name is an alternate name to use for the field flag.
	Name string

	// Shorthand is a one-letter abbreviated flag.
	Shorthand string

	// Usage is the help message.
	Usage string

	// DefaultValue is the default value as text.
	DefaultValue string

	// Deprecated is the usage text to show if this flag is deprecated.
	Deprecated string

	// ShorthandDeprecated is the usage text to show if the shorthand of this
	// flag is deprecated.
	ShorthandDeprecated string

	// Hidden hides the flag from help/usage text.
	Hidden bool
}

// PositionalArgDescriptor describes a positional argument.
type PositionalArgDescriptor struct {
	// ProtoField specifies the proto field to use as the positional arg.
	ProtoField string

	// Varargs makes a positional parameter a varargs parameter. This can only
	// be applied to the last positional parameter and the proto field must be
	// a repeated field. Mutually exclusive with Optional.
	Varargs bool

	// Optional makes the last positional parameter optional.
	// Mutually exclusive with Varargs.
	Optional bool
}
