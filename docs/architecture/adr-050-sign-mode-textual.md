# ADR 050: SIGN_MODE_TEXTUAL

## Changelog

- Dec 06, 2021: Initial Draft.
- Feb 07, 2022: Draft read and concept-ACKed by the Ledger team.
- May 16, 2022: Change status to Accepted.
- Aug 11, 2022: Require signing over tx raw bytes.
- Sep 07, 2022: Add custom `Msg`-renderers.

## Status

Accepted. Implementation started. Small value renderers details still need to be polished.

## Abstract

This ADR specifies SIGN_MODE_TEXTUAL, a new string-based sign mode that is targetted at signing with hardware devices.

## Context

Protobuf-based SIGN_MODE_DIRECT was introduced in [ADR-020](./adr-020-protobuf-transaction-encoding.md) and is intended to replace SIGN_MODE_LEGACY_AMINO_JSON in most situations, such as mobile wallets and CLI keyrings. However, the [Ledger](https://www.ledger.com/) hardware wallet is still using SIGN_MODE_LEGACY_AMINO_JSON for displaying the sign bytes to the user. Hardware wallets cannot transition to SIGN_MODE_DIRECT as:

- SIGN_MODE_DIRECT is binary-based and thus not suitable for display to end-users. Technically, hardware wallets could simply display the sign bytes to the user. But this would be considered as blind signing, and is a security concern.
- hardware cannot decode the protobuf sign bytes due to memory constraints, as the Protobuf definitions would need to be embedded on the hardware device.

In an effort to remove Amino from the SDK, a new sign mode needs to be created for hardware devices. [Initial discussions](https://github.com/cosmos/cosmos-sdk/issues/6513) propose a string-based sign mode, which this ADR formally specifies.

## Decision

We propose to have SIGN_MODE_TEXTUAL’s signing payload `SignDocTextual` to be an array of strings, encoded as a `\n`-delimited string (see point #9). Each string corresponds to one "screen" on the hardware wallet device, with no (or little) additional formatting done by the hardware wallet itself.

```proto
message SignDocTextual {
  repeated string screens = 1;
}
```

The string array MUST follow the specifications below.

### 1. Bijectivity with Protobuf transactions

The encoding and decoding operations between a Protobuf transaction (whose definition can be found [here](https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/tx/v1beta1/tx.proto#L13)) and the string array MUST be bijective.

We concede that bijectivity is not strictly needed. Avoiding transaction malleability only requires collision resistance on the encoding. Lossless encoding also does not require decodability. However, bijectivity assures both non-malleability and losslessness.

Bijectivity will be tested in two ways:

- by providing a set of test fixtures between a transaction's Proto JSON representation and its TEXTUAL representation, and checking that encoding/decoding in both directions matches the fixtures,
- by using property testing on the proto transaction itself, and testing that the composition of encoding and decoding yields the original transaction itself.

This also prevents users signing over any hashed transaction data (fee, transaction body, `Msg` content that might be hashed etc), which is a security concern for Ledger's security team.

We propose to maintain functional tests using bijectivity in the SDK.

### 2. Only ASCII 32-127 characters allowed

Ledger devices have limited character display capabilities, so all strings MUST only contain ASCII characters in the 32-127 range.

In particular, the newline `"\n"` (ASCII: 10) character is forbidden.

### 3. Strings SHOULD have the `<key>: <value>` format

Given the Regex `/^(\* )?(>* )?(.*: )?(.*)$/`, all strings SHOULD match the Regex with capture groups 3 and 4 non-empty. This is helpful for UIs displaying SignDocTextual to users.

- The initial `*` character is optional and denotes the Ledger Expert mode, see #5.
- Strings can also include a number of `>` character to denote nesting.
- In the case where the first Regex capture group is not empty, it represents an indicative key, whose associated value is given in the second capture group. This MAY be used in the Ledger app to perform custom on-screen formatting, for example to break long lines into multiple screens.

This Regex is however not mandatory, to allow for some flexibility, for example to display an English sentence to denote end of sections.

The `<value>` itself can contain the `": "` characters.

### 4. Values are encoded using Value Renderers

Value Renderers describe how Protobuf types are encoded to and decoded from a string array. The full specification of Value Renderers can be found in [Annex 1](./adr-050-sign-mode-textual-annex1.md).

### 5. Strings starting with `*` are only shown in Expert mode

Ledger devices have the an Expert mode for advanced users. Expert mode needs to be manually activated by the device holder, inside the device settings. There is currently no official documentation on Expert mode, but according to the [@Ledger_Support twitter account](https://twitter.com/Ledger_Support/status/1364524431800950785),

> Expert mode enables further, more sophisticated features. This could be useful for advanced users

Strings starting with the `*` character will only be shown in Expert mode. These strings are either hardcoded in the transaction envelope (see point #7).

For hardware wallets that don't have an expert mode, all strings MUST be shown on the device.

### 6. Strings MAY contain `>` characters to denote nesting

Protobuf objects can be arbitrarily complex, containing nested arrays and messages. In order to help the Ledger-signing users, we propose to use the `>` symbol in the beginning of strings to represent nested objects, where each additional `>` represents a new level of nesting.

### 7. Encoding of the Transaction Envelope

We define "transaction envelope" as all data in a transaction that is not in the `TxBody.Messages` field. Transaction envelope includes fee, signer infos and memo, but don't include `Msg`s. `//` denotes comments and are not shown on the Ledger device.

```
Chain ID: <string>
Account number: <uint64>
*Public Key: <hex_string>
Sequence: <uint64>
<TxBody>                                                    // See #8.
Fee: <coins>                                                // See value renderers for coins encoding.
*Fee payer: <string>                                        // Skipped if no fee_payer set
*Fee granter: <string>                                      // Skipped if no fee_granter set
Memo: <string>                                              // Skipped if no memo set
*Gas Limit: <uint64>
*Timeout Height:  <uint64>                                  // Skipped if no timeout_height set
Tipper: <string>                                            // If there's a tip
Tip: <string>
*This transaction has <int> body extension:                 // Skipped if no body extension options
*<repeated Any>
*This transaction has <int> body non-critical extensions:   // Skipped if no body non-critical extension options
*<repeated Any>                                             // See value renderers for Any and array encoding.
*This transaction has <int> body auth info extensions:      // Skipped if no auth info extension options
*<repeated Any>
*This transaction has <int> other signers:                  // Skipped if there is only one signer
*Signer (<int>/<int>):
*Public Key: <hex_string>
*Sequence: <uint64>
*End of other signers
*Hash of raw bytes: <hex_string>                            // Hex encoding of bytes defined in #10, to prevent tx hash malleability.
```

### 8. Encoding of the Transaction Body

Transaction Body is the `Tx.TxBody.Messages` field, which is an array of `Any`s, where each `Any` packs a `sdk.Msg`. Since `sdk.Msg`s are widely used, they have a slightly different encoding than usual array of `Any`s (Protobuf: `repeated google.protobuf.Any`) described in Annex 1.

```
This transaction has <int> message:   // Optional 's' for "message" if there's is >1 sdk.Msgs.
// For each Msg, print the following 2 lines:
Msg (<int>/<int>): <string>           // E.g. Msg (1/2): bank v1beta1 send coins
<value rendering of Msg struct>
End of transaction messages
```

#### Example

Given the following Protobuf message:

```proto
message Grant {
  google.protobuf.Any       authorization = 1 [(cosmos_proto.accepts_interface) = "Authorization"];
  google.protobuf.Timestamp expiration    = 2 [(gogoproto.stdtime) = true, (gogoproto.nullable) = false];
}

message MsgGrant {
  option (cosmos.msg.v1.signer) = "granter";
  option (cosmos.msg.v1.textual.type_url) = "authz v1beta1 grant";

  string granter = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string grantee = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

and a transaction containing 1 such `sdk.Msg`, we get the following encoding:

```
This transaction has 1 message:
Msg (1/1): authz v1beta1 grant
Granter: cosmos1abc...def
Grantee: cosmos1ghi...jkl
End of transaction messages
```

### 9. Custom `Msg` Renderers

Application developers may choose to not follow default renderer value output for their own `Msg`s. In this case, they can implement their own custom `Msg` renderer. This is similar to [EIP4430](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-4430.md), where the smart contract developer chooses the description string to be shown to the end user.

This is done by setting the `cosmos.msg.v1.textual.expert_custom_renderer` Protobuf option to a non-empty string. This option CAN ONLY be set on a Protobuf message representing transaction message object (implementing `sdk.Msg` interface).

```proto
message MsgFooBar {
  // Optional comments to describe in human-readable language the formatting
  // rules of the custom renderer.
  option (cosmos.msg.v1.textual.expert_custom_renderer) = "<unique algorithm identifier>";

  // proto fields
}
```

When this option is set on a `Msg`, a registered function will transform the `Msg` into an array of one or more strings, which MAY use the key/value format (described in point #3) with the expert field prefix (described in point #5) and arbitrary indentation (point #6). These strings MAY be rendered from a `Msg` field using a default value renderer, or they may be generated from several fields using custom logic.

The `<unique algorithm identifier>` is a string convention chosen by the application developer and is used to identify the custom `Msg` renderer. For example, the documentation or specification of this custom algorithm can reference this identifier. This identifier CAN have a versioned suffix (e.g. `_v1`) to adapt for future changes (which would be consensus-breaking). We also recommend adding Protobuf comments to describe in human language the custom logic used.

Moreover, the renderer must provide 2 functions: one for formatting from Protobuf to string, and one for parsing string to Protobuf. These 2 functions are provided by the application developer. To satisfy point #1, these 2 functions MUST be bijective with each other. Bijectivity of these 2 functions will not be checked by the SDK at runtime. However, we strongly recommend the application developer to include a comprehensive suite in their app repo to test bijectivity, as to not introduce security bugs. A simple bijectivity test looks like:

```
// for renderer, msg, and ctx of the right type
keyvals, err := renderer.Format(ctx, msg)
if err != nil {
    fail_check()
}
msg2, err := renderer.Parse(ctx, keyvals)
if err != nil {
    fail_check()
}
if !proto.Equal(msg, msg2) {
    fail_check()
}
pass_check()
```

### 10. Require signing over the `TxBody` and `AuthInfo` raw bytes

Recall that the transaction bytes merklelized on chain are the Protobuf binary serialization of [TxRaw](https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/proto/cosmos/tx/v1beta1/tx.proto#L33), which contains the `body_bytes` and `auth_info_bytes`. Moreover, the transaction hash is defined as the SHA256 hash of the `TxRaw` bytes. We require that the user signs over these bytes in SIGN_MODE_TEXTUAL, more specifically over the following string:

```
*Hash of raw bytes: <HEX(sha256(len(body_bytes) ++ body_bytes ++ len(auth_info_bytes) ++ auth_info_bytes))>
```

where:
- `++` denotes concatenation,
- `HEX` is the hexadecimal representation of the bytes, all in capital letters, no `0x` prefix,
- and `len()` is encoded as a Big-Endian uint64.

This is to prevent transaction hash malleability. The point #1 about bijectivity assures that transaction `body` and `auth_info` values are not malleable, but the transaction hash still might be malleable with point #1 only, because the SIGN_MODE_TEXTUAL strings don't follow the byte ordering defined in `body_bytes` and `auth_info_bytes`. Without this hash, a malicious validator or exchange could intercept a transaction, modify its transaction hash _after_ the user signed it using SIGN_MODE_TEXTUAL (by tweaking the byte ordering inside `body_bytes` or `auth_info_bytes`), and then submit it to Tendermint.

By including this hash in the SIGN_MODE_TEXTUAL signing payload, we keep the same level of guarantees as [SIGN_MODE_DIRECT](./adr-020-protobuf-transaction-encoding.md).

These bytes are only shown in expert mode, hence the leading `*`.

### 11. Signing Payload and Wire Format

This string array is encoded as a single `\n`-delimited string before transmitted to the hardware device, and this long string is the signing payload signed by the hardware wallet.

## Additional Formatting by the Hardware Device

Hardware devices differ in screen sizes and memory capacities. The above specifications are all verified on the protocol level, but we still allow the hardware device to add custom formatting rules that are specific to the device. Rules can include:

- if a string is too long, show it on multiple screens,
- break line between the `key` and `value` from #3,
- perform line breaks on a number or a coin values only when necessary. For example, a `sdk.Coins` with multiple denoms would be better shown as one denom per line instead of an coin amount being cut in the middle.

## Examples

#### Example 1: Simple `MsgSend`

JSON:

```json
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from": "cosmos1...abc",
        "to": "cosmos1...def",
        "amount": [
          {
            "denom": "uatom",
            "amount": 10000000
          }
        ]
      }
    ]
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": "iQ...==",
        "mode_info": { "single": { "mode": "SIGN_MODE_TEXTUAL" } },
        "sequence": 2
      }
    ],
    "fee": {
      "amount": [
        {
          "denom": "atom",
          "amount": 0.002
        }
      ],
      "gas_limit": 100000
    }
  },
  // Additional SignerData.
  "chain_id": "simapp-1",
  "account_number": 10
}
```

SIGN_MODE_TEXTUAL:

```
Chain ID: simapp-1
Account number: 10
*Public Key: iQ...==        // Hex pubkey
Sequence: 2
This transaction has 1 message:
Message (1/1): bank v1beta1 send coins
From: cosmos1...abc
To: cosmos1...def
Amount: 10 atom            // Conversion from uatom to atom using value renderers
End of transaction messages
Fee: 0.002 atom
*Gas: 100'000
*Hash of raw bytes: <hex_string>
```

#### Example 2: Multi-Msg Transaction with 3 signers

#### Example 3: Legacy Multisig

#### Example 4: Fee Payer with Tips

```json
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from": "cosmos1...tipper",
        "to": "cosmos1...abc",
        "amount": [
          {
            "denom": "uatom",
            "amount": 10000000
          }
        ]
      }
    ]
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": "iQ...==",
        "mode_info": { "single": { "mode": "SIGN_MODE_DIRECT_AUX" } },
        "sequence": 42
      },
      {
        "public_key": "iR...==",
        "mode_info": { "single": { "mode": "SIGN_MODE_TEXTUAL" } },
        "sequence": 2
      }
    ],
    "fee": {
      "amount": [
        {
          "denom": "atom",
          "amount": 0.002
        }
      ],
      "gas_limit": 100000,
      "payer": "cosmos1...feepayer"
    },
    "tip": {
      "amount": [
        {
          "denom": "ibc/CDC4587874B85BEA4FCEC3CEA5A1195139799A1FEE711A07D972537E18FDA39D",
          "amount": 200
        }
      ],
      "tipper": "cosmos1...tipper"
    }
  },
  // Additional SignerData.
  "chain_id": "simapp-1",
  "account_number": 10
}
```

SIGN_MODE_TEXTUAL for the feepayer:

```
Chain ID: simapp-1
Account number: 10
*Public Key: iR...==
Sequence: 2
This transaction has 1 message:
Message (1/1): bank v1beta1 send coins
From: cosmos1...abc
To: cosmos1...def
Amount: 10 atom
End of transaction messages
Fee: 0.002 atom
Fee Payer: cosmos1...feepayer
Tipper: cosmos1...tipper
Tip: 200 ibc/CDC4587874B85BEA4FCEC3CEA5A1195139799A1FEE711A07D972537E18FDA39D
*Gas: 100'000
*This transaction has 1 other signer:
*Signer (1/2):
*Public Key: iQ...==
*Sign mode: Direct Aux
*Sequence: 42
*End of other signers
*Hash of raw bytes: <hex_string>
```

#### Example 5: Complex Transaction with Nested Messages

JSON:

```json
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from": "cosmos1...abc",
        "to": "cosmos1...def",
        "amount": [
          {
            "denom": "uatom",
            "amount": 10000000
          }
        ]
      },
      {
        "@type": "/cosmos.gov.v1.MsgSubmitProposal",
        "proposer": "cosmos1...ghi",
        "messages": [
          {
            "@type": "/cosmos.bank.v1beta1.MsgSend",
            "from": "cosmos1...jkl",
            "to": "cosmos1...mno",
            "amount": [
              {
                "denom": "uatom",
                "amount": 20000000
              }
            ]
          },
          {
            "@type": "/cosmos.authz.v1beta1.MsgExec",
            "grantee": "cosmos1...pqr",
            "msgs": [
              {
                "@type": "/cosmos.bank.v1beta1.MsgSend",
                "from": "cosmos1...stu",
                "to": "cosmos1...vwx",
                "amount": [
                  {
                    "denom": "uatom",
                    "amount": 30000000
                  }
                ]
              },
              {
                "@type": "/cosmos.bank.v1beta1.MsgSend",
                "from": "cosmos1...abc",
                "to": "cosmos1...def",
                "amount": [
                  {
                    "denom": "uatom",
                    "amount": 40000000
                  }
                ]
              }
            ]
          }
        ],
        "initial_deposit": [
          {
            "denom": "atom",
            "amount": 100.01
          }
        ]
      }
    ]
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": "iQ...==",
        "mode_info": { "single": { "mode": "SIGN_MODE_TEXTUAL" } },
        "sequence": 2
      },
      {
        "public_key": "iR...==",
        "mode_info": { "single": { "mode": "SIGN_MODE_DIRECT" } },
        "sequence": 42
      }
    ],
    "fee": {
      "amount": [
        {
          "denom": "atom",
          "amount": 0.002
        }
      ],
      "gas_limit": 100000
    }
  },
  // Additional SignerData.
  "chain_id": "simapp-1",
  "account_number": 10
}
}
```

SIGN_MODE_TEXTUAL for 1st signer `cosmos1...abc`:

```
Chain ID: simapp-1
Account number: 10
*Public Key: iQ...==
Sequence: 2
This transaction has 2 messages:
Message (1/2): bank v1beta1 send coins
From: cosmos1...abc
To: cosmos1...def
Amount: 10 atom
Message (2/2): gov v1 submit proposal
Messages: 2 Messages
> Message (1/2): bank v1beta1 send coins
> From: cosmos1...jkl
> To: cosmos1...mno
> Amount: 20 atom
> Message (2/2): authz v1beta exec
> Grantee: cosmos1...pqr
> Msgs: 2 Msgs
>> Msg (1/2): bank v1beta1 send coins
>> From: cosmos1...stu
>> To: cosmos1...vwx
>> Amount: 30 atom
>> Msg (2/2): bank v1beta1 send coins
>> From: cosmos1...abc
>> To: cosmos1...def
>> Amount: 40 atom
> End of Msgs
End of transaction messages
Proposer: cosmos1...ghi
Initial Deposit: 100.01 atom
End of transaction messages
Fee: 0.002 atom
*Gas: 100'000
*This transaction has 1 other signer:
*Signer (2/2):
*Public Key: iR...==
*Sign mode: Direct
*Sequence: 42
*End of other signers
*Hash of raw bytes: <hex_string>
```

## Consequences

### Backwards Compatibility

SIGN_MODE_TEXTUAL is purely additive, and doesn't break any backwards compatibility with other sign modes.

### Positive

- Human-friendly way of signing in hardware devices.
- Once SIGN_MODE_TEXTUAL is shipped, SIGN_MODE_LEGACY_AMINO_JSON can be deprecated and removed. On the longer term, once the ecosystem has totally migrated, Amino can be totally removed.

### Negative

- Some fields are still encoded in non-human-readable ways, such as public keys in hexadecimal.
- New ledger app needs to be released, still unclear

### Neutral

- If the transaction is complex, the string array can be arbitrarily long, and some users might just skip some screens and blind sign.

## Further Discussions

- Some details on value renderers need to be polished, see [Annex 1](./adr-050-sign-mode-textual-annex1.md).
- Are ledger apps able to support both SIGN_MODE_LEGACY_AMINO_JSON and SIGN_MODE_TEXTUAL at the same time?
- Open question: should we add a Protobuf field option to allow app developers to overwrite the textual representation of certain Protobuf fields and message? This would be similar to Ethereum's [EIP4430](https://github.com/ethereum/EIPs/pull/4430), where the contract developer decides on the textual representation.
- Internationalization.

## References

- [Annex 1](./adr-050-sign-mode-textual-annex1.md)

- Initial discussion: https://github.com/cosmos/cosmos-sdk/issues/6513
- Living document used in the working group: https://hackmd.io/fsZAO-TfT0CKmLDtfMcKeA?both
- Working group meeting notes: https://hackmd.io/7RkGfv_rQAaZzEigUYhcXw
- Ethereum's "Described Transactions" https://github.com/ethereum/EIPs/pull/4430
