# Agoric enhancements to Cosmovisor

This particular Cosmovisor accommodates download and building from URLs referencing one of:

- as in upstream: single executable file
- as in upstream: zip file extracted verbatim to `.`, with executable file found
  at one of:
  - `./bin/$DAEMON_NAME`
  - `./$DAEMON_NAME`
- new in this fork: zip file containing a single root directory, whose extracted
  contents are moved to `.`, with executable file found at one of:
  - `./bin/$DAEMON_NAME`
  - `./$DAEMON_NAME`

To install this fork of `cosmovisor`, run the following command:

```
go install github.com/agoric-sdk/cosmos-sdk/cosmovisor/cmd/cosmovisor@Agoric
```

Note that cosmovisor does not use the the version of cosmos-sdk that it is
contained in. This is true of both the Cosmos and Agoric versions.

Have fun,
The team at [Agoric](https://github.com/Agoric).
