package descriptor

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Chain exposes the descriptor for the chain
type Chain interface {
	// Config returns the sdk.Config specific
	Config() Config
	// Deliverables returns the list of deliverable messages
	Deliverables() Deliverables
	// Queriers returns the list of available queriers
	Queriers() Queriers
}

// Deliverables allows to interact with a list of deliverables
type Deliverables interface {
	// Len returns the number of deliverables
	Len() int
	// Get returns the deliverable by index
	Get(i int) Deliverable
	// ByName returns the deliverable given its name
	// if not found nil will be returned
	ByName(string) Deliverable
}

// Deliverable defines the descriptor for the given message
type Deliverable interface {
	// Descriptor returns the protoreflect.MessageDescriptor
	Descriptor() protoreflect.MessageDescriptor
	// ExpectedSigners // TODO
}

// Queriers exposes the queriers descriptor interface
type Queriers interface {
	// Len returns the number of queriers
	Len() int
	// Get returns the querier by index
	Get(i int) Querier
	// ByInput returns the querier given its input proto.Message
	ByInput(input proto.Message) Querier
	// ByName gets the Querier descriptor
	// given its service method name
	ByName(name string) Querier
	// ByTMName gets the Querier descriptor
	// given the tendermint query path
	ByTMName(path string) Querier
}

// Querier exposes the query descriptor interface
type Querier interface {
	// TMQueryPath returns the path that needs to be used in tendermint to execute the query
	TMQueryPath() string
	// Descriptor returns the protoreflect.MethodDescriptor
	Descriptor() protoreflect.MethodDescriptor
}

// Config provides reflection capabilities for sdk.Config type
type Config interface {
	// Bech32Prefix returns the chain bech32 prefix
	Bech32Prefix() string
	// Bech32AccAddressPrefix returns the account address prefix
	Bech32AccAddressPrefix() string
	// Bech32AccPubPrefix returns the bech32 prefix of an account's public key
	Bech32AccPubPrefix() string
	// Bech32ValAddrPrefix returns the bech32 prefix of a validator's operator address
	Bech32ValAddrPrefix() string
	// Bech32ValPubPrefix returns the bech32 prefix of validator's operator public key
	Bech32ValPubPrefix() string
	// Bech32ConsAddrPrefix returns the bech32 prefix of a consensus node address
	Bech32ConsAddrPrefix() string
	// Bech32ConsPubPrefix returns the bech32 prefix of a consensus node public key
	Bech32ConsPubPrefix() string
	// Purpose returns the purpose as defined in SLIP44
	Purpose() uint
	// CoinType returns the coin type as defined in SLIP44
	CoinType() uint
}
