# Buf

To generate the protos that were present in this folder run:

```bash
buf export buf.build/cosmos/cosmos-sdk:${commit} --output .
```

where `${commit}` is the commit of the buf commit of version of the Cosmos SDK you are using.
That commit can be found [here](https://github.com/cosmos/cosmos-sdk/blob/main/proto/README.md).
