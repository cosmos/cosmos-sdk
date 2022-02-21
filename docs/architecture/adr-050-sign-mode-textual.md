# ADR 050: SIGN_MODE_TEXTUAL

## Changelog

- Dec 06, 2021: Initial Draft.
- Feb 07, 2022: Draft read and concept-ACKed by the Ledget team.

## Status

Draft

## Abstract

This ADR specifies SIGN_MODE_TEXTUAL, a new string-based sign mode that is targetted at signing with hardware devices.

## Context

Protobuf-based SIGN_MODE_DIRECT has been introduced in [ADR-020](./adr-020-protobuf-transaction-encoding.md) and is intended to replace SIGN_MODE_LEGACY_AMINO_JSON in most situations, such as mobile wallets and CLI keyrings. However, the [Ledger](https://www.ledger.com/) hardware wallet is still using SIGN_MODE_LEGACY_AMINO_JSON for displaying the sign bytes to the user. Hardware wallets cannot transition to SIGN_MODE_DIRECT as it is binary-based and thus not suitable for display to end-users.

In an effort to remove Amino from the SDK, a new sign mode needs to be created for hardware devices. [Initial discussions](https://github.com/cosmos/cosmos-sdk/issues/6513) propose a string-based sign mode, which this ADR formally specifies.

## Decision

We propose to have SIGN_MODE_TEXTUAL’s signing payload `SignDocTextual` to be an array of strings. Each string would correspond to one "screen" on the hardware wallet device, with no (or little, TBD) additional formatting done by the hardware wallet app itself.

```proto
message SignDocTextual {
  repeated string screens = 1;
}
```

The string array MUST follow the specifications below.

### 1. Bijectivity with Protobuf transactions

The encoding and decoding operations between a Protobuf transaction (whose definition can be found [here](https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/tx/v1beta1/tx.proto#L13)) and the string array MUST be bijective.

We concede that bijectivity is not strictly needed. Avoiding transaction malleability only requires collision resistance on the encoding. Lossless encoding also does not require decodability. However, bijectivity assures both non-malleability and losslessness.

This also prevents users signing over hashed transaction metadata, which is a security concern for Ledger (the company).

We propose to maintain functional tests using bijectivity in the SDK to assure losslessness and no malleability.

### 2. Only ASCII 32-127 characters allowed

Ledger devices have limited character display capabilities, so all strings MUST only contain ASCII characters in the 32-127 range.

In particular, the newline `"\n"` (ASCII: 10) character is forbidden.

### 3. All strings have the `<key>: <value>` format

All strings MUST match the following Regex: `TODO`.

This is helpful for UIs displaying SignDocTextual to users. This MAY be used in the Ledger app to perform custom on-screen formatting, for example to break long lines into multiple screens.

The `<value>` itself can contain the `": "` characters.

### 4. Values are encoded using Value Renderers

Value Renderers describe how values of different types are rendered in the string array. The full specification of Value Renderers can be found in [Annex 1](./adr-048-sign-mode-textual-annex1.md).

### 5. Strings starting with `*` are only shown in Expert mode

Ledger devices have the an Expert mode for advanced users. Strings starting with the `*` character will only be shown in Expert mode.

### 6. The string array format

Below is the general format of a TX with N msgs. Each new line corresponds to a new screen on the Ledger device. `//` denotes comments and are not shown on the Ledger device.

### 7. Encoding of the Transaction Envelope

We define "transaction envelope" as all data in a transaction that is not in the `TxBody`. Transaction envelope includes fee, signer infos and memo, but don't include `Msg`s.

```
Chain ID: <string>
Account number: <uint64>
*Public Key: <base64_string>
Sequence: <uint64>
<TxBody>                                    // See 8.
Fee: <coins>
*Fee payer: <string>                        // Skipped if no fee_payer set
*Fee granter: <string>                      // Skipped if no fee_granter set
Memo: <string>                              // Skipped if no memo set
*Gas Limit: <uint64>
*Timeout Height:  <uint64>                  // Skipped if no timeout_height set
Tipper: <string>                            // If there's a tip
Tip: <string>
*This transaction has <int> other signers:  // Skipped if there is only one signer
*Signer (<int>/<int>):
*Public Key: <base64_string>
*Sign mode: <string>                        // "Direct", "Direct Aux", "Legacy Amino Json", Enum value renderer
*Sequence: <uint64>
*End of other signers
```

### 8. Encoding of the Transaction Body

We call "transaction body" the `Tx.TxBody.Messages` field, which is an array of Anys.

```
This transaction has <int> messages:
// For each Msg, print the following 2 lines:
Msg (<int>/<int>): <string>           // E.g. Msg (1/2): bank v1beta1 send coins
<value rendering of Msg struct>
End of transaction messages
```

### Wire Format

This string array is encoded as a single `\n`-delimited string before transmitted to the hardware device.

### Examples

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
*Public Key: iQ...==        // Base64 pubkey
Sequence: 2
This transaction has 1 message:
Message (1/1): bank v1beta1 send coins
From: cosmos1...abc
To: cosmos1...def
Amount: 10 atom            // Conversion from uatom to atom using value renderers
End of transaction messages
Fee: 0.002 atom
*Gas: 100'000
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
        "@type": "/cosmos.gov.v1beta2.MsgSubmitProposal",
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
Message (2/2): gov v1beta2 submit proposal
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
```

## Consequences

### Backwards Compatibility

### Positive

### Negative

### Neutral

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## References

- Initial discussion: https://github.com/cosmos/cosmos-sdk/issues/6513
- Living document used in the working group: https://hackmd.io/fsZAO-TfT0CKmLDtfMcKeA?both
- Working group meeting notes: https://hackmd.io/7RkGfv_rQAaZzEigUYhcXw
- Ethereum's "Described Transactions" https://github.com/ethereum/EIPs/pull/4430
