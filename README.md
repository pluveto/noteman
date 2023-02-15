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
            "/home/pluveto/Workspace/notes/blog-src": "/home/pluveto/Workspace/notes/blog/content/posts"
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

## Thanks to

[Lorem Markdownum](https://jaspervdj.be/lorem-markdownum/)

[JSON-to-Go: Convert JSON to Go instantly](https://mholt.github.io/json-to-go/)
