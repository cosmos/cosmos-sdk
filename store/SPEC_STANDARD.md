## What is a store specification document?

A store specification document is a design document describing a particular store feature such as a store interface or implementation that it is expected to be used (or already in use) by the Cosmos SDK. This also includes other store-related features such as caches and wrappers.

## Sections

An store specification document consists of, synopsis, overview and basic concepts, technical specification, history log, and copyright notice. All top-level sections are required. References should be included inline as links, or tabulated at the bottom of the section if necessary.  Included sub-sections should be listed in the order specified below. 

### Synopsis

The document should include a brief (~200 word) synopsis providing a high-level
description of and rationale for the specification.

### Overview and basic concepts

This section should include a motivation sub-section and a definitions sub-section if required:

- *Motivation* - A rationale for the existence of the proposed interface, implementation, or the proposed changes to an existing feature.
- *Definitions* - A list of new terms or concepts utilised in the document or required to understand it.

### System model and properties

This section should include an assumptions sub-section if any, and the mandatory properties sub-section. Note that both sub-sections are tightly coupled: how to enforce a property will depend directly on the assumptions made.

- *Assumptions* - A list of any assumptions made by the feature designer. This may for instance include how it is assumed that the feature will be used by upper layers in the stack.
- *Properties* - The content of this sub-section depends on the type of specification. For implementations, the sub-section should include a list of properties that the implementation guarantees and list which properties it does not. The latter is as important as the former. In the case of a store interface, this sub-section should include a list of the desired properties that an implementation of the interface should guarantee. Ideally, along with each of the properties should go a text arguing why the property is required. Examples of properties are whether the feature is or must be thread-safe, atomicity-related properties or behaviour under failures.

### Technical specification

This is the main section of the document, and should contain protocol documentation, design rationale,
required references, and technical details where appropriate. 
Apart from the API sub-section which is required, the section may have any or all of the following sub-sections, as appropriate to the particular specification.

- *API* - A detailed description of the features's API including which interfaces implements (for implementations) or extends (for interfaces).
- *Technical Specification* - All technical details including syntax, diagrams, semantics, protocols, data structures, algorithms, and pseudocode as appropriate. The technical specification should be detailed enough such that separate correct implementations of the specification without knowledge of each other are compatible.
- *Backwards Compatibility* - A discussion of compatibility (or lack thereof) with previous feature or protocol versions.
- *Known Issues* - A list of known issues. This sub-section is specially important for specifications of already in-use features.
- *Example Implementation* - A concrete example implementation or description of an expected implementation to serve as the primary reference for implementers. This sub-section is only expected to be included in newly proposed interfaces.

### History

A store specification should include a history section, listing any inspiring documents and a plaintext log of significant changes.

See an example history section [below](#history-1).

### Copyright

A store specification should include a copyright section waiving rights via [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## Formatting

### General

Specifications must be written in GitHub-flavoured Markdown.

For a GitHub-flavoured Markdown cheat sheet, see [here](https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet). For a local Markdown renderer, see [here](https://github.com/joeyespo/grip).

### Language

Specifications should be written in Simple English, avoiding obscure terminology and unnecessary jargon. For excellent examples of Simple English, please see the [Simple English Wikipedia](https://simple.wikipedia.org/wiki/Main_Page).

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in specifications are to be interpreted as described in [RFC 2119](https://tools.ietf.org/html/rfc2119).

### Pseudocode

Pseudocode in specifications should be language-agnostic and formatted in a simple imperative standard, with line numbers, variables, simple conditional blocks, for loops, and
English fragments where necessary to explain further functionality such as scheduling timeouts. LaTeX images should be avoided because they are difficult to review in diff form.

Pseudocode for structs should be written in simple Typescript, as interfaces.

Example pseudocode struct:

```typescript
interface CacheKVStore {
  cache: Map<Key, Value>
  parent: KVStore
  deleted: Key
}
```

Pseudocode for algorithms should be written in simple Typescript, as functions.

Example pseudocode algorithm:

```typescript
function get(
  store: CacheKVStore,
  key: Key): Value {

  value = store.cache.get(Key)
  if (value !== null) {
    return value
  } else {
    value = store.parent.get(key)
    store.cache.set(key, value)
    return value
  }
}
```

## History

This specification was significantly inspired by and derived from IBC's [ICS](https://github.com/cosmos/ibc/blob/main/spec/ics-001-ics-standard/README.md), which
was in turn derived from Ethereum's [EIP 1](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1.md).

Nov 24, 2022 - Initial draft finished and submitted as a PR

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).