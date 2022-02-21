# ADR 049: SIGN_MODE_TEXTUAL - Annex 1 Value Renderers

## Changelog

- Dec 06, 2021: Initial Draft

## Status

Draft

## Abstract

This Annex describes value renderers, which are used for displaying values in a human-friendly way.

## Value Renderers

Value Renderers describe how values of different Protobuf types should be automatically rendered. Value renderers can be formalized as a set of bijective functions `func renderT(value T) string`, where `T` is one of the below Protobuf types on which this spec is defined.

### Protobuf `number`

- Applies to numeric integer types (`uint64`, etc.) and their casted types (e.g. in Golang: `sdk.Dec`, `sdk.Int`).
- Formatting with `'`s for every three integral digits.
- Usage of `.` to denote the decimal delimitier.

### Examples

- sdk.Int, integers: `1000` -> `1'000`
- sdk.Dec; `1000000.00` -> `1'000'000.00`

### `coin`

- Applies to `cosmos.base.v1beta1.Coin`.
- Denoms are converted to `display` denoms using `Metadata` (if available). **This requires a state query**.
- Amounts are converted to `display` denom amounts and rendered as `number`s above
- One space between the denom and amount
- In the future, IBC denoms could maybe be converted to DID/IIDs, if we can find a robust way for doing this (ex. `cosmos:hub:atom`)
- Ex:
  - `1000000000uatom` -> `1'000 atom`

### `type_url`

- all protobuf messages to be used with `SIGN_MODE_TEXTUAL` should have a short name associated with them that can be used in format strings whenever the type url is explicitly referenced (as in the `MsgRevoke` examples below).
- these could be options in a proto messages or config files

```proto
message MsgSend {
  option (cosmos.textual) {
    msg_name = "bank send coins"
  }
}
```

- they should be unique per message, per chain
- Ex:
  - `cosmos.bank.v1beta1.MsgSend` -> `bank send coins`
  - `cosmos.gov.v1beta1.MsgVote` -> `governance vote`

### `repeated`

TODO

### `message`

TODO

### Enums

- String case convention: snake case to sentence case
- Allow optional annotation for textual name (TBD)
- E.g `enum VoteOption`
  - convert enum name (`VoteOption`) to snake_case (`VOTE_OPTION`)
  - truncate that prefix + `_` from the enum name if it exists (`VOTE_OPTION_` gets stripped from `VOTE_OPTION_YES` -> `YES`)
  - convert rest to sentence case: `YES` -> `Yes`
  - in summary: `VOTE_OPTION_YES` -> `Yes`

### `google.protobuf.Timestamp` (TODO)

Rendered as either ISO8601 (`2021-01-01T12:00:00Z`) or a more standard English-language date format (`Jan. 1, 2021 12:00 UTC`)

### `google.protobuf.Duration` (TODO)

- rendered in terms of weeks, days, hours, minutes and seconds as these time units can be measured independently of any calendar and duration values are in seconds (so months and years can't be used precisely)
- total seconds values included at the end so users have both pieces of information
- Ex:
  - `1483530 seconds` -> `2 weeks, 3 days, 4 hours, 5 minutes, 30 seconds (1483530 seconds total)`

### address bytes

We currently use `string` types in protobuf for addresses so this may not be needed, but if any address bytes are used in sign mode textual they should be rendered with bech32 formatting
