<!--
order: 2
-->

# State

The internal state of the `x/upgrade` module is relatively minimal and simple. The
state only contains the currently active upgrade `Plan` (if one exists) by key
`0x0` and if a `Plan` is marked as "done" by key `0x1`.

The `x/upgrade` module contains no genesis state.
