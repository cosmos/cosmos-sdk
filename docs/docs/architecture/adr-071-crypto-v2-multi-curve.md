# ADR 071: Cryptography v2- Multi-curve support

## Change log

* May 7th 2024: Initial Draft (Zondax AG: @raynaudoe @juliantoledano @jleni @educlerici-zondax @lucaslopezf)
* June 13th 2024: Add CometBFT implementation proposal (Zondax AG: @raynaudoe @juliantoledano @jleni @educlerici-zondax @lucaslopezf)
* July 2nd 2024: Split ADR proposal, add link to ADR in cosmos/crypto (Zondax AG: @raynaudoe @juliantoledano @jleni @educlerici-zondax @lucaslopezf)

## Status

DRAFT

## Abstract

This ADR proposes the refactoring of the existing `Keyring` and `cosmos-sdk/crypto` code to implement [ADR-001-CryptoProviders](https://github.com/cosmos/crypto/blob/main/docs/architecture/adr-001-crypto-provider.md).

For in-depth details of the `CryptoProviders` and their design please refer to ADR mentioned above.

## Introduction

The introduction of multi-curve support in the cosmos-sdk cryptographic package offers significant advantages. By not being restricted to a single cryptographic curve, developers can choose the most appropriate curve based on security, performance, and compatibility requirements. This flexibility enhances the application's ability to adapt to evolving security standards and optimizes performance for specific use cases, helping to future-proofing the sdk's cryptographic capabilities.


The enhancements in this proposal not only render the ["Keyring ADR"](https://github.com/cosmos/cosmos-sdk/issues/14940) obsolete, but also encompass its key aspects, replacing it with a more flexible and comprehensive approach. Furthermore, the gRPC service proposed in the mentioned ADR can be easily implemented as a specialized `CryptoProvider`. 


### Glossary

1. **Interface**: In the context of this document, "interface" refers to Go's interface.

2. **Module**: In this document, "module" refers to a Go module.

3. **Package**: In the context of Go, a "package" refers to a unit of code organization.

## Context

In order to fully understand the need for changes and the proposed improvements, it's crucial to consider the current state of affairs:

* The Cosmos SDK currently lacks a comprehensive ADR for the cryptographic package.

* If a blockchain project requires a cryptographic curve that is not supported by the current SDK, the most likely scenario is that they will need to fork the SDK repository and make modifications. These modifications could potentially make the fork incompatible with future updates from the upstream SDK, complicating maintenance and integration.

* Type leakage of specific crypto data types expose backward compatibility and extensibility challenges.

* The demand for a more flexible and extensible approach to cryptography and address management is high.

* Architectural changes are necessary to resolve many of the currently open issues related to new curves support.

* There is a current trend towards modularity in the Interchain stack (e.g., runtime modules).

* Security implications are a critical consideration during the redesign work.

## Objectives

The key objectives for this proposal are:

* Leverage `CryptoProviders`: Utilize them as APIs for cryptographic tools, ensuring modularity, flexibility, and ease of integration.

Developer-Centric Approach

* Prioritize clear, intuitive interfaces and best-practice design principles.

Quality Assurance

* Enhanced Test Coverage: Improve testing methodologies to ensure the robustness and reliability of the module.

## Technical Goals

New Keyring:

* Design a new `Keyring` interface with modular backends injection system to support hardware devices and cloud-based HSMs. This feature is optional and tied to complexity; if it proves too complex, it will be deferred to a future release as an enhancement.


## Proposed architecture


### Components

The main components to be used will be the same as those found in the [ADR-001](https://github.com/cosmos/crypto/blob/main/docs/architecture/adr-001-crypto-provider.md#components).


#### Storage and persistence

The storage and persistence layer is tasked with storing a `CryptoProvider`s. Specifically, this layer must:

* Securely store the crypto provider's associated private key (only if stored locally, otherwise a reference to the private key will be stored instead).
* Store the [`ProviderMetadata`](https://github.com/cosmos/crypto/blob/main/docs/architecture/adr-001-crypto-provider.md#metadata) struct which contains the data that distinguishes that provider.

The purpose of this layer is to ensure that upon retrieval of the persisted data, we can access the provider's type, version, and specific configuration (which varies based on the provider type). This information will subsequently be utilized to initialize the appropriate factory, as detailed in the following section on the factory pattern.

The storage proposal involves using a modified version of the [Record](https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/crypto/keyring/v1/record.proto) struct, which is already defined in **Keyring/v1**. Additionally, we propose utilizing the existing keyring backends (keychain, filesystem, memory, etc.) to store these `Record`s in the same manner as the current **Keyring/v1**.

*Note: This approach will facilitate a smoother migration path from the current Keyring/v1 to the proposed architecture.*

Below is the proposed protobuf message to be included in the modified `Record.proto` file

##### Protobuf message structure


The [record.proto](https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/crypto/keyring/v1/record.proto) file will be modified to include the `CryptoProvider` message as an optional field as follows.

```protobuf

// record.proto

message Record {
  string name = 1;
  google.protobuf.Any pub_key = 2;

  oneof item {
    Local local = 3;
    Ledger ledger = 4;
    Multi multi = 5;
    Offline offline = 6;
    CryptoProvider crypto_provider = 7; // <- New
  }

  message Local {
    google.protobuf.Any priv_key = 1;
  }

  message Ledger {
    hd.v1.BIP44Params path = 1;
  }

  message Multi {}

  message Offline {}
}
```

##### Creating and loading a `CryptoProvider`

For creating providers, we propose a *factory pattern* and a *registry* for these builders. Examples of these
patterns can be found [here](https://github.com/cosmos/crypto/blob/main/docs/architecture/adr-001-crypto-provider.md#illustrative-code-snippets)


##### Keyring

The new `Keyring` interface will serve as a central hub for managing and fetching `CryptoProviders`. To ensure a smoother migration path, the new Keyring will be backward compatible with the previous version. Since this will be the main API from which applications will obtain their `CryptoProvider` instances, the proposal is to extend the Keyring interface to include the methods:


```go
type KeyringV2 interface {
  // methods from Keyring/v1
  
  // ListCryptoProviders returns a list of all the stored CryptoProvider metadata.
  ListCryptoProviders() ([]ProviderMetadata, error)
  
  // GetCryptoProvider retrieves a specific CryptoProvider by its id.
  GetCryptoProvider(id string) (CryptoProvider, error)
}
```

*Note*: Methods to obtain a provider from a public key or other means that make it easier to load the desired provider can be added.

##### Especial use case: remote signers

It's important to note that the `CryptoProvider` interface is versatile enough to be implemented as a remote signer. This capability allows for the integration of remote cryptographic operations, which can be particularly useful in distributed or cloud-based environments where local cryptographic resources are limited or need to be managed centrally.


## Alternatives

It is important to note that all the code presented in this document is not in its final form and could be subject to changes at the time of implementation. The examples and implementations discussed should be interpreted as alternatives, providing a conceptual framework rather than definitive solutions. This flexibility allows for adjustments based on further insights, technical evaluations, or changing requirements as development progresses.

## Decision

We will:

* Leverage crypto providers
* Refactor the module structure as described above.
* Define types and interfaces as the code attached.
* Refactor existing code into new structure and interfaces.
* Implement Unit Tests to ensure no backward compatibility issues.

## Consequences

### Impact on the SDK codebase

We can divide the impact of this ADR into two main categories: state machine code and client related code.

#### Client

The major impact will be on the client side, where the current `Keyring` interface will be replaced by the new `KeyringV2` interface. At first, the impact will be low since `CryptoProvider` is an optional field in the `Record` message, so there's no mandatory requirement for migrating to this new concept right away. This allows a progressive transition where the risks of breaking changes or regressions are minimized.


#### State Machine

The impact on the state machine code will be minimal, the modules affected (at the time of writing this ADR)
are the `x/accounts` module, specifically the `Authenticate` function and the `x/auth/ante` module. This function will need to be adapted to use a `CryptoProvider` service to make use of the `Verifier` instance.

Worth mentioning that there's also the alternative of using `Verifier` instances in a standalone fashion (see note below).

The specific way to adapt these modules will be deeply analyzed and decided at implementation time of this ADR.


*Note*: All cryptographic tools (hashers, verifiers, signers, etc.) will continue to be available as standalone packages that can be imported and utilized directly without the need for a `CryptoProvider` instance. However, the `CryptoProvider` is the recommended method for using these tools as it offers a more secure way to handle sensitive data, enhanced modularity, and the ability to store configurations and metadata within the `CryptoProvider` definition.


### Backwards Compatibility

The proposed migration path is similar to what the cosmos-sdk has done in the past. To ensure a smooth transition, the following steps will be taken:

Once ADR-001 is implemented with a stable release:

* Deprecate the old crypto package. The old crypto package will still be usable, but it will be marked as deprecated and users can opt to use the new package.
* Migrate the codebase to use the new cosmos/crypto package and remove the old crypto one.

### Positive

* Single place of truth
* Easier to use interfaces
* Easier to extend
* Unit test for each crypto package
* Greater maintainability
* Incentivize addition of implementations instead of forks
* Decoupling behavior from implementation
* Sanitization of code

### Negative

* It will involve an effort to adapt existing code.
* It will require attention to detail and audition.

### Neutral

* It will involve extensive testing.

## Test Cases

* The code will be unit tested to ensure a high code coverage
* There should be integration tests around Keyring and CryptoProviders.

> While an ADR is in the DRAFT or PROPOSED stage, this section should contain a
> summary of issues to be solved in future iterations (usually referencing comments
> from a pull-request discussion).
>
> Later, this section can optionally list ideas or improvements the author or
> reviewers found during the analysis of this ADR.
