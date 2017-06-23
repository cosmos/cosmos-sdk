# Release Process

Basecoin is the heart of most demo apps and the testnets, but the last few releases have been a little chaotic.  In order to guarantee a higher, production-quality release in the future, we will work on a release process to check before the push to master.  This is a work-in-progress and should be trialed on the 0.6.x patches, and used for the 0.7.0 release.

This is a rough-guide.  Please add comments here, let's try it out for 0.6.1 and see what is annoying and useless, and what is missing and useful.

## Planning

* Create issues (and invite others to do so)
* Create WIP PR for release as placeholder
  * Clarify scope of release in text
* Create labels, eg. (0.6.1 and 0.6.x)
* Tag all issues for this release with 0.6.1
  * Other, less urgent enhancements should get the 0.6.x label

## Coding

* Freeze tagging more issues for this release
  * Update PR to note this
  * If you want an exception, you need a good excuse ;)
* Handle all issues
  * Write code
  * Update CHANGELOG
  * Review and merge
* Update version
* Remove WIP flag on PR
* Organize QA
* Prepare blog post (optional for patch/bugfix releases?)

## QA

Once we have a PR for the release and think it is ready, we should test it out internally:

* Code review
  * Hopefully dealt with by individual code reviews on the merged issues
  * A general run-through is always good to find dead-code, things to cleanup
* Review blog post (and run-through)
* Manual run-through of tutorials (and feedback on bad UX)
* Deployment of a private testnet, multiple users test out manually (feedback on bugs, or annoying UX)
* Test out upgrading existing testnet from last version, document or add tools for easier upgrade.
* If problems arrise here:
  * Create bugfix issues
  * Fix them
  * Repeat QA

## Release

Once QA passes, we need to orchestrate the release.

* Merge to master
* Set all glide dependencies to proper master versions of repos
* Push code with new version tag
* Link CHANGELOG to the [github release](https://github.com/tendermint/basecoin/releases)
* Package up new version as binaries (and upload to s3)
* Upgrade our public-facing testnets with the latest versions
* Release blog post
