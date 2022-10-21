# Cosmos SDK v0.46.3 Release Notes

This is a security release for the [Dragonberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702). 
Please upgrade ASAP.

Chains must add the following to their go.mod for the application:

```go
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

Bumping the SDK version should be smooth, however, feel free to tag core devs to review your upgrading PR:

* **CET**: @tac0turtle, @okwme, @AdityaSripal, @colin-axner, @julienrbrt
* **EST**: @ebuchman, @alexanderbez, @aaronc
* **PST**: @jtremback, @nicolaslara, @czarcas7ic, @p0mvn
* **CDT**: @ValarDragon, @zmanian

Other updates:

* `ApplicationQueryService` was introduced to enable additional query service registration. Applications should implement `RegisterNodeService(client.Context)` method to automatically expose chain information query service implemented in [#13485](https://github.com/cosmos/cosmos-sdk/pull/13485). 
* Next to this, we have also included a few minor bugfixes.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.
