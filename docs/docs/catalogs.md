---
sidebar_position: 13
title: OPDS and KOSync
description: Connect e-readers with OPDS 1.2, OPDS 2, comics/Kindle feeds, and KOReader sync.
---

# OPDS and KOSync

When users exist, catalog endpoints require authentication (HTTP Basic with
the same credentials as the web UI, or an API key).

## OPDS 1.2

| Path | Description |
| ---- | ----------- |
| `/opds/` | Navigation catalog |
| `/opds/recent` | Recent acquisitions |
| `/opds/series` | Series navigation |
| `/opds/series/{name}` | Books in one series |
| `/opds/search?q=` | Search feed |
| `/opds/comics` | Comics-oriented feed |
| `/opds/kindle` | Kindle-friendly feed |

Feeds include cover thumbnails, per-user library filtering, and reading
progress in entry summaries where available. Point KOReader, Pluto, or similar
clients at `https://your-host/opds/`.

## OPDS 2

JSON catalogs:

| Path | Description |
| ---- | ----------- |
| `/opds/v2/` | OPDS 2 root |
| `/opds/v2/recent` | Recent acquisitions (OPDS 2) |

## KOSync (KOReader)

KOReader progress sync:

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | `/kosync/users/auth` | Authenticate |
| GET | `/kosync/syncs/progress/{document}` | Get progress |
| PUT | `/kosync/syncs/progress` | Save progress |

Use the same username/password (or API key where supported by the client) as
your Athenaeum account. Document IDs follow KOReader's KOSync conventions.
