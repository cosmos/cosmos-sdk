# Cosmos SDK Transaction Malleability Risks

One naive assumption about Cosmos SDK transactions is that hashing the raw bytes of a submitted transaction creates a safe unique identifier for the transaction. In reality, there are multiple ways in which transactions could be manipulated to create different transaction bytes (and as a result different hashes) that still pass signature verification.

This document attempts to enumerate the various potential transaction "malleability" risks that we have identified and the extent to which they have or have not been addressed in various sign modes. We also identify vulnerabilities that could be introduced if developers make changes in the future without careful consideration of the complexities involved with transaction encoding, sign modes and signatures.

## Risks Associated with Malleability

The malleability of transactions poses the following potential risks to end users:
* unsigned data could get added to transactions and be processed by state machines
* clients often rely on transaction hashes for checking transaction status, but whether or not submitted transaction hashes match processed transaction hashes depends primarily on good network actors rather than fundamental protocol guarantees
* transactions could potentially get executed more than once (faulty replay protection)

If a client generates a transaction, keeps a record of its hash and then attempts to query nodes to check the transaction's status, this process may falsely conclude that the transaction had not been processed if an intermediary
processor decoded and re-encoded the transaction with different encoding rules (either maliciously or unintentionally).
As long as no malleability is present the signature bytes themselves, clients SHOULD query transactions by signature instead of hash.

Not being cognizant of this risk may lead clients to submit the same transaction multiple times if they believe that 
earlier transactions had failed or gotten lost in processing, which is somewhat akin to faulty replay protection.
If the state machine were to rely on transaction hashes as a replay mechanism itself, this would be faulty and not 
provide the intended replay protection. Instead, the state machine should rely on deterministic representations of
transactions rather than raw encoding if they want to provide some replay protection that doesn't rely on a monotonically
increasing account sequence number.


## Sources of Malleability

### Non-deterministic Protobuf Encoding

Cosmos SDK transactions are encoded using protobuf binary encoding when they are submitted to the network. Protobuf binary is not inherently a deterministic encoding meaning that the same logical payload could have several valid bytes representations. In a basic sense, this means that protobuf in general can be decoded and re-encoded to produce a different byte stream (and thus different hash) without changing the logical meaning of the bytes. [ADR 027: Deterministic Protobuf Serialization](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-027-deterministic-protobuf-serialization.md) describes in detail what needs to be done to produce what we consider to be a "canonical", deterministic protobuf serialization. Briefly, the following sources of malleability at the encoding level have been identified and are addressed by this specification:
* fields can be emitted in any order
* default field values can be included or omitted, and this doesn't change meaning unless `optional` is used
* `repeated` fields of scalars may use packed or "regular" encoding
* `varint`s can include extra ignored bits
* extra fields may be added and are usually simply ignored by decoders. [ADR 020](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-020-protobuf-transaction-encoding.md#unknown-field-filtering) specifies that in general such extra fields should cause messages and transactions to be rejected)

When using `SIGN_MODE_DIRECT` most of the above malleabilities will not be tolerated because:
* signatures of messages and extensions must be done over the raw encoded bytes of those fields
* the outer tx envelope (`TxRaw`) must follow ADR 027 rules or be rejected

Transactions signed with `SIGN_MODE_LEGACY_AMINO_JSON`, however, have no way of protecting against the above malleabilities because what is signed is a JSON representation of the logical contents of the transaction. These logical contents could have any number of valid protobuf binary encodings, so in general there are no guarantees regarding transaction hash with Amino JSON signing.

In addition to being aware of the general non-determinism of protobuf binary, developers need to pay special attention to make sure that unknown protobuf fields get rejected when developing new capabilities related to protobuf transactions. The protobuf serialization format was designed with the assumption that unknown data known to encoders could safely be ignored by decoders. This assumption may have been fairly safe within the walled garden of Google's centralized infrastructure. However, in distributed blockchain systems, this assumption is generally unsafe. If a newer client encodes a protobuf message with data intended for a newer server, it is not safe for an older server to simply ignore and discard instructions that it does not understand. These instructions could include critical information that the transaction signer is relying upon and just assuming that it is unimportant is not safe.

[ADR 020](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-020-protobuf-transaction-encoding.md#unknown-field-filtering) specifies some provisions for "non-critical" fields which can safely be ignored by older servers. In practice, I have not seen any valid usages of this. It is something in the design that maintainers should be aware of, but it may not be necessary or even 100% safe.

### Non-deterministic Value Encoding (Signatures, Numbers, etc.)

In addition to the non-determinism present in protobuf binary itself, some protobuf field data is encoded using a micro-format which itself may not be deterministic. Consider for instance integer or decimal encoding. Some decoders may allow for the presence of leading or trailing zeros without changing the logical meaning, ex. `00100` vs `100` or `100.00` vs `100`. So if a sign mode encodes numbers deterministically, but decoders accept multiple representations,
a user may sign over the value `100` while `0100` gets encoded. This would be possible with Amino JSON to the extent that the integer decoder accepts leading zeros. I believe the current `Int` implementation will reject this, however, it is
probably possible to encode a octal or hexadecimal representation in the transaction whereas the user signs over a decimal integer.

Signatures themselves are encoded using a micro-format specific to the signature algorithm being used and sometimes these
micro-formats can allow for non-determinism (multiple valid bytes for the same signature).
I believe we have made sure that the algorithms used in the SDK reject non-canonical signature bytes.
Any new or custom signature algorithms must make sure that they reject any non-canonical bytes, otherwise even
with `SIGN_MODE_DIRECT` there can be transaction hash malleability by re-encoding signatures with a non-canonical
representation.

### Fields not covered by Amino JSON

Another area that needs to be addressed carefully is the discrepancy between `StdSignDoc` used for `SIGN_MODE_LEGACY_AMINO_JSON` and the actual contents of `TxBody` and `AuthInfo`. If fields get added
to `TxBody` or `AuthInfo`, they must either have a corresponding representing in `StdSignDoc` or Amino
JSON signatures must be rejected when those new fields are set. Making sure that this is done is a
highly manual process, and developers could easily make the mistake of updating `TxBody` or `AuthInfo`
without paying any attention to the implementation of `GetSignBytes` for Amino JSON. This is a critical
vulnerability in which unsigned content can now get into the transaction and signature verification will
pass.