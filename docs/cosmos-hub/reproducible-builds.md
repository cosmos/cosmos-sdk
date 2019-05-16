## Build Gaia Deterministically

Gitian is the deterministic build process that is used to build the Gaia executables. It provides a way to be reasonably sure that the executables are really built from the git source. It also makes sure that the same, tested dependencies are used and statically built into the executable.

Multiple developers build the source code by following a specific descriptor ("recipe"), cryptographically sign the result, and upload the resulting signature. These results are compared and only if they match, the build is accepted and provided for download.

More independent Gitian builders are needed, which is why this guide exists. It is preferred you follow these steps yourself instead of using someone else's VM image to avoid 'contaminating' the build.

This page contains all instructions required to build and sign reproducible Gaia binaries for Linux, Mac OS X, and Windows.

## Prerequisites

Make sure your system satisfy minimum requisites as outlined in https://github.com/devrandom/gitian-builder#prerequisites.

All the following instructions have been tested on *Ubuntu 18.04.2 LTS* with *docker 18.06.1-ce* and *docker 18.09.6-ce*.

If you are on Mac OS X, make sure you have prepended your `PATH` environment variable with GNU coreutils's path before running the build script:

```
export PATH=/usr/local/opt/coreutils/libexec/gnubin/:$PATH
```

## Build and sign

Clone cosmos-sdk:

```
git clone git@github.com:cosmos/cosmos-sdk
```

Checkout the commit, branch, or release tag you want to build:

```
cd cosmos-sdk/
git checkout v0.35.0
```

Run the following command to launch a build for `linux` and sign the final build
report (replace `user@example.com` with the GPG identity you want to sign the report with):

```
./cmd/gaia/contrib/gitian-build.sh -s user@example.com linux
```

The above command generates two directories in the current working directory:
* `gitian-build-linux` containing the `gitian-builder` clone used to drive the build process.
* `gaia.sigs` containing the signed build report.

Replace `linux` in the above command with `darwin` or `windows` to run builds for Mac OS X and Windows respectively.
Run the following command to build binaries for all platforms (`darwin`, `linux`, and `windows`):

```
cd cosmos-sdk/
for platform in darwin linux windows; do ./cmd/gaia/contrib/gitian-build.sh -s user@example.com $platform `pwd`; done
```

If you want to generate unsigned builds, just remove the option `-s` from the command line:

```
./cmd/gaia/contrib/gitian-build.sh linux
```

At the end of the procedure, build results can be found in the `./gaia.sigs` directory:

Please refer to the `cmd/gaia/contrib/gitian-build.sh`'s help screen for further information on its usage.

## Signatures upload

Once signatures are generated, they could be uploaded to gaia's dedicated repository: https://github.com/cosmos/gaia.sigs.

## Troubleshooting

### Docker gitian-target container cannot be killed

The issue is due to a relatively recent kernel apparmor change, [see here](https://github.com/moby/moby/issues/36809#issuecomment-379325713) for more information on a potential mitigation for the issue.

On Ubuntu 18.04, when the container hangs and `docker` is unable to kill it you can try to use `pkill` to forcibly terminate `containerd-shim`:

```
sudo pkill containerd-shim
docker system prune
```
