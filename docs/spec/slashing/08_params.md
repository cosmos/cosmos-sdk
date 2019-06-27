# Parameters

The slashing module contains the following parameters:

| Key                     | Type             | Example                |
|-------------------------|------------------|------------------------|
| MaxEvidenceAge          | string (time ns) | "120000000000"         |
| SignedBlocksWindow      | string (int64)   | "100"                  |
| MinSignedPerWindow      | string (dec)     | "0.500000000000000000" |
| DowntimeJailDuration    | string (time ns) | "600000000000"         |
| SlashFractionDoubleSign | string (dec)     | "0.050000000000000000" |
| SlashFractionDowntime   | string (dec)     | "0.010000000000000000" |
