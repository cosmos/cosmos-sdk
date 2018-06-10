# TESTNET STATUS

## *June 5, 2018, 21:00 EST* - New Release

- Released gaia
  [v0.17.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.17.5) 
  with update for Tendermint
  [v0.19.9](https://github.com/tendermint/tendermint/releases/tag/v0.19.9)
- Fixes many bugs!
    - evidence gossipping 
    - mempool deadlock
    - WAL panic
    - memory leak
- Please update to this to put a stop to the rampant invalid evidence gossiping
  :)

## *May 31, 2018, 14:00 EST* - New Release

- Released gaia
  [v0.17.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.17.4) with update for Tendermint v0.19.7
- Fixes a WAL bug and some more
- Please update to this if you have trouble restarting a node

## *May 31, 2018, 2:00 EST* - Testnet Halt

- A validator equivocated last week and Evidence is being rampantly gossipped
- Peers that can't process the evidence (either too far behind or too far ahead) are disconnecting from the peers that
  sent it, causing high peer turn-over
- The high peer turn-over may be causing a memory-leak, resulting in some nodes
  crashing and the testnet halting
- We need to fix some issues in the EvidenceReactor to address this and also
  investigate the possible memory-leak

## *May 29, 2018* - New Release

- Released v0.17.3 with update for Tendermint v0.19.6
- Fixes fast-sync bug
- Please update to this to sync with the testnet
