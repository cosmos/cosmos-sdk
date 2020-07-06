# Stable Releases

The following release series are currently supported and receieve bugfixes:

* **0.37**
* **0.39 "Launchpad"**

Stable releases continue to receieve bugfixes until they reach End Of Life.
The `0.37.x` release series will continue receiving bugfixes until the Cosmos Hub
migrates to a newer release of the Cosmos-SDK.

The `0.38.x` release series will reach End of Life once the `0.39.0` release is published.

## Point Releases

Once a Cosmos-SDK release has been completed and published, updates for it are released under certain circumstances
and must follow the [Point Release Procedure](CONTRIBUTING.md).

## Rationale

Unlike in-development `master` branch snapshots, **Cosmos-SDK** releases are subject to much wider adoption,
and by a significantly different demographic of users. During development, changes in the `master` branch 
affect SDK users, application developers, early adopters, and other advanced users that elect to use
unstable experimental software at their own risk.

Conversely, users of a stable release expect a high degree of stability. They build their applications on it, and the
problems they experience with it could be potentially highly disruptive to their projects.

Stable release updates are recommended to the vast majority of developers, and so it is crucial to treat them
with great caution. Hence, when updates are proposed, they must be accompanied by a strong rationale and present
a low risk of regressions, i.e. even one-line changes could cause unexpected regressions due to side effects or
poorly tested code. We never assume that any change, no matter how little or non-intrusive, is completely exempt
of regression risks.

Therefore, the requirements for stable changes are different than those that are candidates to be merged in
the `master` branch. When preparing future major releases, our aim is to design the most elegant, user-friendly and
maintainable SDK possible which often entails fundamental changes to the SDK's architecture design, rearranging and/or
renaming packages as well as reducing code duplication so that we maintain common functions and data structures in one
place rather than leaving them scattered all over the code base. However, once a release is published, the
priority is to minimise the risk caused by changes that are not strictly required to fix qualifying bugs; this tends to
be correlated with minimising the size of such changes. As such, the same bug may need to be fixed in different
ways in stable releases and `master` branch.

### What qualifies as a Stable Release Update (SRU)

* **High-impact bugs**
  * Bugs that may directly cause a security vulnerability.
  * *Severe regressions* from a Cosmos-SDK's previous release. This includes all sort of issues
    that may cause the core packages or the `x/` modules unusable.
  * Bugs that may cause **loss of user's data**.
* Other safe cases:
  * Bugs which don't fit in the aforementioned categories for which an obvious safe patch is known.
  * Relatively small yet strictly non-breaking changes that introduce forward-compatible client
    features to smoothen the migration to successive releases.

### What does not qualify as SRU

* State machine changes.
* New features that introduces API breakages (e.g. public functions removal/renaming).
* Cosmetic fixes, such as formatting or linter warning fixes.

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
phone/video meeting. Subsequently at the end of the meeting, one of the existing
code owners should open a PR modifying the `CODEOWNERS` file. The other code
owners should then all approve this PR to publicly display their support.

Only if unanimous consensus is reached among all the existing code-owners will
an invitation be extended to a new potential-member. Likewise, when an existing
member is suggested to be removed/or have their privileges reduced, the member
in question must agree on the decision for their removal or else no action
should be taken. If however, a code-owner is verifiably shown to intentionally
have had acted maliciously or grossly negligent, code-owner privileges may be
stripped with no prior warning or consent from the member in question.

Earning this privilege should be considered to be no small feat and is by no
means guaranteed by any quantifiable metric. It is a symbol of great trust of
the community of this project.

# 0.37

The `0.37.x` release series will continue receiving bugfixes until the Cosmos Hub
migrates to a newer release of the Cosmos-SDK.

# 0.38

Discontinued since **0.39.0** is published.

# 0.39 - The Launchpad Release Series

The **0.39 Launchpad** release series will be supported until 6 months after **0.40.0** is published. A fairly strict **bugfix-only** rule applies to pull requests that are requested to be included into a stable point-release.

## What pull requests will be included in stable point-releases

Pull requests that fix bugs that fall in the following categories do not require a **Stable Release Exception** to be granted to be included in a stable point-release:

 * **Severe regressions**.
 * Bugs that may cause **client applications** to be **largely unusable**.
 * Bugs that may cause **state corruption or data loss**.
 * Bugs that may directly or indirectly cause a **security vulnerability**.

## What pull requests will NOT be automatically included in stable point-releases

As rule of thumb, the following changes will **NOT** be automatically accepted into stable point-releases:

 * **State machine changes**.
 * **Client application's code-breaking changes**, i.e. changes that prevent client applications to build without modifications to the client's source code.
 
 In some circumstances, PRs that don't meet the aforementioned criteria might be raised and asked to be granted a *Stable Release Exception*.
 
 ## Stable Release Exception
 
 How to ask for a **Stable Release Exception**.
 
 TODO
 
