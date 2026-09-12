# NoteMan

High level markdown notes management tool.

## What can it do?

1. Collect markdown files in diffrent locations
1. Find title / date and generate slug. All automatically. Supports chinese title slug translation.
1. Reformat these files and send to a target dir (such as `content` dir for hugo)
1. Build site and publish to your server.

## Build

Run:

```bash
make
```

## Usage

Configuration

```bash
code ~/.config/noteman/config.jsonc
```

example:

```jsonc
{
    "source": {
        "directories": [
            "/home/pluveto/Workspace/notes/blog-src" // raw source markdown files dir
        ],
        "filters": {
            "regex_filter": {
                "exclude": [
                    ".*\\.pri.*" // exclude files with .pri. in name
                ]
            }
        }
    },
    "target": {
        "mapping": {
            // map source dir to target dir
            "/Users/zijingzhang/Repo/blogws/blog-src": "/Users/zijingzhang/Repo/blogws/blog/content/{{lang_prefix}}/posts"
        }
    },
    "build": {
        "command": "hugo",
        "args": [],
        "working_directory": "/home/pluveto/Workspace/notes/blog"
    },
    "preview": {
        "command": "hugo",
        "args": [
            "server"
        ],
        "working_directory": "/home/pluveto/Workspace/notes/blog"
    },
    "publish": {
        "artifacts": "/home/pluveto/Workspace/notes/blog/public",
        "service": {
            "name": "simple_http_upload",
            "params": {
                "api": "http://www.example.com/upload",
                "auth": "pluveto2xHHm0Z5BLb0M1GlBlpAGgfuxbqzSrDv"
            }
        },
        "preview_url": "https://www.example.com"
    }
}
```

Commands

```shell
# Sync notes to target dir (with preprocessing)
noteman sync
# Generate target files without writing metadata back to source notes
noteman sync --no-write-back
# Preview using your browser
noteman preview
# Build site
noteman build
# Compress and publish to your server
noteman publish
```

## Minimal Example

Math-enabled notes (`mathjax: true`) use the maintained
[`pluveto/goldmark-mathjax-fix`](https://github.com/pluveto/goldmark-mathjax-fix)
module, pinned to an immutable version in `go.mod`. Formatting normalizes CRLF
to LF without modifying the source body. Adjacent display blocks do not require
an extra blank line.

Sync validates and formats every note before writing any target or source file.
Invalid metadata, formatting errors, missing path mappings and duplicate targets
abort the operation with a nonzero exit status. Filesystem write errors are
reported immediately; writes are not a transaction across the entire directory.
For a preview, point the target mapping at a temporary directory and run
`noteman sync --no-write-back`.

Use noteman as a markdown file preprocessor.

```bash
mkdir input
mkdir output
vi input
---
Hello world

$$

E = mc^2 \\

F_G = G \frac{m_1 m_2}{r^2}

$$
```

```bash
./noteman sync
```

## Thanks to

[Lorem Markdownum](https://jaspervdj.be/lorem-markdownum/)

[JSON-to-Go: Convert JSON to Go instantly](https://mholt.github.io/json-to-go/)
