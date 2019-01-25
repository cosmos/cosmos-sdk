# Contributing

Thank you for considering making contributions to Cosmos-SDK and related
repositories!

Contributing to this repo can mean many things such as participated in
discussion or proposing code changes. To ensure a smooth workflow for all
contributors, the general procedure for contributing has been established:

  1. either [open](https://github.com/cosmos/cosmos-sdk/issues/new/choose) or
     [find](https://github.com/cosmos/cosmos-sdk/issues) an issue you'd like to help with,
  2. participate in thoughtful discussion on that issue,
  3. if you would then like to contribute code:
     1. if a the issue is a proposal, ensure that the proposal has been accepted,
     2. ensure that nobody else has already begun working on this issue, if they have
       make sure to contact them to collaborate,
     3. if nobody has been assigned the issue and you would like to work on it
       make a comment on the issue to inform the community of your intentions
       to begin work,
     4. follow standard github best practices: fork the repo, branch from the
       tip of `develop`, make some commits, and submit a PR to `develop`,
     5. include `WIP:` in the PR-title to and submit your PR early, even if it's
       incomplete, this indicates to the community you're working on something and
       allows them to provide comments early in the development process. When the code
       is complete it can be marked as ready-for-review by replacing `WIP:` with
       `R4R:` in the PR-title.

Note that for very small or blatantly obvious problems (such as typos) it is
not required to an open issue to submit a PR, but be aware that for more complex
problems/features, if a PR is opened before an adequate design discussion has
taken place in a github issue, that PR runs a high likelihood of being rejected.

Take a peek at our [coding repo](https://github.com/tendermint/coding) for
overall information on repository workflow and standards. Note, we use `make
get_dev_tools` and `make update_dev_tools` for installing the linting tools.

Other notes:
  - Looking for a good place to start contributing? How about checking out some
    [good first
    issues](https://github.com/cosmos/cosmos-sdk/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
  - Please make sure to use `gofmt` before every commit - the easiest way to do
    this is have your editor run it for you upon saving a file. Additionally
    please ensure that your code is lint compliant by running `make lint`

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

* If your change relates to the core SDK (baseapp, store, ...), please update the docs/gaia folder, the docs/examples folder and possibly the docs/spec folder.
* If your changes relate specifically to the gaia application (not including modules), please modify the docs/gaia folder.
* If your changes relate to a module, please update the module's spec in docs/spec. If the module is used by gaia, you might also need to modify docs/gaia and/or docs/examples.
* If your changes relate to the core of the CLI or Light-client (not specifically to module's CLI/Rest), please modify the docs/clients folder.

## Forking

Please note that Go requires code to live under absolute paths, which complicates forking.
While my fork lives at `https://github.com/rigeyrigerige/cosmos-sdk`,
the code should never exist at  `$GOPATH/src/github.com/rigeyrigerige/cosmos-sdk`.
Instead, we use `git remote` to add the fork as a new remote for the original repo,
`$GOPATH/src/github.com/cosmos/cosmos-sdk `, and do all the work there.

For instance, to create a fork and work on a branch of it, I would:

  - Create the fork on github, using the fork button.
  - Go to the original repo checked out locally (i.e. `$GOPATH/src/github.com/cosmos/cosmos-sdk`)
  - `git remote rename origin upstream`
  - `git remote add origin git@github.com:ebuchman/cosmos-sdk.git`

Now `origin` refers to my fork and `upstream` refers to the Cosmos-SDK version.
So I can `git push -u origin master` to update my fork, and make pull requests to Cosmos-SDK from there.
Of course, replace `ebuchman` with your git handle.

To pull in updates from the origin repo, run

  - `git fetch upstream`
  - `git rebase upstream/master` (or whatever branch you want)

Please don't make Pull Requests to `master`.

## Dependencies

We use [dep](https://github.com/golang/dep) to manage dependencies.

That said, the master branch of every Cosmos repository should just build
with `go get`, which means they should be kept up-to-date with their
dependencies so we can get away with telling people they can just `go get` our
software.

Since some dependencies are not under our control, a third party may break our
build, in which case we can fall back on `dep ensure` (or `make
get_vendor_deps`). Even for dependencies under our control, dep helps us to
keep multiple repos in sync as they evolve. Anything with an executable, such
as apps, tools, and the core, should use dep.

Run `dep status` to get a list of vendor dependencies that may not be
up-to-date.

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

```
<some table>
for tcIndex, tc := range cases {
  <some code>
  for i := 0; i < tc.numTxsToTest; i++ {
      <some code>
			require.Equal(t, expectedTx[:32], calculatedTx[:32],
				"First 32 bytes of the txs differed. tc #%d, i #%d", tcIndex, i)
 ```

## Branching Model and Release

User-facing repos should adhere to the branching model: http://nvie.com/posts/a-successful-git-branching-model/.
That is, these repos should be well versioned, and any merge to master requires a version bump and tagged release.

Libraries need not follow the model strictly, but would be wise to.

The SDK utilizes [semantic versioning](https://semver.org/).

### PR Targeting

Ensure that you base and target your PR on the correct branch:
  - `release/vxx.yy.zz` for a merge into a release candidate
  - `master` for a merge of a release
  - `develop` in the usual case

All feature additions should be targeted against `develop`. Bug fixes for an outstanding release candidate
should be targeted against the release candidate branch. Release candidate branches themselves should be the
only pull requests targeted directly against master.

### Development Procedure:
  - the latest state of development is on `develop`
  - `develop` must never fail `make test` or `make test_cli`
  - `develop` should not fail `make test_lint`
  - no --force onto `develop` (except when reverting a broken commit, which should seldom happen)
  - create a development branch either on github.com/cosmos/cosmos-sdk, or your fork (using `git remote add origin`)
  - before submitting a pull request, begin `git rebase` on top of `develop`

### Pull Merge Procedure:
  - ensure pull branch is rebased on develop
  - run `make test` and `make test_cli` to ensure that all tests pass
  - merge pull request
  - push master may request that pull requests be rebased on top of `unstable`

### Release Procedure:
  - start on `develop`
  - prepare changelog/release issue
  - bump versions
  - push to release-vX.X.X to run CI
  - merge to master
  - merge master back to develop

### Hotfix Procedure:
  - start on `master`
  - checkout a new branch named hotfix-vX.X.X
  - make the required changes
    - these changes should be small and an absolute necessity
    - add a note to CHANGELOG.md
  - bump versions
  - push to hotfix-vX.X.X to run the extended integration tests on the CI
  - merge hotfix-vX.X.X to master
  - merge hotfix-vX.X.X to develop
  - delete the hotfix-vX.X.X branch
