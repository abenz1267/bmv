# BMV - Bulkmove

Wrapper around `mv` which allows bulk operations via stdin.

bmv will also create missing directories and cleanup directories that became empty in the process. Both are configurable via flags.

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
ls | bmv -e
ls | bmv sed 's/old/new/'
bmv sed 's/old/new/' [implies 'ls']
bmv # ls in $EDITOR
```
