# chdbg
Command `chdbg` is a debugger for determining the root cause of a Cosmos chain halt.
Its initial release can analyze two databases and output their differences. As an
example, here's the output when run on the example directories from
[`iavlviewer`](https://github.com/cosmos/iavl/tree/master/cmd/iaviewer):

```
go run cosmossdk.io/tools/chdbg bns-a.db bns-b.db
chdbg: hash mismatch: 96AAD58DBDF2BA87D90BE1F620E80AC3D1662B5113A7667B51303596163A5969 != 56E581EBD9C0A3D726A91579839F7FF8A9251BEB063FDF0FA0415A0B3429DF6E
chdbg: key _i.bchnft_owner:4C97A7423B1782D7C8CAB362247B848DEC96B1EC: key proofs differ
chdbg: key _i.bchnft_owner:E28AE9A6EB94FC88B73EB7CBD6B87BF93EB9BEF0: key proofs differ
chdbg: key _i.tkrnft_owner:E28AE9A6EB94FC88B73EB7CBD6B87BF93EB9BEF0: key proofs differ
chdbg: key _i.usrnft_chainaddr:1152542575310734325L;da3ed6a45429278bac2666961289ca17ad86595d33b31037615d4b8e8f158bba: key proofs differ
chdbg: key _i.usrnft_chainaddr:12256717727036376470L;da3ed6a45429278bac2666961289ca17ad86595d33b31037615d4b8e8f158bba: key proofs differ
chdbg: key _i.usrnft_chainaddr:14285752342776807606L;da3ed6a45429278bac2666961289ca17ad86595d33b31037615d4b8e8f158bba: key proofs differ
chdbg: key _i.usrnft_chainaddr:177168082075485743L;da3ed6a45429278bac2666961289ca17ad86595d33b31037615d4b8e8f158bba: key proofs differ
chdbg: key _i.usrnft_chainaddr:2980033962229439650L;da3ed6a45429278bac2666961289ca17ad86595d33b31037615d4b8e8f158bba: key proofs differ
chdbg: key _i.usrnft_chainaddr:3070406526139113375L;da3ed6a45429278bac2666961289ca17ad86595d33b31037615d4b8e8f158bba: key proofs differ
chdbg: key _i.usrnft_chainaddr:8565302995323734695L;da3ed6a45429278bac2666961289ca17ad86595d33b31037615d4b8e8f158bba: key proofs differ
chdgb: ... (additional diffs omitted)
chdbg: database mismatch at version 190258 with 88 differences
exit status 2
```

It is expected that the tool will develop to cover more chain halt causes.

# LICENSE
Copyright Cosmos Network Authors. All Rights Reserved.
