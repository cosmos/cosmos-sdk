# ADR 606: Rosetta standalone

## Changelog

* May 31, 2023: Initial draft (@Bizk)

## Status

DRAFT

## Abstract

Rosetta implementation runs on top of cosmos-sdk. The tool should rely on stable cosmos-sdk
releases and not in experimental features in order to ensure usability and stability.
Having the tool inside the main cosmos-sdk repo may cause unrequired updates, and makes the
tool maintainance more complicated.|

## Context

In the current context; Rosetta is inside cosmos-sdk project due to easier maintainability.
This is a problem because as seen in some previous commits, it has suffered unrequired changes
like updates due to a refactor or a new experimental feature. Causing issues like missmatch
dependencies or unexpected behaviour that needs to be addressed down the line.

Since Rosetta is a standalone tool that has been implemented using Cosmos-sdk, it should only
rely on top of the latest releases, instead of being affected by experimental features or
interfeering over the cosmos-sdk development.

Decoupling Rosetta from the cosmos-sdk project is no only a good practice since it keeps the
scope well defined, but also it ensures better maintainability over time.

## Alternatives

The 2 altnernatives are:
- Keep the Project inside Cosmos-sdk: This might be easier to handle from a fully inmersed
maintainer point of view, who works on different aspects of cosmos-sdk and might want to try
features into Rosetta.
- Move the project into another repo: This approach intends to keep the scope well defined,
ensure maintainability and stability over time.

## Decision

We will move Rosetta into a standalone repo in order to keep the Scope well defined, a stable
versioning, and to ensure an easier and cleaner maintainability.

We will also add and keep track of quality standards as tests and following cosmos-sdk releases.

## Consequences

The consequence will be a new repo that contains the rosetta tool implemented by cosmos.

### Positive

- Better scope.
- Easier miantainability and testing.
- Avoids unrequired changes.
- Easier to fork / contribute.

### Negative

- Adds one more repo to keep track of.
- More dificult to test experimental features from cosmos-sdk

## Further Discussions

