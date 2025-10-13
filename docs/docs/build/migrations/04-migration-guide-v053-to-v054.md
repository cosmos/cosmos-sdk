---
sidebar_position: 4
---

# Migration Guide: v0.53.x to v0.54.x

This guide provides step-by-step instructions for migrating your Cosmos SDK application from v0.53.x to v0.54.x, which includes the major upgrade from CometBFT v0.x.x to CometBFT v2.

:::warning
🚨 **This upgrade requires a coordinated chain upgrade** 🚨

All validators must upgrade simultaneously as this involves breaking changes to the consensus layer.
:::

## Prerequisites

Before starting the migration, ensure your application meets these requirements:

### Starting Assumptions

- **SDK Version**: You are currently on Cosmos SDK v0.53.x
- **CometBFT Version**: You are on CometBFT v0.38.x
- **IBC-go Version**: You are on IBC-go v10.x
- **Go Version**: Go 1.24.0 or later

### Pre-Migration Checklist

- [ ] Backup your current application state and configuration
- [ ] Test the migration on a testnet first
- [ ] Ensure all validators are coordinated for the upgrade
- [ ] Review the [UPGRADING.md](../../../../UPGRADING.md) for breaking changes
- [ ] Check the [CHANGELOG.md](../../../../CHANGELOG.md) for detailed changes

## Migration Steps

### Step 1: Update Dependencies

Update your `go.mod` file to use the new versions:

```bash
go get github.com/cosmos/cosmos-sdk@v0.54.x
go get github.com/cometbft/cometbft@v2.x.x
go get github.com/cosmos/ibc-go/v10@latest
go mod tidy
```

### Step 2: Configuration Migration

Use the **Confix** tool to migrate your configuration files:

#### Migrate app.toml

```bash
# Using confix standalone
confix migrate v0.54 ~/.your-app/config/app.toml

# Or using confix integrated in your app
your-app config migrate v0.54
```

#### Migrate client.toml

```bash
# Using confix standalone
confix migrate v0.54 ~/.your-app/config/client.toml --client

# Or using confix integrated in your app
your-app config migrate v0.54 --client
```

#### Verify Configuration Changes

```bash
# Check what changes will be made
confix diff v0.54 ~/.your-app/config/app.toml

# View current configuration
confix view ~/.your-app/config/app.toml
```

### Step 3: Genesis Migration (if needed)

If you need to migrate genesis state, use the genesis migration tool:

```bash
# Migrate genesis file
your-app genesis migrate v0.54 /path/to/genesis.json --chain-id=your-chain-id

# Validate migrated genesis
your-app genesis validate-genesis /path/to/migrated-genesis.json
```

### Step 4: Application Code Updates

#### Update Import Statements

Update your imports to use the new SDK version:

```go
import (
    "github.com/cosmos/cosmos-sdk@v0.54.x/..."
    "github.com/cometbft/cometbft@v2.x.x/..."
)
```

#### Handle CometBFT v2 Changes

The main change in CometBFT v2 is the deprecation of `TimeoutCommit` in favor of `NextBlockDelay`.

**Option 1: Use existing timeout_commit values (Recommended)**

No code changes needed. Your existing `timeout_commit` values in `config.toml` will continue to work.

**Option 2: Set NextBlockDelay in your application**

If you want to override the `timeout_commit` values, add this to your `app.go`:

```go
import (
    "time"
    "github.com/cosmos/cosmos-sdk/baseapp"
)

func NewApp() *App {
    // ... existing code ...
    
    app := baseapp.NewBaseApp(
        // ... existing parameters ...
    )
    
    // Set NextBlockDelay to override timeout_commit
    app.SetNextBlockDelay(2 * time.Second) // Adjust as needed
    
    // ... rest of your app setup ...
}
```

### Step 5: Upgrade Handler Setup

Create or update your upgrade handler for the v0.53 to v0.54 migration:

```go
// In your app.go or upgrades.go file
const UpgradeName = "v053-to-v054"

func (app *App) RegisterUpgradeHandlers() {
    app.UpgradeKeeper.SetUpgradeHandler(
        UpgradeName,
        func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
            // Run module migrations
            return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
        },
    )

    upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
    if err != nil {
        panic(err)
    }

    if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
        storeUpgrades := storetypes.StoreUpgrades{
            Added: []string{
                // Add any new modules here if you're adding them
            },
        }

        // Configure store loader for the upgrade
        app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
    }
}
```

### Step 6: Build and Test

```bash
# Build your application
make build

# Test on a local testnet
make testnet

# Run integration tests
make test
```

### Step 7: Deploy the Upgrade

1. **Submit Upgrade Proposal** (if using governance):
   ```bash
   your-app tx gov submit-proposal software-upgrade "v053-to-v054" \
     --title "Upgrade to v0.54.x" \
     --description "Upgrade to Cosmos SDK v0.54.x with CometBFT v2" \
     --upgrade-height <BLOCK_HEIGHT> \
     --from <PROPOSER_KEY>
   ```

2. **Vote on the Proposal**:
   ```bash
   your-app tx gov vote <PROPOSAL_ID> yes --from <VOTER_KEY>
   ```

3. **Coordinate Validator Upgrades**:
   - All validators must upgrade before the upgrade height
   - Use Cosmovisor for automatic upgrades (recommended)

## Cosmovisor Setup (Recommended)

For seamless upgrades, use Cosmovisor:

```bash
# Install Cosmovisor
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Set up environment
export DAEMON_NAME=your-app
export DAEMON_HOME=~/.your-app
export DAEMON_RESTART_AFTER_UPGRADE=true

# Create upgrade directory
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v053-to-v054/bin

# Copy current binary to genesis
cp your-app $DAEMON_HOME/cosmovisor/genesis/bin/

# Copy new binary to upgrade directory
cp your-app-v0.54 $DAEMON_HOME/cosmovisor/upgrades/v053-to-v054/bin/
```

## Verification

After the upgrade, verify everything is working:

```bash
# Check version
your-app version

# Check node status
your-app status

# Check if upgrade was successful
your-app query upgrade applied <UPGRADE_NAME>
```

## Rollback Plan

If issues occur, you can rollback:

1. **Stop the node**
2. **Restore from backup**
3. **Use the previous binary**
4. **Restart the node**

:::warning
Rollback is only possible if the upgrade hasn't been applied yet. Once the upgrade is applied, rollback requires a coordinated effort from all validators.
:::

## Troubleshooting

See the [FAQ section](#frequently-asked-questions) below for common issues and solutions.

## Additional Resources

- [UPGRADING.md](../../../../UPGRADING.md) - Detailed upgrade reference
- [CHANGELOG.md](../../../../CHANGELOG.md) - Complete list of changes
- [CometBFT v2 Documentation](https://github.com/cometbft/cometbft/blob/main/docs/)
- [Cosmovisor Documentation](../tooling/01-cosmovisor.md)
- [Confix Documentation](../tooling/02-confix.md)

## Frequently Asked Questions

### Q: Do I need to change my timeout_commit values?

**A:** No, your existing `timeout_commit` values will continue to work. CometBFT v2 maintains backward compatibility. Only set `NextBlockDelay` if you want to override these values.

### Q: What happens if a validator doesn't upgrade in time?

**A:** The validator will be slashed and may be jailed. This is why coordination is critical for this upgrade.

### Q: Can I test this upgrade on a testnet?

**A:** Yes, absolutely! Always test upgrades on a testnet first. You can use the same migration steps on a testnet environment.

### Q: How long does the migration take?

**A:** The actual migration is very fast (seconds), but you need to coordinate with all validators. The total process depends on your network's governance and coordination.

### Q: What if the upgrade fails?

**A:** If the upgrade fails before being applied, you can rollback. If it fails after being applied, you'll need to coordinate with all validators to fix the issue.

### Q: Do I need to update my client applications?

**A:** Client applications using the standard Cosmos SDK interfaces should continue to work. However, check the changelog for any breaking changes that might affect your specific use case.

### Q: Can I use this guide for other SDK versions?

**A:** This guide is specifically for v0.53.x to v0.54.x. For other versions, refer to the appropriate upgrade guides in the [migrations documentation](./01-intro.md).

### Q: What about IBC connections?

**A:** IBC connections should continue to work normally. However, ensure you're using a compatible version of IBC-go v10.x.

### Q: How do I verify the upgrade was successful?

**A:** Use these commands:
```bash
your-app query upgrade applied v053-to-v054
your-app status
your-app version
```

### Q: Can I skip the configuration migration?

**A:** While the configuration migration is not strictly required, it's recommended to ensure compatibility and take advantage of new features and optimizations.

### Q: What if I have custom modules?

**A:** Custom modules should continue to work without changes. However, review the changelog for any breaking changes that might affect your modules.

### Q: How do I handle the upgrade in a production environment?

**A:** 
1. Test thoroughly on testnet
2. Coordinate with all validators
3. Use Cosmovisor for automatic upgrades
4. Have a rollback plan ready
5. Monitor the network closely during the upgrade

### Q: Are there any performance implications?

**A:** CometBFT v2 includes performance improvements, so you should see better performance after the upgrade.

### Q: What about third-party tools and services?

**A:** Check with your third-party tool providers to ensure compatibility with CometBFT v2. Most major tools should be compatible, but verify before upgrading.

---

## Support

If you encounter issues during the migration:

1. Check the [troubleshooting section](#troubleshooting) above
2. Review the [UPGRADING.md](../../../../UPGRADING.md) for detailed information
3. Search existing [GitHub issues](https://github.com/cosmos/cosmos-sdk/issues)
4. Create a new issue if your problem isn't documented

Remember: Always test on a testnet first and coordinate with your validator community!
