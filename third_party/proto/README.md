# Buf

To generate the protos that were present in this folder run:

```bash
buf export buf.build/cosmos/cosmos-sdk:$(curl -sS https://api.github.com/repos/cosmos/cosmos-sdk/commits/<sdk_version_tag> | jq -r .sha) --output .
```
