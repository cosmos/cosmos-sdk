# Maintaining Cosmos SDK Proto Files

All of the Cosmos SDK proto files are defined here. This folder should
be synced regularly with buf.build/cosmos/cosmos-sdk regularly by
a maintainer by running `buf push` in this folder.

User facing documentation should not be placed here but instead goes in
`buf.md` and in each protobuf package following the guidelines in
https://docs.buf.build/bsr/documentation.
