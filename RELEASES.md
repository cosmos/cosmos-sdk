# Releases

The Cosmos-SDK follows both [0ver](https://0ver.org/) and [Semver](https://semver.org/). While this is confusing lets break it down: 0ver is used for the main Cosmos-SDK dependency (`github.com/cosmos/cosmos-sdk`) and Semver is used for all other dependencies.

## Semver Dependencies

While we follow semver we have made a few changes because of the way blockchains way. The main difference is that the major version (Y.x.x) is only bumped in the case of a consensus breaking change. The minor version (x.Y.x) is bumped in the case of a non-consensus breaking change but a incompatible API breaking change. 

<p align="center">
  <img src="semver.png?raw=true" alt="Releases Semver decision tree" width="40%" />
</p>

## 0ver Dependencies

With the Cosmos-SDK depency we follow 0ver. This is a simpler flow than the previous Semver flow. When there is a consensus breaking change or api breaking change, the Cosmos-sdk team will bump the minor version (x.Y.x). When there is a non-consensus breaking change and a non-api breaking change, the Cosmos-sdk team will bump the patch version (x.x.Y).

<p align="center">
  <img src="0ver.png?raw=true" alt="Releases 0ver decision tree" width="40%" />
</p>
