Install
=======

Cosmos SDK can be installed to
``$GOPATH/src/github.com/cosmos/cosmos-sdk`` like a normal Go program:

::

    go get github.com/cosmos/cosmos-sdk

If the dependencies have been updated with breaking changes, or if
another branch is required, ``dep`` is used for dependency management.
Thus, assuming you've already run ``go get`` or otherwise cloned the
repo, the correct way to install is:

::

    cd $GOPATH/src/github.com/cosmos/cosmos-sdk
    make get_vendor_deps
    make install
    make install_examples

This will install ``gaiad`` and ``gaiacli`` and four example binaries:
``basecoind``, ``basecli``, ``democoind``, and ``democli``.

Verify that everything is OK by running:

::

    gaiad version

you should see:

::

    0.15.0-rc1-9d90c6b

then with:

::

    gaiacli version

you should see:

::

    0.15.0-rc1-9d90c6b
