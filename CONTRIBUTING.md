# Contributing

Thank you for considering making contributions to Cosmos-SDK and related repositories! Start by taking a look at this [coding repo](https://github.com/tendermint/coding) for overall information on repository workflow and standards.

Please follow standard github best practices: fork the repo, branch from the tip of develop, make some commits, and submit a pull request to develop. See the [open issues](https://github.com/cosmos/cosmos-sdk/issues) for things we need help with!

Please make sure to use `gofmt` before every commit - the easiest way to do this is have your editor run it for you upon saving a file. Additionally please ensure that your code is lint compliant by running `make lint`

Looking for a good place to start contributing? How about checking out some [good first issues](https://github.com/cosmos/cosmos-sdk/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)

## Pull Requests

To accommodate review process we suggest that PRs are catagorically broken up. 
Ideally each PR addresses only a single issue. Additionally, as much as possible
code refactoring and cleanup should be submitted as a seperate PRs from bugfixes/feature-additions. 

## Forking

Please note that Go requires code to live under absolute paths, which complicates forking.
While my fork lives at `https://github.com/rigeyrigerige/cosmos-sdk`,
the code should never exist at  `$GOPATH/src/github.com/rigeyrigerige/cosmos-sdk`.
Instead, we use `git remote` to add the fork as a new remote for the original repo,
`$GOPATH/src/github.com/cosmos/cosmos-sdk `, and do all the work there.

For instance, to create a fork and work on a branch of it, I would:

  * Create the fork on github, using the fork button.
  * Go to the original repo checked out locally (i.e. `$GOPATH/src/github.com/cosmos/cosmos-sdk`)
  * `git remote rename origin upstream`
  * `git remote add origin git@github.com:ebuchman/basecoin.git`

Now `origin` refers to my fork and `upstream` refers to the Cosmos-SDK version.
So I can `git push -u origin master` to update my fork, and make pull requests to Cosmos-SDK from there.
Of course, replace `ebuchman` with your git handle.

To pull in updates from the origin repo, run

    * `git fetch upstream`
    * `git rebase upstream/master` (or whatever branch you want)

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

## Branching Model and Release

User-facing repos should adhere to the branching model: http://nvie.com/posts/a-successful-git-branching-model/.
That is, these repos should be well versioned, and any merge to master requires a version bump and tagged release.

Libraries need not follow the model strictly, but would be wise to.

The SDK utilizes [semantic versioning](https://semver.org/).

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
