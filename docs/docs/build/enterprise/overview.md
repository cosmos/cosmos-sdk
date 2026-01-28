# Enterprise Modules

Enterprise modules are production-ready extensions to the Cosmos SDK maintained for specialized blockchain use cases. These modules are designed for permissioned networks, consortium chains, and enterprise deployments.

:::warning License Notice
**Enterprise modules may use different licenses than the core Cosmos SDK.** While the core Cosmos SDK is licensed under Apache-2.0, enterprise modules may have different licensing terms. Always review the LICENSE file in each enterprise module directory before using it in your project.
:::

## Available Enterprise Modules

### Proof of Authority (PoA)

**License**: [Source Available Evaluation License](https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE)

A module that implements Proof of Authority consensus, allowing a designated admin to manage a permissioned validator set.

**Key Features:**
- Admin-controlled validator set management
- Fee distribution proportional to validator power
- Governance integration restricted to validators
- Custom vote tallying based on validator power

**Use Cases:**
- Consortium blockchains
- Permissioned networks
- Testing and development environments
- Private enterprise blockchains

**Documentation:**
- [PoA Module README](https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/README.md)
- [PoA API Reference](https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/docs/api.md)
- [PoA Architecture](https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/docs/architecture.md)
- [Distribution Mechanics](https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/docs/distribution.md)
- [Governance Integration](https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/docs/governance.md)

## Licensing

Enterprise modules support multiple licensing options:

| Module | License Type | Use Case | Commercial Use |
|--------|-------------|----------|----------------|
| PoA | Source Available Evaluation | Testing, evaluation, education | Requires commercial license |

### Commercial Licensing

To use enterprise modules in production or for commercial purposes, contact Cosmos Labs:

**Email**: [legal@cosmoslabs.io](mailto:legal@cosmoslabs.io)

## Integration

Enterprise modules integrate seamlessly with core Cosmos SDK modules and follow the same architectural patterns. To add an enterprise module to your chain:

1. Review the module's LICENSE file
2. Follow the module-specific installation guide
3. Configure the module in your `app.go`
4. Update your genesis configuration
5. Test thoroughly in a development environment

## Support

For technical questions about enterprise modules:
- Review module-specific documentation (linked above)
- Check the [main SDK documentation](https://docs.cosmos.network)
- Join the [Cosmos Discord](https://discord.com/invite/interchain)

For licensing questions:
- Email [legal@cosmoslabs.io](mailto:legal@cosmoslabs.io)

## Repository Structure

Enterprise modules are located in the `enterprise/` directory of the Cosmos SDK repository:

```
cosmos-sdk/
├── enterprise/
│   ├── README.md          # Enterprise modules overview
│   └── poa/               # Proof of Authority module
│       ├── README.md
│       ├── LICENSE
│       ├── docs/
│       └── x/poa/
└── x/                     # Core SDK modules
```

## See Also

- [Core Modules Documentation](../modules/README.md)
- [Building Modules](../building-modules/intro.md)
- [Cosmos SDK Documentation](https://docs.cosmos.network)
