# Contributing

- [Contributing](#contributing)
    - [Dev Calls](#dev-calls)
    - [Architecture Decision Records (ADR)](#architecture-decision-records-adr)
    - [Development Procedure](#development-procedure)
        - [Testing](#testing)
        - [Pull Requests](#pull-requests)
        - [Pull Request Templates](#pull-request-templates)
        - [Requesting Reviews](#requesting-reviews)
        - [Updating Documentation](#updating-documentation)
    - [Dependencies](#dependencies)
    - [Protobuf](#protobuf)
    - [Branching Model and Release](#branching-model-and-release)
        - [PR Targeting](#pr-targeting)
    - [Code Owner Membership](#code-owner-membership)
    - [Concept & Feature Approval Process](#concept--feature-approval-process)

Thank you for considering making contributions to the Cosmos SDK and related repositories!

Contributing to this repo can mean many things such as participating in
discussion or proposing code changes. To ensure a smooth workflow for all
contributors, the general procedure for contributing has been established:

1. Start by browsing [new issues](https://github.com/cosmos/cosmos-sdk/issues) and [discussions](https://github.com/cosmos/cosmos-sdk/discussions). If you are looking for something interesting or if you have something in your mind, there is a chance it was has been discussed.

- Looking for a good place to start contributing? How about checking out some [good first issues](https://github.com/cosmos/cosmos-sdk/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)?

2. Determine whether a GitHub issue or discussion is more appropriate for your needs:
1. If want to propose something new that requires specification or an additional design, or you would like to change a process, start with a [new discussion](https://github.com/cosmos/cosmos-sdk/discussions/new). With discussions, we can better handle the design process using discussion threads. A discussion usually leads to one or more issues.
2. If the issue you want addressed is a specific proposal or a bug, then open a [new issue](https://github.com/cosmos/cosmos-sdk/issues/new/choose).
3. Review existing [issues](https://github.com/cosmos/cosmos-sdk/issues) to find an issue you'd like to help with.
3. Participate in thoughtful discussion on that issue.
4. If you would like to contribute:
   1. Ensure that the proposal has been accepted.
   2. Ensure that nobody else has already begun working on this issue. If they have,
      make sure to contact them to collaborate.
   3. If nobody has been assigned for the issue and you would like to work on it,
      make a comment on the issue to inform the community of your intentions
      to begin work.
5. To submit your work as a contribution to the repository follow standard GitHub best practices. See [pull request guideline](#pull-requests) below.

**Note:** For very small or blatantly obvious problems such as typos, you are
not required to an open issue to submit a PR, but be aware that for more complex
problems/features, if a PR is opened before an adequate design discussion has
taken place in a GitHub issue, that PR runs a high likelihood of being rejected.

## Teams Dev Calls

The Cosmos SDK has many stakeholders contributing and shaping the project. Regen Network Development leads the Cosmos SDK R&D, and welcomes long-term contributors and additional maintainers from other projects. We use self-organizing principles to coordinate and collaborate across organizations in structured "Working Groups" that focus on specific problem domains or architectural components of the Cosmos SDK.

The developers are organized in working groups which are listed on a ["Working Groups & Arch Process" Github Issue](https://github.com/cosmos/cosmos-sdk/issues/9058) (pinned at the top of the [issues list](https://github.com/cosmos/cosmos-sdk/issues)).

The important development announcements are shared on [Discord](https://discord.com/invite/cosmosnetwork) in the \#dev-announcements channel.

To synchronize we have few major meetings:

+ Architecture calls: bi-weekly on Fridays at 14:00 UTC (alternating with the grooming meeting below).
+ Grooming / Planning: bi-weekly on Fridays at 14:00 UTC (alternating with the architecture meeting above).
+ Cosmos Community SDK Development Call on the last Wednesday of every month at 17:00 UTC.
+ Cosmos Roadmap Prioritization every 4 weeks on Tuesday at 15:00 UTC (limited participation).

If you would like to join one of those calls, then please contact us on [Discord](https://discord.com/invite/cosmosnetwork) or reach out directly to Cory Levinson from Regen Network (cory@regen.network).

## Architecture Decision Records (ADR)

When proposing an architecture decision for the Cosmos SDK, please start by opening an [issue](https://github.com/cosmos/cosmos-sdk/issues/new/choose) or a [discussion](https://github.com/cosmos/cosmos-sdk/discussions/new) with a summary of the proposal. Once the proposal has been discussed and there is rough alignment on a high-level approach to the design, the [ADR creation process](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/PROCESS.md) can begin. We are following this process to ensure all involved parties are in agreement before any party begins coding the proposed implementation. If you would like to see examples of how these are written, please refer to the current [ADRs](https://github.com/cosmos/cosmos-sdk/tree/master/docs/architecture).

## Development Procedure

- The latest state of development is on `master`.
- `master` must never fail `make lint test test-race`.
- No `--force` onto `master` (except when reverting a broken commit, which should seldom happen).
- Create a branch to start a wok:
    - Fork the repo (core developers must create a branch directly in the Cosmos SDK repo),
    branch from the HEAD of `master`, make some commits, and submit a PR to `master`.
    - For core developers working within the `cosmos-sdk` repo, follow branch name conventions to ensure a clear
    ownership of branches: `{moniker}/{issue#}-branch-name`.
    - See [Branching Model](#branching-model-and-release) for more details.
- Be sure to run `make format` before every commit. The easiest way
  to do this is have your editor run it for you upon saving a file (most of the editors
  will do it anyway using a pre-configured setup of the programming language mode).
  Additionally, be sure that your code is lint compliant by running `make lint-fix`.
  A convenience git `pre-commit` hook that runs the formatters automatically
  before each commit is available in the `contrib/githooks/` directory.
- Follow the [CODING GUIDELINES](CODING_GUIDELINES.md), which defines criteria for designing and coding a software.

Code is merged into master through pull request procedure.

### Testing

Tests can be executed by running `make test` at the top level of the Cosmos SDK repository.

### Pull Requests

Before submitting a pull request:

- merge the latest master `git merge origin/master`,
- run `make lint test` to ensure that all checks and tests pass.

Then:

1. If you have something to show, **start with a `Draft` PR**. It's good to have early validation of your work and we highly recommend this practice. A Draft PR also indicates to the community that the work is in progress.
   Draft PRs also helps the core team provide early feedback and ensure the work is in the right direction.
2. When the code is complete, change your PR from `Draft` to `Ready for Review`.
3. Go through the actions for each checkbox present in the PR template description. The PR actions are automatically provided for each new PR.
4. Be sure to include a relevant changelog entry in the `Unreleased` section of `CHANGELOG.md` (see file for log format).

PRs must have a category prefix that is based on the type of changes being made (for example, `fix`, `feat`,
`refactor`, `docs`, and so on). The *type* must be included in the PR title as a prefix (for example,
`fix: <description>`). This convention ensures that all changes that are committed to the base branch follow the
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.
Additionally, each PR should only address a single issue.

Pull requests are merged automatically using [`automerge` action](https://mergify.io/features/auto-merge).

NOTE: when merging, GitHub will squash commits and rebase on top of the master.

### Pull Request Templates

There are three PR templates. The [default template](./.github/PULL_REQUEST_TEMPLATE.md) is for types `fix`, `feat`, and `refactor`. We also have a [docs template](./.github/PULL_REQUEST_TEMPLATE/docs.md) for documentation changes and an [other template](./.github/PULL_REQUEST_TEMPLATE/other.md) for changes that do not affect production code. When previewing a PR before it has been opened, you can change the template by adding one of the following parameters to the url:

- `template=docs.md`
- `template=other.md`

### Requesting Reviews

In order to accommodate the review process, the author of the PR must complete the author checklist
(from the pull request template)
to the best of their abilities before marking the PR as "Ready for Review". If you would like to
receive early feedback on the PR, open the PR as a "Draft" and leave a comment in the PR indicating
that you would like early feedback and tagging whoever you would like to receive feedback from.

Codeowners are marked automatically as the reviewers.

All PRs require at least two review approvals before they can be merged (one review might be acceptable in
the case of minor changes to [docs](./.github/PULL_REQUEST_TEMPLATE/docs.md) or [other](./.github/PULL_REQUEST_TEMPLATE/other.md) changes that do not affect production code). Each PR template has a reviewers checklist that must be completed before the PR can be merged. Each reviewer is responsible
for all checked items unless they have indicated otherwise by leaving their handle next to specific
items. In addition, use the following review explanations:

- `LGTM` without an explicit approval means that the changes look good, but you haven't thoroughly reviewed the reviewer checklist items.
- `Approval` means that you have completed some or all of the reviewer checklist items. If you only reviewed selected items, you must add your handle next to the items that you have reviewed. In addition, follow these guidelines:
    - You must also think through anything which ought to be included but is not
    - You must think through whether any added code could be partially combined (DRYed) with existing code
    - You must think through any potential security issues or incentive-compatibility flaws introduced by the changes
    - Naming must be consistent with conventions and the rest of the codebase
    - Code must live in a reasonable location, considering dependency structures (for example, not importing testing modules in production code, or including example code modules in production code).
    - If you approve the PR, you are responsible for any issues mentioned here and any issues that should have been addressed after thoroughly reviewing the reviewer checklist items in the pull request template.
- If you sat down with the PR submitter and did a pairing review, add this information in the `Approval` or your PR comments.
- If you are only making "surface level" reviews, submit notes as a `comment` review.

### Updating Documentation

If you open a PR on the Cosmos SDK, it is mandatory to update the relevant documentation in `/docs`.

- If your change relates to the core SDK (baseapp, store, ...), be sure to update the content in `docs/basics/`, `docs/core/` and/or `docs/building-modules/` folders.
- If your changes relate to the core of the CLI (not specifically to module's CLI/Rest), then modify the content in the `docs/run-node/` folder.
- If your changes relate to a module, then be sure to update the module's spec in `x/moduleName/docs/spec/`.

When writing documentation, follow the [Documentation Writing Guidelines](./docs/DOC_WRITING_GUIDELINES.md).

## Dependencies

We use [Go Modules](https://github.com/golang/go/wiki/Modules) to manage
dependency versions.

The master branch of every Cosmos repository should just build with `go get`,
which means they should be kept up-to-date with their dependencies, so we can
get away with telling people they can just `go get` our software.

Since some dependencies are not under our control, a third party may break our
build, in which case we can fall back on `go mod tidy -v`.

## Protobuf

We use [Protocol Buffers](https://developers.google.com/protocol-buffers) along with [gogoproto](https://github.com/gogo/protobuf) to generate code for use in Cosmos SDK.

For determinstic behavior around Protobuf tooling, everything is containerized using Docker. Make sure to have Docker installed on your machine, or head to [Docker's website](https://docs.docker.com/get-docker/) to install it.

For formatting code in `.proto` files, you can run `make proto-format` command.

For linting and checking breaking changes, we use [buf](https://buf.build/). You can use the commands `make proto-lint` and `make proto-check-breaking` to respectively lint your proto files and check for breaking changes.

To generate the protobuf stubs, you can run `make proto-gen`.

We also added the `make proto-all` command to run all the above commands sequentially.

In order for imports to properly compile in your IDE, you may need to manually set your protobuf path in your IDE's workspace settings/config.

For example, in vscode your `.vscode/settings.json` should look like:

```
{
    "protoc": {
        "options": [
        "--proto_path=${workspaceRoot}/proto",
        "--proto_path=${workspaceRoot}/third_party/proto"
        ]
    }
}
```

## Branching Model and Release

User-facing repos should adhere to the trunk based development branching model: https://trunkbaseddevelopment.com/. User branches should start with a user name, example: `{moniker}/{issue#}-branch-name`.

The Cosmos SDK repository is a [multi Go module](https://github.com/golang/go/wiki/Modules#is-it-possible-to-add-a-module-to-a-multi-module-repository) repository. It means that we have more than one Go module in a single repository.

The Cosmos SDK utilizes [semantic versioning](https://semver.org/).

### PR Targeting

Ensure that you base and target your PR on the `master` branch.

All feature additions and all bug fixes must be targeted against `master`. Exception is for bug fixes which are only related to a released version. In that case, the related bug fix PRs must target against the release branch.

If needed, we backport a commit from `master` to a release branch (excluding consensus breaking feature, API breaking and similar).

## Code Owner Membership

In the ethos of open source projects, and out of necessity to keep the code
alive, the core contributor team will strive to permit special repo privileges
to developers who show an aptitude towards developing with this code base.

Several different kinds of privileges may be granted however most common
privileges to be granted are merge rights to either part of, or the entirety of the
code base (through the GitHub `CODEOWNERS` file). The on-boarding process for
new code owners is as follows: On a bi-monthly basis (or more frequently if
agreeable) all the existing code owners will privately convene to discuss
potential new candidates as well as the potential for existing code-owners to
exit or "pass on the torch". This private meeting is to be a held as a
phone/video meeting.

Subsequently after the meeting, and pending final approval from the ICF,
one of the existing code owners should open a PR modifying the `CODEOWNERS` file.
The other code owners should then all approve this PR to publicly display their support.

Only if unanimous consensus is reached among all the existing code-owners will
an invitation be extended to a new potential-member. Likewise, when an existing
member is suggested to be removed/or have their privileges reduced, the member
in question must agree on the decision for their removal or else no action
should be taken. If however, a code-owner is demonstrably shown to intentionally
have had acted maliciously or grossly negligent, code-owner privileges may be
stripped with no prior warning or consent from the member in question.

Other potential removal criteria:

* Missing 3 scheduled meetings results in ICF evaluating whether the member should be
    removed / replaced
* Violation of Code of Conduct

Earning this privilege should be considered to be no small feat and is by no
means guaranteed by any quantifiable metric. Serving as a code owner is a symbol of great trust from
the community of this project.

## Concept & Feature Approval Process

The process for how Cosmos SDK maintainers take features and ADRs from concept to release
is broken up into three distinct stages: **Strategy Discovery**, **Concept Approval**, and
**Implementation & Release Approval**

### Strategy Discovery

* Develop long term priorities, strategy and roadmap for the Cosmos SDK
* Release committee not yet defined as there is already a roadmap that can be used for the time being

### Concept Approval

* Architecture Decision Records (ADRs) may be proposed by any contributors or maintainers of the Cosmos SDK,
    and should follow the guidelines outlined in the
    [ADR Creation Process](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/PROCESS.md)
* After proposal, a time bound period for Request for Comment (RFC) on ADRs commences
* ADRs are intended to be iterative, and may be merged into `master` while still in a `Proposed` status

**Time Bound Period**

* Once a PR for an ADR is opened, reviewers are expected to perform a first
  review within 1 week of pull request being open
* Time bound period for individual ADR Pull Requests to be merged should not exceed 2 weeks
* Total time bound period for an ADR to reach a decision (`ABANDONED | ACCEPTED | REJECTED`) should not exceed 4 weeks

If an individual Pull Request for an ADR needs more time than 2 weeks to reach resolution, it should be merged
in current state (`Draft` or `Proposed`), with its contents updated to summarize
the current state of its discussion.

If an ADR is taking longer than 4 weeks to reach a final conclusion, the **Concept Approval Committee**
should convene to rectify the situation by either:

- unanimously setting a new time bound period for this ADR
- making changes to the Concept Approval Process (as outlined here)
- making changes to the members of the Concept Approval Committee

**Approval Committee & Decision Making**

In absence of general consensus, decision making requires 1/2 vote from the two members
of the **Concept Approval Committee**.

**Committee Members**

* Core Members: **Aaron** (Regen), **Bez** (IG)

**Committee Criteria**

Members must:

* Participate in all or almost all ADR discussions, both on GitHub as well as in bi-weekly Architecture Review
  meetings
* Be active contributors to the Cosmos SDK, and furthermore should be continuously making substantial contributions
  to the project's codebase, review process, documentation and ADRs
* Have stake in the Cosmos SDK project, represented by:
    * Being a client / user of the Comsos SDK
    * "[giving back](https://www.debian.org/social_contract)" to the software
* Delegate representation in case of vacation or absence

Code owners need to maintain participation in the process, ideally as members of **Concept Approval Committee**
members, but at the very least as active participants in ADR discussions

Removal criteria:

* Missing 3 meetings results in ICF evaluating whether the member should be removed / replaced
* Violation of Code of Conduct

### Implementation & Release Approval

The following process should be adhered to both for implementation PRs corresponding to ADRs, as
well as for PRs made as part of a release process:

* Code reviewers should ensure the PR does exactly what the ADR said it should
* Code reviewers should have more senior engineering capability
* 1/2 approval is required from the **primary repo maintainers** in `CODEOWNERS`

**Note**: For any major release series denoted as a "Stable Release" (e.g. v0.42 "Stargate"), a separate release
committee is often established. Stable Releases, and their corresponding release committees are documented
separately in [Stable Release Policy](./RELEASE_PROCESS.md#stable-release-policy)*
