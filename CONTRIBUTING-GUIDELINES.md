# Contributing Guidelines

This document is an extension to the [CONTRIBUTING](./CONTRIBUTING.md). It describes more in details the coding guidelines and requirements.

## API & Design

+ Code must be well structured:
  + packages should have a limited responsibility (different concerns should go to different packages),
  + types should be easy to compose,
  + think about maintainbility and testability.
+ "Depend upon abstractions, [not] concretions".
+ Try to limit the number of methods you are exposing. It's easier to expose something later than to hide it.
+ Take advantage of `internal` package concept.
+ Follow agreed upon design patterns and naming conventions
+ publicly-exposed functions are named logically, have forward-thinking arguments and return types.
+ Avoid global variables and global configurators.
+ Favor composable and extensible designs.
+ Minimize code duplicaiton.
+ Limit 3rd party dependencies.

Performance:
+ Avoid unnecessary operations or memory allocations.


Security:
+ Pay proper attention to exploits
  + gas usage
  + transaction verification and signatures
  + malleability.
  + code must be always deterministic.
+ Thread safety. If some functionality is not thread safe, or uses something which is not thread safe, then indicate it clearly on each level.


## Testing and

Make sure your code is well tested:
+ Provide unit tests for every unit of your code if possible. Unit tests should take around 70%-80% of your tests.
+ Describe the test scenarios you are implementing for integration tests.
+ Create integration tests for queries and msgs.
+ Use both test cases and property / fuzzy testing. We use [rapid](pgregory.net/rapid) framework for property / fuzzy testing.
+ Code test coverage should not decrease. Explain in a PR if it decreases.



## Quality Assurance

We are forming a QA team which will support the core SDK team and collaborators by:
- Improving the Cosmos SDK and Regen Network QA Processes
- Improving automation in QA and testing
- Defining high quality metrics
- Maintaining and improving testing frameworks (unit tests, integration tests and functional tests)
- Defining test scenarios.
- Verifying user experience and defining a high quality.
    - We want to have **acceptance tests**! Document and list acceptance lists which are implemented and which are still missing.
- Supporting other teams with testing frameworks, automation and User Experience testing.
- Testing chain upgrades for every new breaking change
    - defining automated tests assuring data integrity after an update.

Desired outcomes:

- QA team works with Development Team
- QA is happening in parallel to Core development
- Releases are more predictable (0.40 and 0.43 were not predictable at all).
- QA reports. Goal is to guide with new tasks and be one of the QA measures.


As a developer, you must help the QA team by leaving instructions for user experience and functional testing.

Quality Assurance

+
