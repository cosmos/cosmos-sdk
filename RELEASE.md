# Release Process

Basecoin is the heart of most demo apps and the testnets, but the last few releases have been a little chaotic.  In order to guarantee a higher, production-quality release in the future, we will work on a release process to check before the push to master.  This is a work-in-progress and should be trialed on the 0.6.x patches, and used for the 0.7.0 release.

This is a rough-guide.  Please add comments here, let's try it out for 0.6.1 and see what is annoying and useless, and what is missing and useful.

## Preparation

* Clarify scope of release
* Create issues
* Write code
* Update CHANGELOG
* Create PR for release
* Update version

## QA

Once we have a PR for the release and think it is ready, we should test it out internally:

* Code review (in addition to individual code reviews on the merged issue)
* Manual run-through of tutorials (and feedback on bad UX)
* Deployment of a private testnet, multiple users test out manually (feedback on bugs, or annoying UX)
* Test out upgrading existing testnet from last version, document or add tools for easier upgrade.
* If outstanding issues here, fix the issues, and repeat.

## Release

Once QA passes, we need to orchestrate the release.

* Merge to master
* Set all glide dependencies to proper master versions of repos
* Push code with new version tag
* Upgrade our public-facing testnets with the latest versions
