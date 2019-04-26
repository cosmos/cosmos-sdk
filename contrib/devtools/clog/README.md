# sdkch
Simple tool to maintain modular changelogs

# Usage

```
$ sdkch -help
usage: sdkch [-d directory] [-prune] command

Maintain unreleased changelog entries in a modular fashion.

Commands:
    add section stanza [message]    Add an entry file.
                                    If message is empty, start the editor to edit
                                    the message.
    generate [version]              Generate a changelog in Markdown format and print
                                    it to STDOUT. version defaults to UNRELEASED.

    Sections             Stanzas
         ---                 ---
    breaking                gaia
    features             gaiacli
improvements            gaiarest
    bugfixes                 sdk
                      tendermint

Flags:
  -d string
    	entry files directory (default "/home/alessio/work/tendermint/src/github.com/cosmos/cosmos-sdk/.pending")
  -prune
    	prune old entries after changelog generation
```

## Add a new entry

You can either drop a text file in the appropriate directory or use the `add` command:

```bash
$ sdkch add features gaiacli '#3452 New cool gaiacli command'
```

If no message is provided, a new entry file is opened in an editor is started

## Generate the full changelog

```bash
$ sdkch generate v0.30.0
```

The `-prune` flag would remove the old entry files after the changelog is generated.
