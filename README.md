# NoteMan

High level markdown notes management tool.


## What can it do?

1. Collect markdown files in diffrent locations
1. Find title / date and generate slug. All automatically. Supports chinese title slug translation.
1. Reformat these files and send to a target dir (such as `content` dir for hugo)
1. Build site and publish to your server.

## Usage

Configuration

```
code ~/.config/noteman/config.jsonc
```

Commands

```shell
# Sync notes to target dir (with preprocessing)
noteman sync
# Preview using your browser
noteman preview
# Build site
noteman build
# Compress and publish to your server
noteman publish
```


## Thanks to

[Lorem Markdownum](https://jaspervdj.be/lorem-markdownum/)

[JSON-to-Go: Convert JSON to Go instantly](https://mholt.github.io/json-to-go/)