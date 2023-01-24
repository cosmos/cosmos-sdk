<!--
order: 6
-->

# Client

A user can interact with the `x/quarantine` module using `gRPC`, `CLI`, or `REST`.

## gRPC

A user can interact with and query the `x/quarantine` module using `gRPC`.

For details see [Msg Service](03_messages.md) or [gRPC Queries](05_queries.md).

## CLI

The `gRPC` transaction and query endpoints are made available through CLI helpers.

### Transactions

Each of these commands facilitates generating, signing and sending of a `tx`.
Standard `tx` flags are available unless otherwise noted.

In these commands, the `<to_name_or_address>` can either be a name from your keyring or your account address.
The `--from` flag is ignored since that is being conveyed using the `<to_name_or_address>` 1st argument to each command.

#### OptIn

```shell
$ simd tx quarantine opt-in --help
Activate quarantine for an account.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).

Usage:
  simd tx quarantine opt-in [<to_name_or_address>] [flags]

Examples:

$ simd tx quarantine opt-in cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389
$ simd tx quarantine opt-in personal
$ simd tx quarantine opt-in --from cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389
$ simd tx quarantine opt-in --from personal
```

#### OptOut

```shell
$ simd tx quarantine opt-out --help
Deactivate quarantine for an account.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).

Usage:
  simd tx quarantine opt-out [<to_name_or_address>] [flags]

Examples:

$ simd tx quarantine opt-out cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389
$ simd tx quarantine opt-out personal
$ simd tx quarantine opt-out --from cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389
$ simd tx quarantine opt-out --from personal
```

#### Accept

```shell
$ ./build/simd tx quarantine accept --help
Accept quarantined funds sent to <to_name_or_address> from <from_address>.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).

Usage:
  simd tx quarantine accept <to_name_or_address> <from_address> [<from_address 2> ...] [flags]

Examples:

$ simd tx quarantine accept cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389 cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut
$ simd tx quarantine accept personal cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut
$ simd tx quarantine accept personal cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut cosmos1phx24ecmuw3s7fmy8c87gh3rdq5lwskqur3t00
```

At least one `<from_address>` is required, but multiple can be provided.

A `--permanent` flag is also available with this command:

```shell
      --permanent                Also set auto-accept for sends from any of the from_addresses to to_address
```

#### Decline

```shell
$ simd tx quarantine decline --help
Decline quarantined funds sent to <to_name_or_address> from <from_address>.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).

Usage:
  simd tx quarantine decline <to_name_or_address> <from_address> [<from_address 2> ...] [flags]

Examples:

$ simd tx quarantine decline cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389 cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut
$ simd tx quarantine decline personal cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut
$ simd tx quarantine decline personal cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut cosmos1phx24ecmuw3s7fmy8c87gh3rdq5lwskqur3t00
```

At least one `<from_address>` is required, but multiple can be provided.

A `--permanent` flag is also available with this command:

```shell
      --permanent                Also set auto-decline for sends from any of the from_addresses to to_address
```

#### UpdateAutoResponses

```shell
$ simd tx quarantine update-auto-responses --help
Update auto-responses for transfers to <to_name_or_address> from one or more addresses.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).

The <to_name_or_address> is required.
At least one <auto-response> and <from_address> must be provided.

Valid <auto-response> values:
  "accept" or "a" - turn on auto-accept for the following <from_address>(es).
  "decline" or "d" - turn on auto-decline for the following <from_address>(es).
  "unspecified", "u", "off", or "o" - turn off auto-responses for the following <from_address>(es).

Each <auto-response> value can be repeated as an arg as many times as needed as long as each is followed by at least one <from_address>.
Each <from_address> will be assigned the nearest preceding <auto-response> value.

Usage:
  simd tx quarantine update-auto-responses <to_name_or_address> <auto-response> <from_address> [<from_address 2> ...] [<auto-response 2> <from_address 3> [<from_address 4> ...] ...] [flags]

Aliases:
  update-auto-responses, auto-responses, uar

Examples:

$ simd tx quarantine update-auto-responses cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389 accept cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut
$ simd tx quarantine update-auto-responses personal decline cosmos1phx24ecmuw3s7fmy8c87gh3rdq5lwskqur3t00 unspecified cosmos1lfuwk97g6y9du8altct63vwgz5620t929n8g9l
$ simd tx quarantine auto-responses personal accept cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut cosmos1qsjw3kjaf33qk2urxg54lzxkw525ngghzneujh off cosmos1lfuwk97g6y9du8altct63vwgz5620t929n8g9l
```

### Queries

Each of these commands facilitates running a `gRPC` query.
Standard `query` flags are available unless otherwise noted.

#### IsQuarantined

```shell
$ simd query quarantine is-quarantined --help
Query whether an account is opted into quarantined.

Examples:
  $ simd query quarantine is-quarantined cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389
  $ simd query quarantine is cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut

Usage:
  simd query quarantine is-quarantined <to_address> [flags]

Aliases:
  is-quarantined, is
```

#### QuarantinedFunds

```shell
simd query quarantine funds --help
Query for quarantined funds.

If no arguments are provided, all quarantined funds will be returned.
If only a to_address is provided, only undeclined funds quarantined for that address are returned.
If both a to_address and from_address are provided, quarantined funds will be returned regardless of whether they've been declined.

Examples:
  $ simd query quarantine funds
  $ simd query quarantine funds cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389
  $ simd query quarantine funds cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389 cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut

Usage:
  simd query quarantine funds [<to_address> [<from_address>]] [flags]
```

Standard pagination flags are also available for this command.

#### AutoResponses

```shell
$ simd query quarantine auto-responses --help
Query auto-responses.

If only a to_address is provided, all auto-responses set up for that address are returned. This will only contain accept or decline entries.
If both a to_address and from_address are provided, exactly one result will be returned. This can be accept, decline or unspecified.

Examples:
  $ simd query quarantine auto-responses cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389
  $ simd query quarantine auto-responses cosmos1c7p4v02eayvag8nswm4f5q664twfe6dxjha389 cosmos1ld2qyt9pq5n8dxkp58jn3jyxh8u8ztmrk9vrut

Usage:
  simd query quarantine auto-responses <to_address> [<from_address>] [flags]

Aliases:
  auto-responses, auto, ar
```

Standard pagination flags are also available for this command.

## REST

Each of the quarantine `gRPC` query endpoints is also available through one or more `REST` endpoints.

| Name                        | URL                                                            |
|-----------------------------|----------------------------------------------------------------|
| IsQuarantined               | `/cosmos/quarantine/v1beta1/active/{to_address}`               |
| QuarantinedFunds - all      | `/cosmos/quarantine/v1beta1/funds`                             |
| QuarantinedFunds - some     | `/cosmos/quarantine/v1beta1/funds/{to_address}`                |
| QuarantinedFunds - specific | `/cosmos/quarantine/v1beta1/funds/{to_address}/{from_address}` |
| AutoResponses - some        | `/cosmos/quarantine/v1beta1/auto/{to_address}`                 |
| AutoResponses - specific    | `/cosmos/quarantine/v1beta1/auto/{to_address}/{from_address}`  |

For `QuarantinedFunds` and `AutoResponses`, pagination parameters can be provided using the standard pagination query parameters.