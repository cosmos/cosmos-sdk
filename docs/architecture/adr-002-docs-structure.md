# ADR 002: SDK Documentation Structure

## Context

The current Cosmos SDK documentation structure presents several challenges:

- It includes a significant amount of non-SDK-related material, making it difficult to navigate.
- The documentation is hard to maintain due to its scattered nature.
- Users struggle to find the relevant information efficiently.

To address these issues, we propose a structured and scalable documentation framework that segregates content logically and makes it easier to maintain and update.

### Goals

- Ensure that all documentation related to development frameworks and tools reside in their respective GitHub repositories:
  - The SDK repository should contain SDK-related documentation.
  - The Hub repository should contain Hub-specific documentation.
  - The Lotion repository should contain Lotion-related documentation.
- Store all general and high-level materials, such as FAQs, whitepapers, and introductory overviews, on the Cosmos website.

## Decision

We will restructure the `/docs` directory of the Cosmos SDK GitHub repository as follows:

```plaintext
docs/
├── README.md
├── intro/
│   ├── overview.md
│   ├── getting-started.md
│   ├── installation.md
│   ├── tutorials.md
│   ├── faq.md
├── concepts/
│   ├── baseapp.md
│   ├── types.md
│   ├── store.md
│   ├── server.md
│   ├── modules/
│   │   ├── keeper.md
│   │   ├── handler.md
│   │   ├── cli.md
│   ├── gas.md
│   └── commands.md
├── clients/
│   ├── lite/
│   │   ├── introduction.md
│   │   ├── setup.md
│   ├── service-providers.md
├── modules/
│   ├── governance.md
│   ├── staking.md
│   ├── slashing.md
│   ├── auth.md
│   ├── bank.md
│   ├── distribution.md
│   ├── evidence.md
├── spec/
│   ├── modules/
│   ├── architecture.md
│   ├── security.md
├── translations/
│   ├── README.md
│   ├── en/
│   ├── es/
│   ├── zh/
│   ├── ru/
├── architecture/
│   ├── adr-001.md
│   ├── adr-002.md
│   ├── system-design.md
└── _attic/
    ├── deprecated_docs/
    ├── legacy_guides.md
```

### Explanation of Sections

- **`README.md`**: Serves as the landing page for the documentation, providing an introduction and directing users to key resources.
- **`intro/`**: Contains introductory material, including an overview, setup guides, FAQs, and tutorials to help new users get started.
- **`concepts/`**: Provides high-level explanations of Cosmos SDK abstractions. These are not API specifications but conceptual overviews that do not require frequent updates.
- **`clients/`**: Contains specifications and guides for different Cosmos SDK clients, including Lite clients and service providers.
- **`modules/`**: Lists and details the Cosmos SDK modules, including governance, staking, slashing, authentication, and banking.
- **`spec/`**: Stores technical specifications, including module details and security guidelines.
- **`translations/`**: Houses translated versions of the documentation to support a global audience.
- **`architecture/`**: Includes architecture-related documents such as ADRs (Architecture Decision Records) and system design documents.
- **`_attic/`**: Contains deprecated or legacy documentation that is no longer actively maintained but may still be relevant for reference.

### Website Documentation Structure

The documentation displayed on the Cosmos website will include only the following sections to maintain clarity and focus:

- `README`
- `intro`
- `concepts`
- `clients`

The `architecture` section will be excluded from the website to avoid overwhelming general users with internal architectural decisions.

## Status

**Accepted**

## Consequences

### Positive Outcomes

- **Improved Organization**: The documentation is now logically structured, making it easier to navigate and maintain.
- **Focused SDK Content**: The `/docs` directory now exclusively contains SDK-related material, eliminating unnecessary content.
- **Easier Contribution**: Developers can now update only the `/docs` folder when making changes, without affecting unrelated sections.
- **Cleaner Build Process**: The `vuepress` build for website documentation is simplified, improving efficiency.
- **Foundation for Executable Documentation**: The new structure aligns with ongoing efforts to create executable documentation.

### Neutral Changes

- **Migration of Deprecated Content**: Outdated materials will be moved to the `/_attic` folder for archival purposes.
- **Integration of Existing Content**: Documentation from `docs/sdk/docs/core` will be merged into the `concepts` section.
- **Relocation of Non-SDK Material**: Introductory content, whitepapers, and Lotion-related materials will be transferred to the Cosmos website repository.
- **Update of Documentation Guidelines**: The `DOCS_README.md` file will be revised to reflect the new structure and guidelines.

## References

- [GitHub Issue #1460](https://github.com/cosmos/cosmos-sdk/issues/1460)
- [GitHub Pull Request #2695](https://github.com/cosmos/cosmos-sdk/pull/2695)
- [GitHub Issue #2611](https://github.com/cosmos/cosmos-sdk/issues/2611)

