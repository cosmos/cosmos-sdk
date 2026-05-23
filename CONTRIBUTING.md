# Contributing to Cosmos SDK

Thank you for considering contributing to the Cosmos SDK!

## Ways to Contribute

- 🐛 Bug reports and fixes
- 📝 Documentation improvements
- 💻 Code contributions
- ✅ Testing
- 📖 Examples and tutorials

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Git

### Setup
```bash
git clone https://github.com/cosmos/cosmos-sdk.git
cd cosmos-sdk
make install
```

### Running Tests
```bash
make test
```

## Development Guidelines

### Code Style
- Follow Go conventions
- Run `gofmt` before committing
- Run `golangci-lint run` for linting

### Commit Messages
Use conventional commits:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation
- `test:` for tests
- `refactor:` for refactoring

### Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add/update tests
5. Update documentation
6. Submit PR with clear description

## Resources
- [Documentation](https://docs.cosmos.network)
- [Discord](https://discord.gg/cosmosnetwork)

## License
By contributing, you agree that your contributions will be licensed under Apache 2.0.
