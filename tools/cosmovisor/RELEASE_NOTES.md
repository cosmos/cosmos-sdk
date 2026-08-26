# Cosmovisor v1.7.3 Release Notes

Cosmovisor v1.7.3 makes prebuilt binaries available for the v1.7.2 fix for builds with Go 1.26 and later failing with `invalid reference to encoding/json.unquoteBytes`. The v1.7.2 GitHub release did not include those binaries.

See the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/tools/cosmovisor/v1.7.3/tools/cosmovisor/CHANGELOG.md) for details on the changes in v1.7.3.

## Installation instructions

```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.7.3
```
