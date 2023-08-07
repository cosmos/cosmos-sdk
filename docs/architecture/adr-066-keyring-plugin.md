# ADR-066: Keyring plugin

## Changelog

* Jun 20, 2023: Initial Draft (@JulianToledano & @bizk)

## Status

DRAFT

## Abstract

This ADR describes a keyring implementation based on the hashicorp plugins over gRPC.

## Context

Currently, in the cosmos-sdk, the keyring implementation depends on
[99designs/keyring](https://github.com/99designs/keyring) module that isn't under actively maintenance.

For that reason, it is proposed to develop a new keyring implementation that leverages
HashiCorp plugins over gRPC. These can be abstracted and implemented in any language,
while maintaining its capacity to use the rest of the cosmos-sdk features.
This approach adds extendability to the system itself, and provides more adoption.

## Alternatives

There are several other options available, such as plugins over RPC or utilizing the Go standard
library plugin package. However, in the end, all these plugin alternatives share similarities.

Another alternative is to reimplement the keystore [db](https://github.com/cosmos/cosmos-sdk/blob/v0.47.3/crypto/keyring/keyring.go#L207)
interface. This db is where the `99designs` dependency lays. 

## Decision

We will define the gRPC service with all the necessary messages. Additionally, we will
implement one plugin for each backend currently supported in the current keyring.


### gRPC Definitions

The idea is to define a service method for each method already present in the keyring interface.
Similarly, a request/response message will be defined for each service method.

#### Service Definition

```protobuf
service KeyringService {
  rpc Backend(BackendRequest) returns (BackendResponse);
  rpc List(ListRequest) returns (stream ListResponse);
  rpc SupportedAlgorithms(SupportedAlgorithmsRequest) returns (SupportedAlgorithmsReturns);
  rpc Key(KeyRequest ) returns (KeyReturns);
  rpc KeyByAddress(KeyByAddressRequest ) returns (KeyByAddressReturns);
  rpc Delete(DeleteRequest ) returns (DeleteReturns);
  rpc DeleteByAddress(DeleteByAddressRequest ) returns (DeleteByAddressReturns);
  rpc Rename(RenameRequest) returns (RenameReturns);
  rpc NewMnemonic(NewMnemonicRequest) returns (NewMnemonicReturns);
  rpc NewAccount(NewAccountRequest) returns (NewAccountReturns);
  rpc SaveLedgerKey(SaveLedgerKeyRequest) returns (SaveLedgerKeyReturns);
  rpc SaveOfflineKey(SaveOfflineKeyRequest)returns (SaveOfflineKeyReturns);
  rpc SaveMultisig(SaveMultisigRequest) returns (SaveMultisigReturns);
  rpc Sign(SignRequest) returns (SignReturns);
  rpc SignByAddress(SignByAddressRequest) returns (SignByAddressReturns);
  rpc ImportPrivKey(ImportPrivKeyRequest) returns (ImportPrivKeyReturns);
  rpc ImportPubKey(ImportPubKeyRequest) returns (ImportPubKeyReturns);
  rpc ExportPubKeyArmor(ExportPubKeyArmorRequest) returns (ExportPubKeyArmorReturns);
  rpc ExportPubKeyArmorByAddress(ExportPubKeyArmorByAddressRequest) returns (ExportPubKeyArmorByAddressReturns);
  rpc ExportPrivKeyArmor(ExportPrivKeyArmorRequest) returns (ExportPrivKeyArmorReturns);
  rpc ExportPrivKeyArmorByAddress(ExportPrivKeyArmorByAddressRequest) returns (ExportPrivKeyArmorByAddressReturns);
}
```

#### Messages Definition

```protobuf
syntax = "proto3";

message BackendRequest {}
message BackendResponse {
  string backend = 1;
}

message ListRequest {}
message ListResponse {
  bytes record = 1;
}
message SupportedAlgorithmsRequest {}
message SupportedAlgorithmsResponse {
  repeated string supportedAlgorithms = 1;
  repeated string supportedAlgorithmsLedger = 2;
}

message KeyRequest {
  string uid = 1;
}
message KeyResponse {
  bytes record = 1;
}

message KeyByAddressRequest {
  bytes address = 1;
}
message KeyByAddressResponse {
  bytes record = 1;
}

message DeleteRequest {
  string uid = 1;
}
message DeleteResponse {}

message DeleteByAddressRequest {
  bytes address = 1;
}
message DeleteByAddressResponse {}

message RenameRequest {
  string from = 1;
  string to = 2;
}
message RenameResponse {}

message NewMnemonicRequest {
  string uid = 1;
  int32 language = 2;
  string mnemonic = 3;
  string bip39Passphrase = 4;
  string hdPath = 5;
  string keyType = 6;
}
message NewMnemonicResponse {
  bytes record = 1;
  string mnemonic = 2;
}

message NewAccountRequest {
  string uid = 1;
  string mnemonic = 2;
  string bip39Passphrase = 3;
  string hdPath = 4;
  string keyType = 5;
}
message NewAccountResponse{
  bytes record = 1;
}

message SaveLedgerKeyRequest {
  string uid = 1;
  string keyType = 2;
  string hrp =3;
  uint32 coinType = 4;
  uint32 account = 5;
  uint32 index = 6;
}
message SaveLedgerKeyResponse{}

message SaveOfflineKeyRequest {
  string uid = 1;
  google.protobuf.Any pubKey = 2;
}
message SaveOfflineKeyResponse {
  bytes record = 1;
}

message SaveMultisigRequest {
  string uid = 1;
  google.protobuf.Any pubKey = 2;
}
message SaveMultisigResponse {
  bytes record = 1;
}

message SignRequest {
  string uid = 1;
  bytes msg = 2;
  cosmos.tx.signing.v1beta1.SignMode signMode = 3;
}
message SignResponse {
  bytes signedMsg = 1;
  bytes google.protobuf.Any = 2;
}

message SignByAddressRequest {
  bytes address = 1;
  bytes msg = 2;
  cosmos.tx.signing.v1beta1.SignMode signMode = 3;
}
message SignByAddressResponse {
  bytes signedMsg = 1;
  bytes google.protobuf.Any = 2;
}
message ImportPrivKeyRequest {
  string uid = 1;
  string armor = 2;
  string passphrase = 3;
}
message ImportPrivKeyResponse {}

message ImportPubKeyRequest {
  string uid = 1;
  string armor = 2;
}
message ImportPubKeyResponse{}

message ExportPubKeyArmorRequest {
  string uid = 1;
}
message ExportPubKeyArmorResponse{
  string armor = 1;
}

message ExportPubKeyArmorByAddressRequest {
  bytes address = 1;
}
message ExportPubKeyArmorByAddressResponse{
  string armor = 1;
}

message ExportPrivKeyArmorRequest {
  string uid = 1;
  string encryptPassphrase = 2;
}
message ExportPrivKeyArmorResponse {
  string armor = 1;
}

message ExportPrivKeyArmorByAddressRequest {
  bytes address = 1;
}
message ExportPrivKeyArmorByAddressResponse {
  string armor = 1;
} 
```

Some messages are identical and could be consolidated into a single message.
However, creating separate messages for each RPC allows for individual extensibility in the future.


### Types and Interfaces

The idea is to avoid defining any new concrete types outside the required protobuf
messages and the keyring itself. As a result, this implementation will rely on the
currently defined types, such as `Record`, `PubKey`, and `PrivKey`, in the same manner as
the existing keyring does.

#### Keyring Plugin GRPC Definition

`PluginKeyring` interface implements the gRPC service defined before in the proto files.

```go
type PluginKeyring interface {
	Backend(BackendRequest) BackendResponse
	List(ListRequest)  ListResponse
	SupportedAlgorithms(SupportedAlgorithmsRequest) SupportedAlgorithmsReturns
	Key(KeyRequest ) KeyReturns
	KeyByAddress(KeyByAddressRequest ) KeyByAddressReturns
	Delete(DeleteRequest ) DeleteReturns
	DeleteByAddress(DeleteByAddressRequest ) DeleteByAddressReturns
	Rename(RenameRequest) RenameReturns
	NewMnemonic(NewMnemonicRequest) NewMnemonicReturns
	NewAccount(NewAccountRequest) NewAccountReturns
	SaveLedgerKey(SaveLedgerKeyRequest) SaveLedgerKeyReturns
	SaveOfflineKey(SaveOfflineKeyRequest)SaveOfflineKeyReturns
	SaveMultisig(SaveMultisigRequest) SaveMultisigReturns
	Sign(SignRequest) SignReturns
	SignByAddress(SignByAddressRequest) SignByAddressReturns
	ImportPrivKey(ImportPrivKeyRequest) ImportPrivKeyReturns
	ImportPubKey(ImportPubKeyRequest) ImportPubKeyReturns
	ExportPubKeyArmor(ExportPubKeyArmorRequest) ExportPubKeyArmorReturns
	ExportPubKeyArmorByAddress(ExportPubKeyArmorByAddressRequest) ExportPubKeyArmorByAddressReturns
	ExportPrivKeyArmor(ExportPrivKeyArmorRequest) ExportPrivKeyArmorReturns
	ExportPrivKeyArmorByAddress(ExportPrivKeyArmorByAddressRequest) ExportPrivKeyArmorByAddressReturns
}
```

#### Client

Client is what the main process will instantiate.

```go
type pluginClient interface {
    PluginKeyring
}

type Client struct {
    client KeyringServiceClient // generated with protoc
}

// This is how methods will look
func (c *Client) Backend(r *BackendRequest) (*BackendResponse, error) {
    return c.client.Backend(context.Background(), r)
}

```

#### Server

Server is what plugins will instantiate.

```go
type PluginServer interface {
    KeyringServiceServer // generated with protoc
}

type Server struct {
    UnimplementedKeyringServiceServer // generated with protoc 
    Impl PluginKeyring // This is the real implementation
}

// This is how methods will look
func (s Server) Backend(ctx context.Context, r *BackendRequest) (*BackendResponse, error) {
    return s.Impl.Backend(r)
}
```

#### KeyringGrpc

```go
type keyringGRPC interface {
	plugin.Plugin       // hashicorp/go-plugin
	plugin.GRPCPlugin   // hashicorp/go-plugin
}

// KeyringGRPC fulfils hashicorp GRPCPlugin interface with methods:
//  GRPCServer: called internally at plugin.Serve() 
//  GRPCClient: called internally at ClientProtocol.Dispense()
type KeyringGRPC struct {
    plugin.Plugin
    Impl PluginKeyring
}

// this method is what the plugins will call.
func (p *KeyringGRPC) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
    // RegisterKeyringServiceServer is generated with protoc
    RegisterKeyringServiceServer(s, &Server{Impl: p.Impl}) // Server is defined above
    return nil
}

// this method is what the main process (keyring) will call. Instantiating a grpc client.
func (p *KeyringGRPC) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
    // Client is defined above
    // NewKeyringServiceClient is generated with protoc
    return &Client{Client: NewKeyringServiceClient(c)}, nil
}
```

#### Keyring Concrete Type

```go
import "github.com/hashicorp/go-plugin"

// this is what will implement SDK keyring
type PluginsKeyStore struct {
	client  *plugin.Client
	backEnd PluginKeyring // defined above
	cdc     codec.Codec
}

func NewKeyring(...) *PluginsKeyStore {
	client := plugin.NewClient(&plugin.ClientConfig{...})
	// Connect via RPC
	rpcClient, _ := client.Client()
	
	// Request the keyStore
	raw, _ := rpcClient.Dispense("keyring") // here's where KeyringGRPC.GRPCClient() will be called

	return &PluginsKeyStore{
		client:  client,
		backEnd: raw.(PluginKeyring), // PluginKeyring is defined above
		cdc:     cdc,
	}
```

#### Plugin Concrete type

```go
import "github.com/hashicorp/go-plugin"

// backend is responsible for storing keys. It will have different flavors.
type backend interface {
	Get(string) (secret, error)
	Set(string) error
	Remove(string) error
}

type keyring struct {
    cdc codec.Codec
    db backend
}

func newKeyring(...) *keyring {...}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: ...,
		Plugins: map[string]plugin.Plugin{
			"keyring": &KeyringGRPC{Impl: newKeyring(...)},
		},
		// A non-nil value here enables gRPC serving...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

```

### Security

Plugins must run in the same local network as the main process thus some security is guaranteed as the data never
leaves the same machine. In any case connection can be secured with tls. 

### Project Structure

This new keyring could be implemented either in the keyring package of the SDK or in a new repository.

#### New Repository

If it is implemented in a different repository, the `keyring.New` method should be updated to support
the instantiation of a keyring for plugins.

A new dependency will be added to the SDK.

#### Keyring Package

If the new keyring implementation is defined in the keyring package, it would be beneficial to
consider the following cosmetic refactors:

* keyring.go: This file should define abstract types only, along with some common functions.
* keyStore: Move the [keyStore](https://github.com/cosmos/cosmos-sdk/blob/v0.47.3/crypto/keyring/keyring.go#L206) code to its own file, along with all its associated functions. 
* factory.go: Create a factory file responsible for instantiating different keyrings. 


## Consequences

### Backwards Compatibility

Backwards compatibility is guaranteed as the current keyring implementation and this new one can coexist.

### Positive

As plugins communicate with the main process using gRPC, plugins can be written in any language.

Teams can easily develop plugins to meet their specific requirements,
which opens the door to new functionalities in key management.

### Negative

The `Record` relies on its `cachedValue` field to retrieve the address, public and private keys. This
field cannot be sent over gRPC. This will generate occasions when the `Record` must be deserialized
both in the plugin and the SDK (main process), leading to some overhead.

### Neutral

Since plugins are separate subprocesses initiated from the main process, it is important to close
these subprocesses properly. To achieve this, the current keyring interface should be extended with
a `Close()` method.

Some work may be needed to provide a way to migrate keys between the current keyring implementation and 
this new one.

## Further Discussions


## References

* https://github.com/cosmos/cosmos-sdk/issues/14940
* [Keyring Plugins Poc Implementation](https://github.com/Zondax/keyringPoc)
* [Hashicorp plugins](https://github.com/hashicorp/go-plugin)

