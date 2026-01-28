# Enterprise Modules

Enterprise modules are production-ready extensions to the Cosmos SDK maintained for specialized blockchain use cases. These modules are designed for permissioned networks, consortium chains, and enterprise deployments that require features beyond traditional public blockchain architectures.

## Important: Licensing

**Enterprise modules may use different licenses than the core Cosmos SDK.** While the core Cosmos SDK is licensed under Apache-2.0, enterprise modules may have different licensing terms. Always review the LICENSE file in each enterprise module directory before using it in your project.

| Module | License | Use Case |
|--------|---------|----------|
| [PoA](./poa) | Source Available Evaluation License | Non-commercial evaluation, testing, educational purposes |

For commercial licensing inquiries, contact [legal@cosmoslabs.io](mailto:legal@cosmoslabs.io).

## Available Enterprise Modules

### [Proof of Authority (PoA)](./poa/README.md)

**License**: [Source Available Evaluation License](./poa/LICENSE)

A Cosmos SDK module that implements a Proof of Authority (PoA) consensus mechanism, allowing a designated admin to manage a permissioned validator set and integrate with governance for validator-only participation.

**Key Features:**
- Admin-controlled validator set management
- Fee distribution proportional to validator power
- Governance integration restricted to validators
- Custom vote tallying based on validator power

**Documentation:**
- [README](./poa/README.md) - Quick start and usage guide
- [API Reference](./poa/docs/api.md) - gRPC queries and transactions
- [Architecture](./poa/docs/architecture.md) - System design and module interactions
- [Distribution](./poa/docs/distribution.md) - Fee distribution mechanics
- [Governance](./poa/docs/governance.md) - Governance integration details

**Quick Links:**
- [Installation & Setup](./poa/README.md#quick-start)
- [Usage Examples](./poa/README.md#usage)
- [Configuration](./poa/README.md#configuration)
- [Development](./poa/README.md#development)

## Integration with Core SDK

Enterprise modules are designed to integrate seamlessly with the core Cosmos SDK modules. They follow the same architectural patterns and can be included in your blockchain application alongside core modules.

### Module Dependencies

Enterprise modules typically depend on and integrate with:
- **Auth** - Account management and authentication
- **Bank** - Token transfers and balance management
- **Governance** - Proposal submission and voting
- **Distribution** - Fee distribution mechanisms

### Adding Enterprise Modules to Your Chain

1. Review the enterprise module's LICENSE file
2. Follow the module-specific installation guide
3. Configure the module in your app.go
4. Update your genesis configuration
5. Test thoroughly in a development environment

## Support

For technical questions about enterprise modules:
- Review the module-specific documentation
- Check the [main SDK documentation](../docs)
- Join the [Cosmos Discord](https://discord.com/invite/interchain)

For licensing questions:
- Email [legal@cosmoslabs.io](mailto:legal@cosmoslabs.io)

## Contributing

Contributions to enterprise modules follow the same process as the core SDK. Please review the specific module's documentation for contribution guidelines.

## Comparison: Core vs Enterprise Modules

| Aspect | Core Modules (x/) | Enterprise Modules (enterprise/) |
|--------|------------------|----------------------------------|
| License | Apache-2.0 | Varies (see individual LICENSE files) |
| Maintenance | Cosmos Labs | Cosmos Labs |
| Production Ready | Yes | Yes |
| Commercial Use | Permitted under Apache-2.0 | Requires review of specific license |

## See Also

- [Core Modules Documentation](../x/README.md)
- [Cosmos SDK Documentation](https://docs.cosmos.network)
- [Main Repository README](../README.md)
