<!--
order: 3
-->

# The keyring

This document describes how to configure and use the keyring and its various backends for an [**application**](../basics/app-anatomy.md). A separate document for implementing a CLI for an SDK [**module**](../building-modules/intro.md) can be found [here](#../building-modules/module-interfaces.md#cli). {synopsis}

Starting with the v0.38.0 release, Cosmos SDK comes with a new keyring implementation
that provides a set of commands to manage cryptographic keys in a secure fashion. The
new keyring supports multiple storage backends, some of which may not be available on
all operating systems.

## The `os` backend

The `os` backend relies on operating system-specific defaults to handle key storage
securely. Typically, operating systems credentials sub-systems handle passwords prompt,
private keys storage, and user sessions according to their users password policies. Here
is a list of the most popular operating systems and their respective passwords manager:

* macOS (since Mac OS 8.6): [Keychain](https://support.apple.com/en-gb/guide/keychain-access/welcome/mac)
* Windows: [Credentials Management API](https://docs.microsoft.com/en-us/windows/win32/secauthn/credentials-management)
* GNU/Linux:
  * [libsecret](https://gitlab.gnome.org/GNOME/libsecret)
  * [kwallet](https://api.kde.org/frameworks/kwallet/html/index.html)

GNU/Linux distributions that use GNOME as default desktop environment typically come with
[Seahorse](https://wiki.gnome.org/Apps/Seahorse). Users of KDE based distributions are
commonly provided with [KDE Wallet Manager](https://userbase.kde.org/KDE_Wallet_Manager).
Whilst the former is in fact a `libsecret` convenient frontend, the former is a `kwallet`
client.

`os` is the default option since operating system's default credentials managers are
designed to meet users' most common needs and provide them with a comfortable
experience without compromising on security.

## The `file` backend

The `file` backend more closely resembles the keybase implementation used prior to
v0.38.1. It stores the keyring encrypted within the apps configuration directory. This
keyring will request a password each time it is accessed, which may occur multiple
times in a single command resulting in repeated password prompts. If using bash scripts
to execute commands using the `file` option you may want to utilize the following format
for multiple prompts:

```sh
# assuming that KEYPASSWD is set in the environment
$ gaiacli config keyring-backend file                             # use file backend
$ (echo $KEYPASSWD; echo $KEYPASSWD) | gaiacli keys add me        # multiple prompts
$ echo $KEYPASSWD | gaiacli keys show me                          # single prompt
```

::: tip
The first time you add a key to an empty keyring, you will be prompted to type the password twice.
:::

## The `pass` backend

The `pass` backend uses the [pass](https://www.passwordstore.org/) utility to manage on-disk
encryption of keys' sensitive data and metadata. Keys are stored inside `gpg` encrypted files
within app-specific directories. `pass` is available for the most popular UNIX
operating systems as well as GNU/Linux distributions. Please refer to its manual page for
information on how to download and install it.

::: tip
**pass** uses [GnuPG](https://gnupg.org/) for encryption. `gpg` automatically invokes the `gpg-agent`
daemon upon execution, which handles the caching of GnuPG credentials. Please refer to `gpg-agent`
man page for more information on how to configure cache parameters such as credentials TTL and
passphrase expiration.
:::

The password store must be set up prior to first use:

```sh
$ pass init <GPG_KEY_ID>
```

Replace `<GPG_KEY_ID>` with your GPG key ID. You can use your personal GPG key or an alternative
one you may want to use specifically to encrypt the password store.

## The `test` backend

The `test` backend is a password-less variation of the `file` backend. Keys are stored
unencrypted on disk. This backend is meant for testing purposes only and **should never be used
in production environments**.

## The `kwallet` backend

The `kwallet` backend uses `KDE Wallet Manager`, which comes installed by default on the
GNU/Linux distributions that ships KDE as default desktop environment. Please refer to
[KWallet Handbook](https://docs.kde.org/stable5/en/kdeutils/kwallet5/index.html) for more
information.
