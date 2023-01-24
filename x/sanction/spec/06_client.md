<!--
order: 6
-->

# Client

A user can interact with the `x/sanction` module using `gRPC`, `CLI`, or `REST`.

## gRPC

A user can interact with and query the `x/sanction` module using `gRPC`.

For details see [Msg Service](03_messages.md) or [gRPC Queries](05_queries.md).

## CLI

The `gRPC` transaction and query endpoints are made available through CLI helpers.

### Transactions

The transaction endpoints are only for use with governance proposals.
As such, the CLI's `tx gov` commands can be used to interact with them.

### Queries

Each of these commands facilitates running a `gRPC` query.
Standard `query` flags are available unless otherwise noted.

#### IsSanctioned

```shell
$ simd query sanction is-sanctioned --help
Check if an address is sanctioned.

Examples:
  $ simd query sanction is-sanctioned cosmos1v4uxzmtsd3j4zat9wfu5zerywgc47h6luruvdf
  $ simd query sanction is cosmos1v4uxzmtsd3j4zat9wfu5zerywgc47h6luruvdf
  $ simd query sanction check cosmos1v4uxzmtsd3j4zat9wfu5zerywgc47h6luruvdf

Usage:
  simd query sanction is-sanctioned <address> [flags]

Aliases:
  is-sanctioned, is, check, is-sanction
```

#### SanctionedAddresses

```shell
$ simd query sanction sanctioned-addresses --help
List all the sanctioned addresses.

Examples:
  $ simd query sanction sanctioned-addresses
  $ simd query sanction addresses
  $ simd query sanction all

Usage:
  simd query sanction sanctioned-addresses [flags]

Aliases:
  sanctioned-addresses, addresses, all
```

Standard pagination flags are also available for this command.

#### TemporaryEntries

```shell
simd query sanction temporary-entries --help
List all temporarily sanctioned/unsanctioned addresses.
If an address is provided, only temporary entries for that address are returned.
Otherwise, all temporary entries are returned.

Examples:
  $ simd query sanction temporary-entries
  $ simd query sanction temporary-entries cosmos1v4uxzmtsd3j4zat9wfu5zerywgc47h6luruvdf
  $ simd query sanction temp-entries
  $ simd query sanction temp-entries cosmos1v4uxzmtsd3j4zat9wfu5zerywgc47h6luruvdf
  $ simd query sanction temp
  $ simd query sanction temp cosmos1v4uxzmtsd3j4zat9wfu5zerywgc47h6luruvdf

Usage:
  simd query sanction temporary-entries [<address>] [flags]

Aliases:
  temporary-entries, temp-entries, temp
```

Standard pagination flags are also available for this command.

#### Params

```shell
$ simd query sanction params --help
Get the sanction module params.

Example:
  $ simd query sanction params

Usage:
  simd query sanction params [flags]
```

## REST

Each of the sanction `gRPC` query endpoints is also available through one or more `REST` endpoints.

| Name                        | URL                                               |
|-----------------------------|---------------------------------------------------|
| IsSanctioned                | `/cosmos/sanction/v1beta1/check/{address}`        |
| SanctionedAddresses         | `/cosmos/sanction/v1beta1/all`                    |
| TemporaryEntries - all      | `/cosmos/sanction/v1beta1/temp`                   |
| TemporaryEntries - specific | `/cosmos/sanction/v1beta1/temp?address={address}` |
| Params                      | `/cosmos/sanction/v1beta1/params`                 |

For `SanctionedAddresses` and `TemporaryEntries`, pagination parameters can be provided using the standard pagination query parameters.