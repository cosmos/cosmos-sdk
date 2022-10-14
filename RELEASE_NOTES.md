# Cosmos SDK v0.45.9 Release Notes

This is a security release for the 
[Dragonberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702). 
Please upgrade ASAP.

Next to this, we have also included a few minor bugfixes.

Chains must add the following to their go.mod for the application:

```go
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23
```

Bumping the SDK version should be smooth, however, feel free to tag core devs to review your upgrading PR:

- **CET**: @tac0turtle, @okwme, @AdityaSripal, @colin-axner, @julienrbrt
- **EST**: @ebuchman, @alexanderbez, @aaronc
- **PST**: @jtremback, @nicolaslara, @czarcas7ic, @p0mvn
- **CDT**: @ValarDragon, @zmanian
