# Bulk-Rename files/folders

Simple cli-tool to bulk rename files. Piped from stdin.

## Install

Arch: `yay -S bmv-bin`

## Usage Example

```
donttouchme.txt // will be ignored, no destination.
somefile.txt 1.txt // will be moved
someotherfile.txt newfolder/1.txt // will be moved, creating directory if needed
moveme.txt 1.txt // error, file '1.txt' already exists
idontexist.txt bla.txt // error, file doesn't exist
```

## Tip

You can easily bulk rename files with vim/nvim this way. Simply do f.e. `ls | nvim -` and you'll get a buffer with cwd's content. You can of course also use ":r!ls" inside vim to fill the buffer with the output of the command.

To pipe the buffers content into `bmv` just do `:w !bmv`. Done.
