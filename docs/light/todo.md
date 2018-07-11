# TODO

This document is a place to gather all points for future development.

## API

* finalise ICS0 - TendermintAPI
  * make sure that the explorer and voyager can use it
* add ICS21 - StakingAPI
* add ICS22 - GovernanceAPI
* split Gaia Light into reusable components that other zones can leverage
  * it should be possible to register extra standards on the light client
  * the setup should be similar to how the app is currently started
* implement Gaia light and the general light client in Rust
  * export the API as a C interface
  * write thin wrappers around the C interface in JS, Swift and Kotlin/Java
