# ADR 043: Interchain Identifiers

## Context

In the Internet of Blockchains We need a way to identify and locate assets within their chain namespace, using web-standard ([RFC39](https://datatracker.ietf.org/doc/html/rfc3986)) Universal Resource Indentifiers (URIs).

When we refer to on-chain assets, this includes tokens, wallets, name records, or any other uniquely identifiable entity within a chain namespace.

**Interchain Identifiers** (IIDs) are a standards-compliant method (based on [W3C DID-Core](https://w3c.github.io/did-core/)) for uniquely identifying, locating, referring to and verifying digital assets which are accessed within, or through, chain namespaces.

IIDs also enable (off-chain) assertions to be made about (on-chain) digital assets – for instance, in the form of Verifiable Credentials.

An IID Method DID is **self-referential**, which means the asset describes and locates itself in its chain namespace.

Compare this with a generic, all-purpose DID which can be used to identify and refer to _any_ entity. For instance, the canonical application of DIDs is Self-Sovereign Identity, which provides people and other entities with identifiers together with cryptographic methods to authenticate as the subject and to exercise control over the state of the DID Document.

---
We define **Interchain Identifiers** as being explicitly purposed to identify and locate assets within a chain namespace.

---

- An asset on a Cosmos chain <-------- Cosmos IID identifies and locates
- Cosmos DID --------> Identifies and locates an external thing

All DIDs have an associated **DID Document**, which should subscribe to the W3C DID-Core standard.
This defines the cryptographic control structure for a DID, and may also specify the Service types and endpoints which are related to the DID subject.

The properties of a DID-Core standard DID Document include:
- **Identifier** (the DID) of the subject with the Identifier of the Controller who is authorized to update the DID Document.
- **Verificaton Method/s** (The cryptographic key material).
- **Verification Relationship/s** (Describing and restricting the purposes for which the verification Methods can be used, such as Authentication).
- **Services** related to the DID subject, with endpoints for reaching these services, such as data stores or blockchain services.

For the IID Method, two additional DID Document properties have been specified, which are required for many types of on-chain assets:
- **Linked Resources** for describing the resources which are associated with an on-chain asset, such as a media file together with its resource identifier, type, format, location and cryptographic proof (such as a hash). 
- **Accorded Rights** for recording machine-executable capabilities, legal rights, or any other entitlements, which are accorded to the owner/controller of an asset.

Linked Resources can demonstrate the distinctions between an IID and DID:
```
Asset IID   // self-referrant <--
├── Linked Resources  // examples
│   ├── Resource DID  // externally referrant -->
│   ├── Other Resource URL
│   ├── Another Resource CID (Content ID)
```

The **DID Syntax** for an IID conforms to the [DID-Core Syntax](), which would be expressed as follows for an asset minted from the NFT Module on a Cosmos chain:

`DID:COSMOS:CHAIN:MODULE:abc123`

Where: 
- `COSMOS` is the method _(it is not the Cosmos Zone namespace)_.
- `CHAIN` is the namespace, such as `IXO` or `REGEN` _(it is not the chain-id)_.
- `NAMESPACE` is the asset namespace, such as `NFT` or `DOMAIN` _(as named within the context of each application chain)_.
- `{unique identifier string} abc123` is defined by applications developers _(NFTs in two different networkscould even have the identical Inique Identifier String value and will still have completely unique IIDs)_   

Unique Identifier strings need only be unique within the context of their own namespace. These may be Serial Numbers, Content Identifiers, or any another identifier method which is either based on a published standard, or is proprietary.

[DID Resolvers](https://www.w3.org/TR/did-core/#:~:text=7.1%20DID%20Resolution.-,DID,-resolver) are required for each DID Method, so that users are able to successfully retrieve the DID Document for a given DID (in the W3C DID-Core standard format).

It is not necessary for application chains to have a separate IID Module (or DID Module), if a specific asset module, such as `NFT`, provides the service of creating and updating IIDs and IID Document state. 

An example of the extensibility of the DID syntax: a derivative demom of fungible tokens could be minted to represent the shares  IIDs could   

## Decision

1. Cosmos IID Method DIDs SHOULD follow the standard format:
    -  `DID:COSMOS:` for the Cosmos Method (IID)
    -  `DID:COSMOS:CHAIN:` for a Cosmos chain namespace
    -  `DID:COSMOS:CHAIN:NAMESPACE:` for an asset namespace in a Cosmos application chain
    -  `DID:COSMOS:CHAIN:NAMESPACE:{unique identifier string}` for the application's chosen method for deriving a unique identifier string
2. IID implementations in the Cosmos SDK MUST conform to the Interchain Identifier (IID) specification RFC-09.
3. All Cosmos chains SHOULD provide a standard IID Interface for registering asset IIDs, creating and updating IID Document properties, and for resolving the IID to return an IID Document.
4. Cosmos assets MAY be identified using IIDs and if implementers choose this format, they SHOULD conform to the IID Method.
5. Application developers SHOULD define their own application-specific Unique Identifier String formats for the asset namespaces in their chain context.
6. Any chain registry service on the Cosmos Hub MUST support the Cosmos IID Method (did:cosmos).


## Status

Draft

## Consequences

### Positive

- Adoption of an interoperable Interchain standard for identifying and locating assets across all chain namespaces.
- IIDs may be used both within the context of a chain namespace, as well as off-chain, to reference and interact with an on-chain asset.
- Supports application developers to build more features and service capabilities into Cosmos modules, with advantages such as run-time composibility, fully decentralized authorization, more sophisticated account wallets, and service extensions.
- Conformance with W3C standards.
- Can immediately be extended by a family of DID-related  specifications, such as Verifiable Credentials and Authorization Capabilities.
- Immediately compatible with the extensive tooling (such as wallets) which has already been built for the DID ecosystem. 
- Cosmos application chains have the option to either adopt the IID specification to create their own IID DID Method, or or use did:cosmos to add DIDs to their application chain, without needing to formally specify a new DID Method.

### Neutral

- Minimal changes to existing Cosmos SDK modules (none of which are state-breaking).
- No changes to IBC.
- Any existing DID implementations within Cosmos which have used the DID:COSMOS syntax may not be backward-compatible with the IID Method (we don't know of any).
- Cosmos is already used by a number of application chains to provide DID Registry sevices for externally-referenced DIDs. The syntax of these DIDs do not have to change, but there would be advantages to having a standard way to resolve any DID which uses the IID Method.

### Negative
-  Legacy assets which have already been identified by some other method, will not be interoperable with new assets which use IIDs.

## References
- [#9337](https://github.com/cosmos/cosmos-sdk/discussions/9337)
- [Comment 885051](https://github.com/cosmos/cosmos-sdk/discussions/9065?sort=new#discussioncomment-885051)
- [RFC-09](https://github.com/interNFT/nft-rfc/blob/main/nft_rfc_009.md) Interchain Identifier Specification (Draft)
