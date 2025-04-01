# Cosmos SDK Fork

Fork of Cosmos SDK v0.50.x for Celestia App.
The fork include the following changes compared to upstream:

* Store app version in consensus param store
* Modify continuous vesting account to add start time 
* Re-add query router for custom abci queries
* Add v0.52 helpers to facilitate testing
* Disable heavy bank migrations
* Backport improvements for DOS protection for x/authz
* Support historical account number queries 

Read the [CHANGELOG.md](CHANGELOG.md) for more details.
