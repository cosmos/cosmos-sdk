# Cosmos SDK v0.44.9 Release Notes:

Upon working on making archive nodes for Juno, it became clear that the migration path wasn't very pleasant to walk at all. 

This release was mainly prepared with crypto.org in mind, after preparing some updates to v0.44.5-patch.

This should EOL the v0.44.* series, because keeping it secure is very challenging.  

Use v0.44.9 to move away from the 44 series SDK.


# Cosmos SDK v0.44.5-patch Release Notes - Dragonberry Patch

This is a security release for the [Dragonberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702).
Please upgrade ASAP.

Next to this, we have also included a few minor bugfixes.

Chains must add the following to their go.mod for the application:

```go
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go
```

Bumping the SDK version should be smooth, however, feel free to tag core devs to review your upgrading PR:

- **CET**: @tac0turtle, @okwme, @AdityaSripal, @colin-axner, @julienrbrt
- **EST**: @ebuchman, @alexanderbez, @aaronc
- **PST**: @jtremback, @nicolaslara, @czarcas7ic, @p0mvn
- **CDT**: @ValarDragon, @zmanian
# Cosmos SDK v0.44.8 Release Notes

This release introduces only a Tendermint dependency update to v0.34.19 which
itself includes two bug fixes related to consensus. See the full changelog from
v0.34.17-v0.34.19 [here](https://github.com/tendermint/tendermint/blob/v0.34.19/CHANGELOG.md#v0.34.19).

See the [Cosmos SDK v0.44.8 Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.44.8/CHANGELOG.md)
for the exhaustive list of all changes.

**Full Changelog**: https://github.com/cosmos/cosmos-sdk/compare/v0.44.7...v0.44.8
