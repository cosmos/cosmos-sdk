# Contributing

- [Contributing](#contributing)
  - [Architecture Decision Records (ADR)](#architecture-decision-records-adr)
  - [Pull Requests](#pull-requests)
    - [Process for reviewing PRs](#process-for-reviewing-prs)
    - [Updating Documentation](#updating-documentation)
  - [Forking](#forking)
  - [Dependencies](#dependencies)
  - [Protobuf](#protobuf)
  - [Testing](#testing)
  - [Branching Model and Release](#branching-model-and-release)
    - [PR Targeting](#pr-targeting)
    - [Development Procedure](#development-procedure)
    - [Pull Merge Procedure](#pull-merge-procedure)
    - [Release Procedure](#release-procedure)
    - [Point Release Procedure](#point-release-procedure)
  - [Code Owner Membership](#code-owner-membership)

Thank you for considering making contributions to Cosmos-SDK and related
repositories!

Contributing to this repo can mean many things such as participated in
discussion or proposing code changes. To ensure a smooth workflow for all
contributors, the general procedure for contributing has been established:

1. Either [open](https://github.com/cosmos/cosmos-sdk/issues/new/choose) or
   [find](https://github.com/cosmos/cosmos-sdk/issues) an issue you'd like to help with
2. Participate in thoughtful discussion on that issue
3. If you would like to contribute:
   1. If the issue is a proposal, ensure that the proposal has been accepted
   2. Ensure that nobody else has already begun working on this issue. If they have,
      make sure to contact them to collaborate
   3. If nobody has been assigned for the issue and you would like to work on it,
      make a comment on the issue to inform the community of your intentions
      to begin work
   4. Follow standard Github best practices: fork the repo, branch from the
      HEAD of `master`, make some commits, and submit a PR to `master`
      - For core developers working within the cosmos-sdk repo, to ensure a clear
        ownership of branches, branches must be named with the convention
        `{moniker}/{issue#}-branch-name`
   5. Be sure to submit the PR in `Draft` mode submit your PR early, even if
      it's incomplete as this indicates to the community you're working on
      something and allows them to provide comments early in the development process
   6. When the code is complete it can be marked `Ready for Review`
   7. Be sure to include a relevant change log entry in the `Unreleased` section
      of `CHANGELOG.md` (see file for log format)

Note that for very small or blatantly obvious problems (such as typos) it is
not required to an open issue to submit a PR, but be aware that for more complex
problems/features, if a PR is opened before an adequate design discussion has
taken place in a github issue, that PR runs a high likelihood of being rejected.

Take a peek at our [coding repo](https://github.com/tendermint/coding) for
overall information on repository workflow and standards. Note, we use `make tools` for installing the linting tools.

Other notes:

- Looking for a good place to start contributing? How about checking out some
  [good first issues](https://github.com/cosmos/cosmos-sdk/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
- Please make sure to run `make format` before every commit - the easiest way
  to do this is have your editor run it for you upon saving a file. Additionally
  please ensure that your code is lint compliant by running `golangci-lint run`.
  A convenience git `pre-commit` hook that runs the formatters automatically
  before each commit is available in the `contrib/githooks/` directory.

## Architecture Decision Records (ADR)

When proposing an architecture decision for the SDK, please create an [ADR](./docs/architecture/README.md)
so further discussions can be made. We are following this process so all involved parties are in
agreement before any party begins coding the proposed implementation. If you would like to see some examples
of how these are written refer to [Tendermint ADRs](https://github.com/tendermint/tendermint/tree/master/docs/architecture)

## Pull Requests

To accommodate review process we suggest that PRs are categorically broken up.
Ideally each PR addresses only a single issue. Additionally, as much as possible
code refactoring and cleanup should be submitted as a separate PRs from bugfixes/feature-additions.

### Process for reviewing PRs

All PRs require two Reviews before merge (except docs changes, or variable name-changes which only require one). When reviewing PRs please use the following review explanations:

- `LGTM` without an explicit approval means that the changes look good, but you haven't pulled down the code, run tests locally and thoroughly reviewed it.
- `Approval` through the GH UI means that you understand the code, documentation/spec is updated in the right places, you have pulled down and tested the code locally. In addition:
  - You must also think through anything which ought to be included but is not
  - You must think through whether any added code could be partially combined (DRYed) with existing code
  - You must think through any potential security issues or incentive-compatibility flaws introduced by the changes
  - Naming must be consistent with conventions and the rest of the codebase
  - Code must live in a reasonable location, considering dependency structures (e.g. not importing testing modules in production code, or including example code modules in production code).
  - if you approve of the PR, you are responsible for fixing any of the issues mentioned here and more
- If you sat down with the PR submitter and did a pairing review please note that in the `Approval`, or your PR comments.
- If you are only making "surface level" reviews, submit any notes as `Comments` without adding a review.

### Updating Documentation

If you open a PR on the Cosmos SDK, it is mandatory to update the relevant documentation in /docs.

- If your change relates to the core SDK (baseapp, store, ...), please update the `docs/basics/`, `docs/core/` and/or `docs/building-modules/` folders.
- If your changes relate to the core of the CLI (not specifically to module's CLI/Rest), please modify the `docs/run-node/` folder.
- If your changes relate to a module, please update the module's spec in `x/moduleName/docs/spec/`.

## Forking

Please note that Go requires code to live under absolute paths, which complicates forking.
While my fork lives at `https://github.com/rigeyrigerige/cosmos-sdk`,
the code should never exist at `$GOPATH/src/github.com/rigeyrigerige/cosmos-sdk`.
Instead, we use `git remote` to add the fork as a new remote for the original repo,
`$GOPATH/src/github.com/cosmos/cosmos-sdk`, and do all the work there.

For instance, to create a fork and work on a branch of it, I would:

- Create the fork on github, using the fork button.
- Go to the original repo checked out locally (i.e. `$GOPATH/src/github.com/cosmos/cosmos-sdk`)
- `git remote rename origin upstream`
- `git remote add origin git@github.com:rigeyrigerige/cosmos-sdk.git`

Now `origin` refers to my fork and `upstream` refers to the Cosmos-SDK version.
So I can `git push -u origin master` to update my fork, and make pull requests to Cosmos-SDK from there.
Of course, replace `rigeyrigerige` with your git handle.

To pull in updates from the origin repo, run

- `git fetch upstream`
- `git rebase upstream/master` (or whatever branch you want)

Please don't make Pull Requests from `master`.

## Dependencies

We use [Go 1.14 Modules](https://github.com/golang/go/wiki/Modules) to manage
dependency versions.

The master branch of every Cosmos repository should just build with `go get`,
which means they should be kept up-to-date with their dependencies, so we can
get away with telling people they can just `go get` our software.

Since some dependencies are not under our control, a third party may break our
build, in which case we can fall back on `go mod tidy -v`.

## Protobuf

We use [Protocol Buffers](https://developers.google.com/protocol-buffers) along with [gogoproto](https://github.com/gogo/protobuf) to generate code for use in Cosmos-SDK.

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

## Testing

All repos should be hooked up to [CircleCI](https://circleci.com/).

If they have `.go` files in the root directory, they will be automatically
tested by circle using `go test -v -race ./...`. If not, they will need a
`circle.yml`. Ideally, every repo has a `Makefile` that defines `make test` and
includes its continuous integration status using a badge in the `README.md`.

We expect tests to use `require` or `assert` rather than `t.Skip` or `t.Fail`,
unless there is a reason to do otherwise.
When testing a function under a variety of different inputs, we prefer to use
[table driven tests](https://github.com/golang/go/wiki/TableDrivenTests).
Table driven test error messages should follow the following format
`<desc>, tc #<index>, i #<index>`.
`<desc>` is an optional short description of whats failing, `tc` is the
index within the table of the testcase that is failing, and `i` is when there
is a loop, exactly which iteration of the loop failed.
The idea is you should be able to see the
error message and figure out exactly what failed.
Here is an example check:

```go
<some table>
for tcIndex, tc := range cases {
  <some code>
  for i := 0; i < tc.numTxsToTest; i++ {
      <some code>
			require.Equal(t, expectedTx[:32], calculatedTx[:32],
				"First 32 bytes of the txs differed. tc #%d, i #%d", tcIndex, i)
```

## Branching Model and Release

User-facing repos should adhere to the trunk based development branching model: https://trunkbaseddevelopment.com/.

Libraries need not follow the model strictly, but would be wise to.

The SDK utilizes [semantic versioning](https://semver.org/).

### PR Targeting

Ensure that you base and target your PR on the `master` branch.

All feature additions should be targeted against `master`. Bug fixes for an outstanding release candidate
should be targeted against the release candidate branch. Release candidate branches themselves should be the
only pull requests targeted directly against master.

### Development Procedure

- the latest state of development is on `master`
- `master` must never fail `make lint test test-race`
- `master` should not fail `make lint`
- no `--force` onto `master` (except when reverting a broken commit, which should seldom happen)
- create a development branch either on github.com/cosmos/cosmos-sdk, or your fork (using `git remote add origin`)
- before submitting a pull request, begin `git rebase` on top of `master`

### Pull Merge Procedure

- ensure pull branch is rebased on `master`
- run `make test` to ensure that all tests pass
- merge pull request

### Release Procedure

- Start on `master`
- Create the release candidate branch `rc/v*` (going forward known as **RC**)
  and ensure it's protected against pushing from anyone except the release
  manager/coordinator
  - **no PRs targeting this branch should be merged unless exceptional circumstances arise**
- On the `RC` branch, prepare a new version section in the `CHANGELOG.md`
  - All links must be link-ified: `$ python ./scripts/linkify_changelog.py CHANGELOG.md`
  - Copy the entries into a `RELEASE_CHANGELOG.md`, this is needed so the bot knows which entries to add to the release page on github.
- Kick off a large round of simulation testing (e.g. 400 seeds for 2k blocks)
- If errors are found during the simulation testing, commit the fixes to `master`
  and create a new `RC` branch (making sure to increment the `rcN`)
- After simulation has successfully completed, create the release branch
  (`release/vX.XX.X`) from the `RC` branch
- Create a PR to `master` to incorporate the `CHANGELOG.md` updates
- Tag the release (use `git tag -a`) and create a release in Github
- Delete the `RC` branches

### Point Release Procedure

At the moment, only a single major release will be supported, so all point releases will be based
off of that release.

In order to alleviate the burden for a single person to have to cherry-pick and handle merge conflicts
of all desired backporting PRs to a point release, we instead maintain a living backport branch, where
all desired features and bug fixes are merged into as separate PRs.

Example:

Current release is `v0.38.4`. We then maintain a (living) branch `sru/release/v0.38.N`, given N as
the next patch release number (currently `0.38.5`) for the `0.38` release series. As bugs are fixed
and PRs are merged into `master`, if a contributor wishes the PR to be released as SRU into the
`v0.38.N` point release, the contributor must:

1. Add `0.38.N-backport` label
2. Pull latest changes on the desired `sru/release/vX.X.N` branch
3. Create a 2nd PR merging the respective SRU PR into `sru/release/v0.38.N`
4. Update the PR's description and ensure it contains the following information:
   - **[Impact]** Explanation of how the bug affects users or developers.
   - **[Test Case]** section with detailed instructions on how to reproduce the bug.
   - **[Regression Potential]** section with a discussion how regressions are most likely to manifest, or might
     manifest even if it's unlikely, as a result of the change. **It is assumed that any SRU candidate PR is
     well-tested before it is merged in and has an overall low risk of regression**.

It is the PR's author's responsibility to fix merge conflicts, update changelog entries, and
ensure CI passes. If a PR originates from an external contributor, it may be a core team member's
responsibility to perform this process instead of the original author.
Lastly, it is core team's responsibility to ensure that the PR meets all the SRU criteria.

Finally, when a point release is ready to be made:

1. Create `release/v0.38.N` branch
2. Ensure changelog entries are verified
   1. Be sure changelog entries are added to `RELEASE_CHANGELOG.md`
3. Add release version date to the changelog
4. Push release branch along with the annotated tag: **git tag -a**
5. Create a PR into `master` containing ONLY `CHANGELOG.md` updates
   1. Do not push `RELEASE_CHANGELOG.md` to `master`

Note, although we aim to support only a single release at a time, the process stated above could be
used for multiple previous versions.

## Code Owner Membership

In the ethos of open source projects, and out of necessity to keep the code
alive, the core contributor team will strive to permit special repo privileges
to developers who show an aptitude towards developing with this code base.

Several different kinds of privileges may be granted however most common
privileges to be granted are merge rights to either part of, or the entirety of the
code base (through the github `CODEOWNERS` file). The on-boarding process for
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
means guaranteed by any quantifiable metric. It is a symbol of great trust of
the community of this project.


## Concept & Release Approval Process

The process for how Cosmos SDK maintainers take features and ADRs from concept to release
is broken up into three distinct stages: **Strategy Discovery**, **Concept Approval**, and 
**Implementation & Release Approval**


### Strategy Discovery

* Develop long term priorities, strategy and roadmap for the SDK
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

In absense of general consensus, decision making requires ⅔ vote from the three members
of the **Concept Approval Committee**.

**Committee Members**

* Core Members: **Aaron** (Regen), **Bez** (Fission), **Alessio** (AiB)
* Secondary pool of candidates to replace / substitute:
    * **Chris Goes** (IG), **Sunny** (Sikka)

**Committee Criteria**

Members must:

* Participate in all or almost all ADR discussions, both on Github as well as in bi-weekly Architecture Review
  meetings
* Be active contributors to the SDK, and furthermore should be continuously making substantial contributions
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
* ⅔ approval is required from the **primary repo maintainers** in `CODEOWNERS`
  * Secondary pool of candidates to replace / substitute are listed as **secondary repo maintainers** in `CODEOWNERS`
 
*Note: For any major or minor release series denoted as a "Stable Release" (e.g. v0.39 "Launchpad"), a separate release
committee is often established. Stable Releases, and their corresponding release committees are documented
separately in [STABLE_RELEASES.md](./STABLE_RELEASES.md)*
