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
# Preview using your browser
noteman preview
# Build site
noteman build
# Compress and publish to your server
noteman publish
```

## Minimal Example

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
