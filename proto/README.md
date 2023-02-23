# Maintaining Cosmos SDK Proto Files

All of the Cosmos SDK proto files are defined here. This folder should
be synced regularly with buf.build/cosmos/cosmos-sdk regularly by
a maintainer by running `buf push` in this folder.

User facing documentation should not be placed here but instead goes in
`buf.md` and in each protobuf package following the guidelines in
https://docs.buf.build/bsr/documentation.

## SDK x Buf

| Cosmos SDK Version | Buf Commit Version                                                                                            |
| ------------------ | ------------------------------------------------------------------------------------------------------------- |
| Prior v0.46.0      | [Unavailable](https://github.com/bufbuild/buf/issues/1415)                                                    |
| v0.46.x            | [8cb30a2c4de74dc9bd8d260b1e75e176](https://buf.build/cosmos/cosmos-sdk/docs/8cb30a2c4de74dc9bd8d260b1e75e176) |
| v0.47.x            | [2aa7ff2b23df473a85b7a7fe1884105d](https://buf.build/cosmos/cosmos-sdk/docs/2aa7ff2b23df473a85b7a7fe1884105d) |

## Generate

To get the Cosmos SDK proto given a commit, run: 

```bash
buf export buf.build/cosmos/cosmos-sdk:${commit} --output .
```
