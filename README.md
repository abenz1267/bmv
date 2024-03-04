# BMV - Bulkmove

Wrapper around `mv` which allows bulk operations via stdin.

bmv will also create missing directories and cleanup directories that became empty in the process. Both are configurable via flags.

## Features

- bulk renaming via stdin
- rename file(s) in editor ($EDITOR)
- define processor for renaming, f.e. `sed`
- create missing directories on the fly
- delete directories that became empty after moving
- drop-in replacement for `mv`: args/flags will be passed to `mv`
- handles circular renaming

## Installation

```sh
# Arch
yay -S bmv-bin

# via Go
go install github.com/abenz126/bmv@latest
```

`mv` instance being used is `/usr/bin/mv`, unless `$BMV_MV` is specified.

## Usage

```sh
# normal 'mv' actions, simply passed to 'mv':
bmv oldfile newfile

# bmv specific:
<2 column output from external [src dest\n]> | bmv
fzf -m | bmv -e # defaults to $EDITOR
fzf -m | bmv -e=vim
ls | bmv -p sed 's/old/new/'
bmv -p sed 's/old/new/' # implicit call to 'ls'
bmv # same as 'ls | bmv -e'
```
