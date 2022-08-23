# Buf

To generate the protos that were present in this folder run:

```bash
buf export buf.build/cosmos/cosmos-sdk:$(git log -1 --pretty=format:"%H" <sdk_version_tag>) --output .
```