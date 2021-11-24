# ADR 047: Extend Upgrade Plan

## Changelog

- Nov, 23, 2021: Initial Draft

## Status

DRAFT Not Implemented

## Abstract

This ADR expands the existing upgrade `Plan` proto message to include new fields for defining pre-run and post-run processes.
It also defines a structure for providing downloadable assets involved in an upgrade.

## Context

The `upgrade` module in conjunction with Cosmovisor are designed to facilitate and automate a blockchain's transition from one version to another.

Users submit a software upgrade governance proposal containing an upgrade `Plan`.
The `Plan` currently contains the following fields:
- `name`: A short string identifying the new version.
- `height`: The chain height at which the upgrade is to be performed.
- `info`: A string containing information about the upgrade. It can either be a json object (with a specific structure defined through documentation), or it can be a URL that will return such json.

The json object from the `info` field identifies URLs used to download the new blockchain executable for different platforms (OS and Architecture, e.g. "linux/amd64").
Such a URL can either return the executable file directly or can return an archive containing the executable and possibly other assets.
If an archive contains a `bin/pre-upgrade` file, Cosmovisor executes it (and waits for it to finish) before attempting to restart the chain.

Currently, there is no mechanism that makes Cosmovisor run a command after the upgraded chain has been restarted.

## Decision

### Protobuf Updates

We will define a new message for providing upgrade instructions and add it as a new field in the `Plan` message.
The upgrade instructions will contain a list of assets available for each platform.
It will also allow for the definition of a pre-run and post-run command.

```protobuf
message Plan {
  // ... (existing fields)

  UpgradeInstructions upgrade = 6;
}
```

The new `UpgradeInstructions upgrade` field MUST be optional.

```protobuf
message UpgradeInstructions {
  string pre_run = 1;
  string post_run = 2;
  repeated Asset assets    = 3;
  string description = 4;
}
```

All fields in the `UpgradeInstructions` SHOULD be optional.
- `pre_run` is a command to run prior to the upgraded chain restarting.
  If defined, this command SHOULD behave the same as the current [pre-upgrade](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/pre-upgrade.md) command.
  The working directory this command runs from SHOULD be `{DAEMON_HOME}/{upgrade name}`.
- `post_run` is a command to run after the upgraded chain has been started. If defined, this command SHOULD be only executed once.
  The output and exit code SHOULD be logged but SHOULD NOT affect the running of the upgraded chain.
  The working directory this command runs from SHOULD be `{DAEMON_HOME}/{upgrade name}`.
- `assets` SHOULD allow any number of entries.
  It SHOULD have only one entry per platform.
- `description` contains additional information about the upgrade and might contain references to external resources.
  It SHOULD NOT be used for structured processing information.

```protobuf
message Asset {
  string platform = 1;
  string url = 2;
  string checksum = 3;
}
```

- `platform` is a required string that SHOULD be in the format `{OS}/{CPU}`, e.g. `"linux/amd64"`.
  The string `"any"` SHOULD also be allowed.
  An `Asset` with a `platform` of `"any"` SHOULD be used as a fallback when a specific `{OS}/{CPU}` entry is not found.
  That is, if an `Asset` exists with a `platform` that matches the system's OS and CPU, that should be used;
  otherwise, if an `Asset` exists with a `platform` of `any`, that should be used;
  otherwise no asset should be downloaded.
- `url` is a required string that MUST conform to [RFC 1738: Uniform Resource Locators](https://www.ietf.org/rfc/rfc1738.txt).
- `checksum` SHOULD be provided, but is not required.
  It SHOULD be a hex encoded checksum of the expected result of a request to the `url`.
  Implementations utilizing these `UpgradeInstructions` SHOULD fail if the checksum of the result of the `url` is not equal to this provided `checksum`.
  SHA-256 SHOULD be the default hashing algorithm.
  This `checksum` field SHOULD allow other hashing algorithms by allowing the checksum to be prefixed by the algorithm, then a colon.
  E.g. A `checksum` of `"sha256:6e4003b3104a5936d8142191352ba6b7af530fcb883578946b538a816a9571f9"` SHOULD be equal to `"6e4003b3104a5936d8142191352ba6b7af530fcb883578946b538a816a9571f9"`,
  and a `checksum` of `"sha512:2cafa43204f0f51229c01c1bdd3b29b6a446bfdc313a474a7091a87fe122d3defaee0de7aa7534496385c8dad02a8d76a8c39634c742677d20c9fdc2da59448e"` SHOULD be allowed.
  If the `url` contains a `checksum` query parameter, this `checksum` field SHOULD still be populated and if populated MUST equal the query parameter `checksum` value.

### Upgrade module updates.

If an upgrade `Plan` does not use the new `UpgradeInstructions` field, existing functionality will be maintained.

We will update the creation of the `upgrade-info.json` file to include the `UpgradeInstructions`.

We will update the optional validation available via CLI to account for the new `Plan` structure.
We will add the following validation:
1.  If `UpgradeInstructions` are provided:
    1.  There MUST be at least one entry in `assets`.
    1.  All of the `assets` MUST have a unique `platform`.
1.  The following validation is currently done using the `info` field. We will apply similar validation to the `UpgradeInstructions`.
    For each `Asset`:
    1.  The `platform` MUST have the format `{OS}/{CPU}` or be `"any"`.
    1.  The `url` field MUST NOT be empty.
    1.  The `url` field MUST be a proper URL.
    1.  A `checksum` MUST be provided either in the `checksum` field or as a query parameter in the `url`.
    1.  If the `checksum` field has a value and the `url` also has a `checksum` query parameter, the two values MUST be equal.
    1.  The `url` MUST return either a file or an archive containing eitehr `bin/{DAEMON_NAME}` or `{DAEMON_NAME}`.
    1.  If a `checksum` is provided (in the field or as a query param), the checksum of the result of the `url` MUST equal the provided checksum.

### Cosmovisor updates

If the `upgrade-info.json` file does not contain any `UpgradeInstructions`, existing functionality will be maintained.

We will update Cosmovisor to look for and handle the new `UpgradeInstructions` in `upgrade-info.json`.
If the `UpgradeInstructions` are provided, we will do the following:
1.  The `info` field will be ignored.
1.  The `assets` field will be used to identify the asset to download based on the `platform` that Cosmovisor is running in.
1.  If the downloaded asset is an archive that contains a `bin/pre-upgrade` file, that file will be ignored.
1.  If a `checksum` is provided (either in the field or as a query param in the `url`), and the downloaded asset has a different checksum, the upgrade process will be interrupted and Cosmovisor will exit with an error.
1.  If a `pre_run` command is defined, it will be executed at the same point in the process where the `bin/pre-upgrade` file would have been executed.
    It will be executed using the same environment as other commands run by Cosmovisor.
1.  If a `post_run` command is defined, it will be executed after executing the command that restarts the chain.
    It will be executed in a background process using the same environment as the other commands.
    Any output generated by the command will be logged.
    Once complete, the exit code will be logged.

## Consequences

### Backwards Compatibility

Since the only change to existing definitions is the addition of the `upgrade` field to the `Plan` message, and that field MUST be optional, there are no backwards incompatibilities with respects to the proto messages.
Additionally, current behavior will be maintained when no `UpgradeInstructions` are provided, so there are no backwards incompatibilities with respects to either the upgrade module or Cosmovisor.

### Forwards Compatibility

In order to utilize the `UpgradeInstructions` as part of a software upgrade, both of the following must be true:
1.  The chain must already be using a sufficiently advanced version of the Cosmos SDK.
1.  The chain's nodes must be using a sufficiently advanced version of Cosmovisor.

### Positive

1.  The structure for defining downloadable assets is clearer since it is now defined in the proto instead of in documentation.
1.  Availability of a pre-run command becomes more obvious.
1.  A post-run command becomes possible.

### Negative

1.  The `Plan` message becomes larger.
1.  There is no option for providing a URL that will return the `UpgradeInstructions`.
    Existing functionality of the `info` field will remain, but will not allow for the newly defined structure.
1.  There is no built-in mechanism for preventing the `assets` from having two entries with the same `platform`.
    Behavior of this situation is undefined.
1.  All `assets` are assumed to either be the executable or be an archive that contains the executable.
    There is no way to define more than one downloadable asset for a platform.
1.  A chain software upgrade is required before this new functionality can be used.
    I.e. The next upgrade must use the old definitions, but the one after that can use the new structure.

### Neutral

1. Existing functionality is maintained when the `UpgradeInstructions` aren't provided.

## Further Discussions

1.  [Draft PR #10032 Comment](https://github.com/cosmos/cosmos-sdk/pull/10032/files?authenticity_token=pLtzpnXJJB%2Fif2UWiTp9Td3MvRrBF04DvjSuEjf1azoWdLF%2BSNymVYw9Ic7VkqHgNLhNj6iq9bHQYnVLzMXd4g%3D%3D&file-filters%5B%5D=.go&file-filters%5B%5D=.proto#r698708349):
    Consider different names for `UpgradeInstructions upgrade` (either the message type or field name).
1.  [Draft PR #10032 Comment](https://github.com/cosmos/cosmos-sdk/pull/10032/files?authenticity_token=pLtzpnXJJB%2Fif2UWiTp9Td3MvRrBF04DvjSuEjf1azoWdLF%2BSNymVYw9Ic7VkqHgNLhNj6iq9bHQYnVLzMXd4g%3D%3D&file-filters%5B%5D=.go&file-filters%5B%5D=.proto#r754655072):
    1.  Consider putting the `string platform` field inside `UpgradeInstructions` and make `UpgradeInstructions` a repeated field in `Plan`.
    1.  Consider using a `oneof` field in the `Plan` which could either be `UpgradeInstructions` or else a URL that should return the `UpgradeInstructions`.
    1.  Consider allowing `info` to either be a JSON serialized version of `UpgradeInstructions` or else a URL that returns that.
1.  [Draft PR #10032 Comment](https://github.com/cosmos/cosmos-sdk/pull/10032/files?authenticity_token=pLtzpnXJJB%2Fif2UWiTp9Td3MvRrBF04DvjSuEjf1azoWdLF%2BSNymVYw9Ic7VkqHgNLhNj6iq9bHQYnVLzMXd4g%3D%3D&file-filters%5B%5D=.go&file-filters%5B%5D=.proto#r755462876):
    Consider not including the `UpgradeInstructions.description` field, using the `info` field for that purpose instead.
1.  [Draft PR #10032 Comment](https://github.com/cosmos/cosmos-sdk/pull/10032/files?authenticity_token=pLtzpnXJJB%2Fif2UWiTp9Td3MvRrBF04DvjSuEjf1azoWdLF%2BSNymVYw9Ic7VkqHgNLhNj6iq9bHQYnVLzMXd4g%3D%3D&file-filters%5B%5D=.go&file-filters%5B%5D=.proto#r754643691):
    Consider allowing multiple assets to be downloaded for any given `platform` by adding a `name` field to the `Asset` message.

## References

- [Current upgrade.proto](https://github.com/cosmos/cosmos-sdk/blob/master/proto/cosmos/upgrade/v1beta1/upgrade.proto)
- [Upgrade Module README](https://github.com/cosmos/cosmos-sdk/blob/master/x/upgrade/spec/README.md)
- [Cosmovisor README](https://github.com/cosmos/cosmos-sdk/blob/master/cosmovisor/README.md)
- [Pre-upgrade README](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/pre-upgrade.md)
- [Draft/POC PR #10032](https://github.com/cosmos/cosmos-sdk/pull/10032)
- [RFC 1738: Uniform Resource Locators](https://www.ietf.org/rfc/rfc1738.txt)
