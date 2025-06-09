# Cosmos SDK Transaction Malleability Risk Review and Recommendations

## Changelog

* 2025-03-10: Initial draft (@aaronc)

## Status

PROPOSED: Not Implemented

## Abstract

Several encoding and sign mode related issues have historically resulted in the possibility
that Cosmos SDK transactions may be re-encoded in such a way as to change their hash
(and in rare cases, their meaning) without invalidating the signature.
This document details these cases, their potential risks, the extent to which they have been
addressed, and provides recommendations for future improvements.

## Review

One naive assumption about Cosmos SDK transactions is that hashing the raw bytes of a submitted transaction creates a safe unique identifier for the transaction. In reality, there are multiple ways in which transactions could be manipulated to create different transaction bytes (and as a result different hashes) that still pass signature verification.

This document attempts to enumerate the various potential transaction "malleability" risks that we have identified and the extent to which they have or have not been addressed in various sign modes. We also identify vulnerabilities that could be introduced if developers make changes in the future without careful consideration of the complexities involved with transaction encoding, sign modes and signatures.

### Risks Associated with Malleability

The malleability of transactions poses the following potential risks to end users:

* unsigned data could get added to transactions and be processed by state machines
* clients often rely on transaction hashes for checking transaction status, but whether or not submitted transaction hashes match processed transaction hashes depends primarily on good network actors rather than fundamental protocol guarantees
* transactions could potentially get executed more than once (faulty replay protection)

If a client generates a transaction, keeps a record of its hash and then attempts to query nodes to check the transaction's status, this process may falsely conclude that the transaction had not been processed if an intermediary
processor decoded and re-encoded the transaction with different encoding rules (either maliciously or unintentionally).
As long as no malleability is present in the signature bytes themselves, clients _should_ query transactions by signature instead of hash.

Not being cognizant of this risk may lead clients to submit the same transaction multiple times if they believe that 
earlier transactions had failed or gotten lost in processing.
This could be an attack vector against users if wallets primarily query transactions by hash.

If the state machine were to rely on transaction hashes as a replay mechanism itself, this would be faulty and not 
provide the intended replay protection. Instead, the state machine should rely on deterministic representations of
transactions rather than the raw encoding, or other nonces,
if they want to provide some replay protection that doesn't rely on a monotonically
increasing account sequence number.


### Sources of Malleability

#### Non-deterministic Protobuf Encoding

Cosmos SDK transactions are encoded using protobuf binary encoding when they are submitted to the network. Protobuf binary is not inherently a deterministic encoding meaning that the same logical payload could have several valid bytes representations. In a basic sense, this means that protobuf in general can be decoded and re-encoded to produce a different byte stream (and thus different hash) without changing the logical meaning of the bytes. [ADR 027: Deterministic Protobuf Serialization](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-027-deterministic-protobuf-serialization.md) describes in detail what needs to be done to produce what we consider to be a "canonical", deterministic protobuf serialization. Briefly, the following sources of malleability at the encoding level have been identified and are addressed by this specification:

* fields can be emitted in any order
* default field values can be included or omitted, and this doesn't change meaning unless `optional` is used
* `repeated` fields of scalars may use packed or "regular" encoding
* `varint`s can include extra ignored bits
* extra fields may be added and are usually simply ignored by decoders. [ADR 020](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-020-protobuf-transaction-encoding.md#unknown-field-filtering) specifies that in general such extra fields should cause messages and transactions to be rejected)

When using `SIGN_MODE_DIRECT` none of the above malleabilities will be tolerated because:

* signatures of messages and extensions must be done over the raw encoded bytes of those fields
* the outer tx envelope (`TxRaw`) must follow ADR 027 rules or be rejected

Transactions signed with `SIGN_MODE_LEGACY_AMINO_JSON`, however, have no way of protecting against the above malleabilities because what is signed is a JSON representation of the logical contents of the transaction. These logical contents could have any number of valid protobuf binary encodings, so in general there are no guarantees regarding transaction hash with Amino JSON signing.

In addition to being aware of the general non-determinism of protobuf binary, developers need to pay special attention to make sure that unknown protobuf fields get rejected when developing new capabilities related to protobuf transactions. The protobuf serialization format was designed with the assumption that unknown data known to encoders could safely be ignored by decoders. This assumption may have been fairly safe within the walled garden of Google's centralized infrastructure. However, in distributed blockchain systems, this assumption is generally unsafe. If a newer client encodes a protobuf message with data intended for a newer server, it is not safe for an older server to simply ignore and discard instructions that it does not understand. These instructions could include critical information that the transaction signer is relying upon and just assuming that it is unimportant is not safe.

[ADR 020](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-020-protobuf-transaction-encoding.md#unknown-field-filtering) specifies some provisions for "non-critical" fields which can safely be ignored by older servers. In practice, I have not seen any valid usages of this. It is something in the design that maintainers should be aware of, but it may not be necessary or even 100% safe.

#### Non-deterministic Value Encoding

In addition to the non-determinism present in protobuf binary itself, some protobuf field data is encoded using a micro-format which itself may not be deterministic. Consider for instance integer or decimal encoding. Some decoders may allow for the presence of leading or trailing zeros without changing the logical meaning, ex. `00100` vs `100` or `100.00` vs `100`. So if a sign mode encodes numbers deterministically, but decoders accept multiple representations,
a user may sign over the value `100` while `0100` gets encoded. This would be possible with Amino JSON to the extent that the integer decoder accepts leading zeros. I believe the current `Int` implementation will reject this, however, it is
probably possible to encode a octal or hexadecimal representation in the transaction whereas the user signs over a decimal integer.

#### Signature Encoding

Signatures themselves are encoded using a micro-format specific to the signature algorithm being used and sometimes these
micro-formats can allow for non-determinism (multiple valid bytes for the same signature).
Most of the signature algorithms supported by the SDK should reject non-canonical bytes in their current implementation.
However, the `Multisignature` protobuf type uses normal protobuf encoding and there is no check as to whether the
decoded bytes followed canonical ADR 027 rules or not. Therefore, multisig transactions can have malleability in
their signatures.
Any new or custom signature algorithms must make sure that they reject any non-canonical bytes, otherwise even
with `SIGN_MODE_DIRECT` there can be transaction hash malleability by re-encoding signatures with a non-canonical
representation.

#### Fields not covered by Amino JSON

Another area that needs to be addressed carefully is the discrepancy between `AminoSignDoc`(see [`aminojson.proto`](../../x/tx/signing/aminojson/internal/aminojsonpb/aminojson.proto)) used for `SIGN_MODE_LEGACY_AMINO_JSON` and the actual contents of `TxBody` and `AuthInfo` (see [`tx.proto`](../../proto/cosmos/tx/v1beta1/tx.proto)).
If fields get added to `TxBody` or `AuthInfo`, they must either have a corresponding representing in `AminoSignDoc` or Amino JSON signatures must be rejected when those new fields are set. Making sure that this is done is a
highly manual process, and developers could easily make the mistake of updating `TxBody` or `AuthInfo`
without paying any attention to the implementation of `GetSignBytes` for Amino JSON. This is a critical
vulnerability in which unsigned content can now get into the transaction and signature verification will
pass.

## Sign Mode Summary and Recommendations

The sign modes officially supported by the SDK are `SIGN_MODE_DIRECT`, `SIGN_MODE_TEXTUAL`, `SIGN_MODE_DIRECT_AUX`,
and `SIGN_MODE_LEGACY_AMINO_JSON`.
`SIGN_MODE_LEGACY_AMINO_JSON` is used commonly by wallets and is currently the only sign mode supported on Nano Ledger hardware devices
(although `SIGN_MODE_TEXTUAL` was designed to also support hardware devices).
`SIGN_MODE_DIRECT` is the simplest sign mode and its usage is also fairly common.
`SIGN_MODE_DIRECT_AUX` is a variant of `SIGN_MODE_DIRECT` that can be used by auxiliary signers in a multi-signer
transaction by those signers who are not paying gas.
`SIGN_MODE_TEXTUAL` was intended as a replacement for `SIGN_MODE_LEGACY_AMINO_JSON`, but as far as we know it
has not been adopted by any clients yet and thus is not in active use.

All known malleability concerns have been addressed in the current implementation of `SIGN_MODE_DIRECT`.
The only known malleability that could occur with a transaction signed with `SIGN_MODE_DIRECT` would
need to be in the signature bytes themselves.
Since signatures are not signed over, it is impossible for any sign mode to address this directly
and instead signature algorithms need to take care to reject any non-canonically encoded signature bytes
to prevent malleability.
For the known malleability of the `Multisignature` type, we should make sure that any valid signatures
were encoded following canonical ADR 027 rules when doing signature verification.

`SIGN_MODE_DIRECT_AUX` provides the same level of safety as `SIGN_MODE_DIRECT` because

* the raw encoded `TxBody` bytes are signed over in `SignDocDirectAux`, and
* a transaction using `SIGN_MODE_DIRECT_AUX` still requires the primary signer to sign the transaction with `SIGN_MODE_DIRECT`

`SIGN_MODE_TEXTUAL` also provides the same level of safety as `SIGN_MODE_DIRECT` because the hash of the raw encoded
`TxBody` and `AuthInfo` bytes are signed over.

Unfortunately, the vast majority of unaddressed malleability risks affect `SIGN_MODE_LEGACY_AMINO_JSON` and this
sign mode is still commonly used.
It is recommended that the following improvements be made to Amino JSON signing:

* hashes of `TxBody` and `AuthInfo` should be added to `AminoSignDoc` so that encoding-level malleablity is addressed
* when constructing `AminoSignDoc`, [protoreflect](https://pkg.go.dev/google.golang.org/protobuf/reflect/protoreflect) API should be used to ensure that there no fields in `TxBody` or `AuthInfo` which do not have a mapping in `AminoSignDoc` have been set
* fields present in `TxBody` or `AuthInfo` that are not present in `AminoSignDoc` (such as extension options) should
be added to `AminoSignDoc` if possible

## Testing

To test that transactions are resistant to malleability,
we can develop a test suite to run against all sign modes that
attempts to manipulate transaction bytes in the following ways:

* changing protobuf encoding by
    * reordering fields
    * setting default values
    * adding extra bits to varints, or
    * setting new unknown fields
* modifying integer and decimal values encoded as strings with leading or trailing zeros

Whenever any of these manipulations is done, we should observe that the sign doc bytes for the sign mode being
tested also change, meaning that the corresponding signatures will also have to change.

In the case of Amino JSON, we should also develop tests which ensure that if any `TxBody` or `AuthInfo`
field not supported by Amino's `AminoSignDoc` is set that signing fails.

In the general case of transaction decoding, we should have unit tests to ensure that

* any `TxRaw` bytes which do not follow ADR 027 canonical encoding cause decoding to fail, and
* any top-level transaction elements including `TxBody`, `AuthInfo`, public keys, and messages which
have unknown fields set cause the transaction to be rejected
(this ensures that ADR 020 unknown field filtering is properly applied)

For each supported signature algorithm,
there should also be unit tests to ensure that signatures must be encoded canonically
or get rejected.

## References

* [ADR 027: Deterministic Protobuf Serialization](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-027-deterministic-protobuf-serialization.md)
* [ADR 020](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-020-protobuf-transaction-encoding.md#unknown-field-filtering)
* [`aminojson.proto`](../../x/tx/signing/aminojson/internal/aminojsonpb/aminojson.proto)
* [`tx.proto`](../../proto/cosmos/tx/v1beta1/tx.proto)

