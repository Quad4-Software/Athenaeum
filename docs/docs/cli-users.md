---
sidebar_position: 11
title: CLI users
description: Manage Athenaeum accounts from the command line without the HTTP server.
---

# CLI users

`athenaeum users` edits the database under `--data` (or `ATHENAEUM_DATA`)
without starting the HTTP server. Useful for recovery, bootstrap, and
automation.

```sh
./bin/athenaeum users help
./bin/athenaeum users list --data ./data
./bin/athenaeum users add alice --password 'secretpass'
./bin/athenaeum users set-admin bob
./bin/athenaeum users set-admin bob --no-admin
./bin/athenaeum users set-permissions alice --permissions read,edit_metadata
./bin/athenaeum users reset-password alice
./bin/athenaeum users rename alice alice2
./bin/athenaeum users show alice
./bin/athenaeum users delete alice
```

## Commands

| Command | Description |
| ------- | ----------- |
| `list` / `ls` | List users |
| `add` / `create <username>` | Create a local user |
| `reset-password` / `passwd` | Reset password |
| `rename <user> <name>` | Rename a user |
| `set-admin` / `admin` | Grant admin (`--no-admin` to revoke) |
| `set-permissions` | Set non-admin permission names |
| `show` / `get` | Show one user |
| `delete` / `rm` | Delete a user |

## Global flags

| Flag | Description |
| ---- | ----------- |
| `--data <dir>` | Data directory (default `./data` or `ATHENAEUM_DATA`) |
| `--color <mode>` | `auto`, `always`, or `never` |
| `--no-color` | Disable ANSI color |

## Passwords

Password sources, in order: `--password`, `ATHENAEUM_PASSWORD`, or stdin.
Strength rules follow the same policy flags as the server
([Configuration](./configuration)).

Permission names match [Authentication](./authentication): `read`,
`edit_metadata`, `delete_books`, `manage_library`, `manage_users`.
