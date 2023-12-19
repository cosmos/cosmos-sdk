# ADR 071: Cryptography v2

## Change log

* Nov 1st 2023: Initial Draft (Zondax AG: @raynaudoe @bizk @juliantoledano @jleni @educlerici-zondax)

## Status

DRAFT

## Abstract

This ADR proposes a refactor of the crypto module to enhance modularity, re-usability, and maintainability,
while prioritizing developer experience and incorporating best security practices.
The proposal defines a clear division of scope for each component, cleaner interfaces, easier extension,
better test coverage and a single place of truth, allowing the developer to focus on what's important
while ensuring the secure handling of sensitive data throughout the module.

## Introduction

This ADR outlines the redesign and refactoring of the crypto package. The design establishes a clear decoupling via interfaces, extension points, and a much more modular design to allow developers to concentrate on application level aspects while ensuring the adequate handling of sensitive data.

Special focus has been placed on the following key aspects:

* modularity
* extensibility
* security
* maintainability
* developer experience

The proposal determines a clear decoupling via interfaces, additional extension points, and a much more modular design to allow developers to application level aspects while ensuring the secure handling of sensitive data when applying this SDK.

The enhancements in this proposal not only render the ["Keyring ADR"](https://github.com/cosmos/cosmos-sdk/issues/14940) obsolete, but also encompass its key aspects, replacing it with a more flexible and comprehensive approach. Furthermore, the gRPC service proposed in the Keyring ADR can be easily implemented as a specialized implementation of the "CryptoProvider" interface defined later in this ADR. This allows for the integration of HashiCorp-like [go-plugins](https://github.com/hashicorp/go-plugin) over gRPC, providing a robust and extensible solution for keyring functionality.

Furthermore, the grpc service proposed in the Keyring ADR can be easily followed by creating an implementation of the "CryptoProvider" interface defined in this ADR. This allows for the integration of HashiCorp plugins over gRPC, providing a robust and extensible solution for keyring functionality.

By deprecating the previous ADR and introducing these enhancements, the new ADR offers a more comprehensive and adaptable solution for cryptography and address management within the Cosmos SDK ecosystem.

### Glossary

1. **Interface**: In the context of this document, "interface" refers to Go's interface concept.

2. **Module**: In this document, "module" refers to a Go module. The proposed ADR focuses on the Crypto module V2, which suggests the introduction of a new version of the Crypto module with updated features and improvements.

3. **Package**: In the context of Go, a "package" refers to a unit of code organization. Each proposed architectural unit will be organized into packages for better reutilization and extension.

## Context

In order to fully understand the need for changes and improvements to the cryptographic package, it's crucial to consider the current state of affairs:

* The Cosmos SDK currently lacks a comprehensive ADR for the cryptographic package.
* Type leakage outside the current crypto module pose backward compatibility and extensibility challenges.
* The demand for a more flexible and extensible approach to cryptography and address management is high.
* Architectural changes are necessary to resolve many of the currently open issues.
* There is a current tread towards modularity in the ICF stack (e.g. runtime modules)
* Security implications are a critical consideration during the redesign work.

## Objectives

The key objectives for the Cryptography v2 module are:

Modular Design Philosophy

* Establish a flexible and extensible foundation using interfaces to enable the seamless integration of various cryptographic methods and tools.

* Restructure, Refactor, and Decouple: Update the cryptography codebase to ensure modularity and future adaptability.

Documentation & Community Engagement

* Cryptography v2 ADR: Draft a new Architecture Decision Record to guide and document the evolution of the module (this document).

* Enhance documentation to ensure clarity, establish a good practices protocol and promote community engagement, providing a platform for feedback and collaborative growth.

Backward Compatibility & Migration

* Prioritize compatibility with previous module version to avoid disruptions for existing users.

* Design and propose a suitable migration path, ensuring transitions are as seamless as possible.

* Evaluate and decide on the relevance of existing systems and tools, incorporating or deprecating them based on their alignment with the module's new vision.

Developer-Centric Approach

* Prioritize clear, intuitive interfaces and best-practice design principles.
* Improve Developer Experience: Provide tools, samples, and best practices to foster an efficient and user-friendly development environment.

Leverage Extensibility

* Utilize the module's modular design to support a wide range of cryptographic tools, key types, and methods, ensuring adaptability for future technological advancements.
* Integrate support for advanced cryptographic features, ensuring the module's position at the forefront of cryptographic technologies.

Quality Assurance

* Enhanced Test Coverage: Improve testing methodologies to ensure the robustness and reliability of the module.
* Conduct an Audit: After implementation, perform a comprehensive audit to identify potential vulnerabilities and ensure the module's security and stability.

## Technical Goals

As technical goals, the aim is to create a robust, flexible, and future-proof cryptographic module. This is achieved through the following key points:

Wide Hardware Device & Cloud-based HSM Interface Support:

* Design a foundational interface for various hardware devices (Ledger, YubiKey, Thales, etc.) and cloud-based HSMs (Amazon, Azure) to cater to both current and future implementations.

Plugin Architecture and Dependency Injection

* Establish the architectural foundation for an extensible plugin system and integrate a dependency injection framework, ensuring modularity, testability, and third-party integrations.

* Design an environment for plugin testing, ensuring developers can validate integrations without compromising system integrity.

Interface considerations

* Design should take into considerations support for Trusted Platform Module (TPM) 2.0 and similar devices to anticipate future enhancements.

* Design should take into account the Cryptographic Token Interface Standard (PKCS#11)

Increase cryptographic versatility

* Support for a broad spectrum of cryptographic techniques
* Extend support for more hash functions (e.g. pedersen, argon2, Argon2d/I/id, Blake3, etc.)
* Extend support for more signature schemes (e.g. secp256r1, ed25519, ed448, sr25519, etc.)
* More advanced methods ( Post-Quantum Cryptography (PQC) methods
* Threshold signatures and encryption

Community Engagement Infrastructure:

* Structure the design with tools and documentation interfaces in mind, enabling a seamless future rollout of resources for developer engagement.

## Proposed architecture

### Introduction

In the proposed architecture, each package is decoupled and isolated. Adding new implementations consist of implementing the required interfaces.

```mermaid
classDiagram

Keyring <|-- Wallet

SecureStorage <|-- Keyring
SecureItem <|-- SecureStorage
CryptoProvider <|-- SecureItem

Hasher <|-- CryptoProvider
Verifier <|-- CryptoProvider
Signer <|-- CryptoProvider
Cipher <|-- CryptoProvider
Generator <|-- CryptoProvider
```

### Crypto Provider

*Crypto Providers* serve as a middleware responsible for managing the interaction with various instantiated cryptographic packages. It acts as a centralized controller, encapsulating the API of the crypto modules in a single location.
Through each Crypto provider, users can access functionality such as signing, verification, encryption, and hashing.

By abstracting the underlying cryptographic functionality, *Crypto providers* enable a modular and extensible architecture. It allows users to easily switch between different cryptographic implementations without impacting the rest of the system.

```go
type ProviderMetadata interface {
  key string
  value string
}

type ICryptoProviderMetadata interface {
  GetTypeUUID() TypeUUID
  GetName() string
  GetMetadata() []ProviderMetadata
}

type ICryptoProviderBuilder interface {
  ICryptoProviderMetadata
  
  FromSecureItem( item SecureItem ) (ICryptoProvider, error)  

  FromRandomness( source IRandomnessSource ) (ICryptoProvider, error)
  FromSeed( seed []byte ) (ICryptoProvider, error)
  FromMnemonic( mnemonic string ) (ICryptoProvider error)
  FromString( url string ) (ICryptoProvider error)
}

type ICryptoProvider interface {
  Proto.Message
  ICryptoProviderMetadata
  
  GetKeys() (PubKey, PrivKey, error)
  GetSigner() (ISigner, error)
  GetVerifier() (IVerifier, error)
  GetCipher() (ICipher, error)
  GetHasher() (IHasher, error)
}
```

#### Signing

Interface responsible for Signing a message and returning the generated Signature.

```go
type ISigner interface {
  Sign(Blob) (Signature, error)
}
```

#### Verifier

Verifies if given a message belongs to a public key by validating against its respective signature.

```go
type IVerifier interface {
  Verify(Blob, Signature) (bool, error)
}
```

#### Cipher

A cipher is an api for encryption and decryption of data. Given a message it should operate through a secret.

```go
type ICipher interface {
  Encrypt(message Blob) (encryptedMessage Blob, error)
  Decrypt(encryptedMessage Blob) (message Blob, error)
}
```

#### Hasher

This package contains the different hashing algorithms and conventions agreed on this matter.

```go
type IHasher interface {
  Hash(input Blob) Blob
  CanHashIncrementally() bool
}
```

### StorageProvider


A *Secure Storage* represents a secure vault where one or more *Secure Items* can be stored. It serves as a centralized repository for securely storing sensitive data. To access a *Secure Item*, users must interact with the *Secure Storage*, which handles the retrieval and management of keys.
Different implementations of *Secure Storage* will be available to cater to various storage requirements:

* FileSystem: This implementation stores the Secure Items in a designated folder within the file system.
* Memory: This implementation stores the Secure Items in memory, providing fast access but limited persistence.
* KMS: This implementation utilizes the Key Management System available on AWS, GCP, etc.
* others: 1password, OS-integrated secure storage (macOS, Linux, Windows, etc.)

```go
type IStorageProvider interface {
  List() []string

  Get(name string) (SecureItem, error)
  Set(item SecureItem) error
  Remove(name string) error
}
```

A *Secure Item* is a structured data object designed for storing any type of data within a *Secure Storage* instance.
In the context of this ADR, the **Blob** field of a Secure Item represents a "recipe" or blueprint for constructing the corresponding *Crypto Provider*.
The **Blob** can be encoded in any format and should contain all the necessary configuration information required to instantiate the specific cryptographic packages that compose the *Crypto Provider*.

```go
type ISecureItemMetadata interface {
  Type()   TypeUUID     // Relates to the corresponding provider
  Name()   string
  ...
}

type ISecureItem interface {
  ISecureItemMetadata

  // Blob format/encoding will be dependant of the CryptoProvider implementation
  Bytes() []byte
}
```

##### Keyring

*Keyring* serves as a central hub for managing *Crypto Providers* and *Secure Storage* implementations. It provides methods to register *Crypto Providers* and *Secure Storage* implementations.
The **RegisterCryptoProvider** function allows users to register a Crypto Provider blueprint by providing its unique identifier and a builder function. Similarly, the **RegisterSecureStorage** function enables users to register a secure storage implementation by specifying a unique identifier and a builder function.


```go
type IKeyring interface {
  RegisterCryptoProvider(typeUUID TypeUUID, builder CryptoProviderBuilder)
  RegisterAndLoadStorageProvider(typeUUID TypeUUID, provider StorageProvider)

  ListStorageProviders() ([]IStorageProvider, error)
  ListCryptoProviders() ([]ICryptoProvider, error)

  List() ([]SecureItemMetadata, error)
  
  GetCryptoProvider(ItemId) (CryptoProvider, error)
}
```

#### **Wallet**

The Wallet interface contains the blockchain specific use cases of the crypto module. It also serves as an API for:

* Signing and Verifying messages.
* Generating addresses out of keys

Since wallet interacts with the user keys, it contains an instance of the Keyring, it is also where the blockchain specific logic should reside.

Note: Each Wallet implementation should provide the logic to map addresses and ItemIds

```go
type Wallet interface {
	Init(Keyring)
	GetSigner(address string) Signer
	GetVerifier(address string) Verifier
	Generate() string
}
```

#### Additional components

##### **Blob**

This is a wrapper for the widely used `[]byte` type that is used when handling binary data. Since crypto module handles sensitive information, the objective is to provide some extra security capabilities around such type as:

* Zeroing values after a read operation.
* Proper data handling.

These blob structures would be passed within components of the crypto module. For example: Signature information

#### **Keys**

A key object is responsible for containing the **BLOB** key information. Keys might not be passed through functions, and it is
suggested to interact through crypto providers to limit the exposure to vulnerabilities.

```mermaid
classDiagram
  PubKey <|-- PrivKey
  PubKey : Address() string
  PubKey : Key 

  PrivKey : PubKey() PubKey
  PrivKey : key
```

Base Key struct

```go
type KeyStruct struct {
 key Blob
}
```

Base key interface (common to private and public keys)

```go
type BaseKey interface {
 String() string
 Bytes() Blob
}
```

##### PubKey

```go
type PubKey interface {
 BaseKey
}
```

##### PrivKey

```go
type PrivKey interface {
 BaseKey
 Pubkey() PubKey //Generate a public key out of a private key
}
```

#### Signatures

A signature consists of a message/hash signed by one or multiple private keys. The main objective is to authenticate a message signer through their public key.

```go
type Signature struct {
 data Blob
}
```

**Flow overview**

***Initialization***

```mermaid
sequenceDiagram
    participant Application
    participant Wallet
    participant Keyring
    participant CryptoProvider
    participant SecureStorage

    Application->> Wallet: New()
    Wallet->>Keyring: New()

      loop CryptoProvider Registration
        CryptoProvider->>CryptoProvider: GetUUID()
        CryptoProvider->>CryptoProvider: GetBuilderFunction()
        CryptoProvider->>Keyring: RegisterCryptoProvider(UUID, BuilderFunction)
      end
      
      loop StorageSource Registration
        SecureStorage->>Keyring: RegisterStorageSource()
        Keyring->>SecureStorage: NewSecureStorage(SecureStorageSourceConfig)
        SecureStorage->>Keyring: SecureStorage Instance
        Keyring->>SecureStorage: List()
        SecureStorage->>Keyring: SecureItemMetadata list
      end
        Keyring->>Wallet: Keyring Instance

    Wallet->>Wallet: Init(Keyring)


    Wallet->>Application: Wallet Instance
```

***Signing and verifying a message***

```mermaid
sequenceDiagram
    participant Application
    participant Wallet
    participant Keyring
    participant CryptoProvider
    participant Signer
    participant Verifier

    Application->>Wallet: GetSigner(Address)
    Wallet->>Keyring: GetCryptoProvider(ItemId)
    Keyring->>SecureStorage: Get(ItemId.Uuid)
    SecureStorage->>Keyring: SecureItem
    Keyring->>Keyring: GetBuilderFunction(ItemId.Uuid)
    Keyring->>CryptoProvider: Build(SecureItem)
    CryptoProvider->>Wallet: CryptoProvider instance
    Wallet->>CryptoProvider: GetSigner()
    CryptoProvider->>Application: Signer instance
    Application->>Signer: Sign()
    Signer->>Application: Signed message
    Application->>Wallet: GetVerifier(address)
    Wallet->>CryptoProvider: GetVerifier()
    CryptoProvider->>Wallet: Verifier instance
    Application->>Verifier: Verify()
    Verifier->>Application: true/false
```

## Alternatives

The alternatives may vary in the way of distributing the packages, grouping them together as for example verify and signing in
one place. This will affect the granularity of the code, thus the reusability and modularity. We aim to balance between simplicity and
granularity.

## Decision

We will:

* Refactor module structure as described above.
* Define types and interfaces as the code attached.
* Refactor existing code into new structure and interfaces.
* Implement Unit Tests to ensure no backward compatibility issues.

## Consequences

### Backwards Compatibility

Some packages will need a medium to heavy refactor to be compatible with this ADR. 
In short, packages using _Keyring_ (current SDK) will need to be adapted to use the new Keyring and CryptoProvider interfaces.
Other special cases where a refactor will be needed, are the ones that make use crypto components in isolation like the  _PrivateKey_ and _PublicKey_ structs
to sign and verify transactions respectively.

As first approach, the most affected packages are:
- crypto/types
- client/rpc
- client/tx
- client/keys
- types/tx/signing
- x/auth
- x/auth/client
- x/slashing
- simapp/simd

### Positive

* Single place of truth
* Easier to use interfaces
* Easier to extend
* Unit test for each crypto package
* Greater maintainability
* Incentivize addition of implementations instead of forks
* Decoupling behaviour from implementation
* Sanitization of code

### Negative

* It will involve an effort to adapt existing code.
* It will require attention to detail and audition.

### Neutral

* It will involve extensive testing.

## Test Cases

*The code will be unit tested to ensure a high code coverage
- There should be integration tests around Wallet, keyring and crypto providers.
- There should be benchmark tests for hashing, keyring, encryption, decryption, signing and verifying functions.

## Further Discussions

> While an ADR is in the DRAFT or PROPOSED stage, this section should contain a
> summary of issues to be solved in future iterations (usually referencing comments
> from a pull-request discussion).
>
> Later, this section can optionally list ideas or improvements the author or
> reviewers found during the analysis of this ADR.


## References

* TPM 2.0 support: https://github.com/google/go-tpm
* Initial basic PKCS#11
  https://docs.oasis-open.org/pkcs11/pkcs11-base/v3.0/os/pkcs11-base-v3.0-os.pdf
  https://docs.aws.amazon.com/cloudhsm/latest/userguide/pkcs11-library.html
* https://docs.aws.amazon.com/cloudhsm/latest/userguide/pkcs11-key-types.html
* https://solanacookbook.com/references/keypairs-and-wallets.html#how-to-generate-a-new-keypair
* https://www.nist.gov/news-events/news/2022/07/nist-announces-first-four-quantum-resistant-cryptographic-algorithms
* https://blog.cloudflare.com/nist-post-quantum-surprise/
* https://pkg.go.dev/crypto#Hash
* https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html

## Appendix

### Tentative Primitive Building Blocks

This is a tentative list of primitives that we might want to support. 
This is not a final list or comprehensive, and it is subject to change. 
Moreover, it is important to emphasize the purpose of this work allows extensibility so any other primitive can be added in the future.

* digital signatures
    *  RSA (PSS)
    *  ECDSA (secp256r1, secp256k1, etc.)
    *  EdDSA (ed25519, ed448)
    *  SR25519
    *  Schnorr
    *  Lattice-based (Dilithium)
    *  BLS (BLS12-381, 377?)

* encryption
    * AES (AES-GCM, AES-CCM)
    * RSA (OAEP)
    * salsa20
    * (x)chacha20 / (x)ChaCha20-Poly1305 (AEAD)
    * Dilithium
    * Ntru

* Hashing
    * sha2 / sha3
    * RIPEMD-160  
    * blake2b,2s,3
    * Keccak-256 / shake256
    * bcrypt / scrypt / argon2, Argon2d/i/id
    * Pedersen
