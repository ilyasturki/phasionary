# Sync architecture (A2): readable files locally, sync as an app feature

Status: **accepted** — decided 2026-08-29. This document is the in-tree record
of the decision and the design the mobile/web slice builds on.

## The decision

Phasionary gains mobile and web clients backed by an **optional self-hosted
server**, while the terminal app stays purely local. Chosen architecture, out
of four candidates:

**A2 — JSON files locally, app-level sync, SQLite only in the server.**

- Every client owns a full local store and works with zero connectivity.
  The TUI/CLI store stays what it is today: readable JSON files.
- Sync is a feature of the app, not of the filesystem: each device keeps a
  small **change journal** and exchanges journal entries with the server over
  HTTP when one is configured and reachable.
- The server is a single Go binary on the user's own machine. It holds its
  state in SQLite, merges changes per-item, serves the JSON API for the phone,
  and serves the web app itself.
- No server configured → the app is exactly the local, single-device tool it
  is today. Nothing degrades.

### Rejected alternatives, and why

| Option | Shape | Why not |
| --- | --- | --- |
| A1 — SQLite everywhere | Same sync design, binary local store | Gives up greppable, hand-editable, git-diffable data files for robustness this data size doesn't need. A2 → A1 stays a contained swap behind `data.ProjectRepository` if scale ever demands it. |
| B — merge-safe event-log files | Per-device append-only logs, any file syncer is the pipe | The most engineering on the table (event engine, compaction, causal ordering, snapshots — on every platform) to serve users who run a file syncer but won't run one small server. The server must exist anyway for phone/web. |
| C — CRDT storage (Automerge / cr-sqlite) | Library-merged replicas over any transport | Heavyweight young dependency in every client (Rust bindings in Go, WASM on web), binary library-versioned on-disk format. Overkill when "conflict" means one user racing themself. |
| Server as source of truth | All clients online-first | Violates the fixed rule below. |
| Git as the pipe | Commit/merge on every save | Unusable from a phone. |
| Google Tasks / CalDAV backend | Sync into an existing ecosystem | Phases, categories, separators, priorities and estimates don't map; the domain would be flattened into someone else's model. |
| Hosted sync (Turso, PowerSync, Firebase…) | Third-party cloud | Against the self-hosted, offline-by-default ethos. |

## Fixed rules (hold in every future slice)

1. **The TUI and CLI are local-only first.** No network, no daemon, no server
   required, ever. A server is an addition, never a dependency.
2. **Quick capture works offline everywhere.** A phone with no signal still
   captures; entries queue and sync later.
3. **Single user, self-hosted.** No accounts, no multi-tenancy. Server auth is
   a bearer credential per device.
4. **One pipe per data directory.** When app-level sync is enabled, file-level
   syncers (Syncthing, Dropbox, …) must stop carrying the data directory —
   they demote to backup. Two pipes moving the same data would fight; the app
   should detect tell-tale syncer droppings (`*.sync-conflict-*` files) and
   warn.

## Current on-disk layout (schema v1)

What exists after this slice, all under `~/.local/share/phasionary/` (or
`PHASIONARY_DATA_PATH`):

```
projects/<project-id>.json    one project per file, schema-versioned
projects/<id>.json.lock       flock files (advisory, content ignored)
projects/.global.lock         serializes cross-project operations
state.json                    UI state: folds, cursors, project order, dir links
```

Config is separate, `~/.config/phasionary/config.json` (or
`PHASIONARY_CONFIG_PATH`), mode 0600, per-device preferences only.

Properties already in place that sync builds on:

- **Stable random IDs** on projects, categories and tasks (`domain.NewID`),
  never reused, safe as filenames (`domain.ValidateID`).
- **Atomic, fsync'd writes** (`writeFileAtomic`): readers see the old file or
  the new one, never a torn one.
- **Advisory flocks** serialize concurrent local writers (TUI + CLI).
- **Schema versioning** (added this slice): every save stamps
  `"schema": <domain.ProjectSchemaVersion>`; a missing field reads as v1; a
  file declaring a *newer* schema is refused with an "upgrade phasionary"
  error rather than rewritten through an older format — that is what protects
  a half-upgraded fleet.
- **Write-boundary text validation** rejects control characters once, for
  every writer (`domain.ValidateLine` / `ValidateMultiline`).

## Sync design (implemented in later slices)

### Device identity (implemented, slice 2)

Each device that syncs gets a random, immutable device ID, minted at
enrollment (`phasionary sync login <url>`). It lives in
**`$XDG_STATE_HOME/phasionary/device.json`** (`~/.local/state/phasionary/`,
override `PHASIONARY_STATE_PATH`), deliberately **not** in the data or config
directories: both of those are historically carried by file syncers, and a
device ID that syncs to another machine is two devices claiming one identity.
Copying a data directory to a new machine is therefore safe by construction —
the new machine has no state dir, so it enrolls as a new device.

The sync credential (server URL + per-device bearer token, granted at
enrollment) lives in the same 0600 file, in a 0700 state dir.

### Change journal (implemented, slice 2)

A per-device, purely local outbox — bookkeeping, not history
(`internal/journal`). Also in `$XDG_STATE_HOME/phasionary/`
(`journal.jsonl` + `journal.lock` + `journal.head`), never in the synced
data dir.

- **Ops are derived by diffing at the save boundary**, not by instrumenting
  every mutation: one hook (`data.ChangeRecorder`) covers every writer — the
  TUI, the CLI, import.
- Each op is one JSONL line: `journal.Op`, kinds in the `Kind*` constants.
  Update ops carry only changed fields, with cleared values present as
  explicit empties.
- **Order is a field of the container**: whenever a list's ID sequence
  changes, a single `*.reorder` op carries the full new sequence; the server
  never infers order from creates.
- `seq` is a per-device monotonic counter, and never regresses across a
  prune: `journal.head` records the acked high-water mark.
  Cross-device ordering uses the server's arrival order plus per-field
  last-writer-wins with the op timestamp as tiebreaker (single user racing
  themself: LWW per field is the right amount of machinery).
- **Deletes are tombstone ops.** State files just remove the entity; the
  journal entry is what tells the server "deleted", so a delete can never be
  resurrected by a stale peer pushing an old snapshot. A whole-project delete
  is one tombstone; the server cascades it to its children.
- Entries are **pruned after server acknowledgment** (`PruneThrough`). The
  recorder checks for `device.json` on every write rather than once at
  startup, so a TUI that was open when `sync login` ran journals from its
  next save on; while unenrolled it journals nothing and the journal never
  grows. Enrolling starts from a full snapshot upload (exactly the
  diff-from-nothing cascade).

### Protocol (implemented, slice 3)

Wire types live in `internal/syncproto`; the client side in
`internal/syncclient`.

- `POST /v1/enroll`: the device mints its own ID, the server mints the bearer
  token and stores only its hash.
- `POST /v1/sync`: the journal up, snapshots down.

Push and pull happen in one round-trip, and **the server is the only place
merges happen**. A device never applies ops: the response carries a full
snapshot of every project whose server version moved past the device's
cursor — its own push included, so a device sees its edits merged with
everyone else's in the same round — and the device overwrites its files with
those snapshots. Applying ops client-side would not converge on a same-field
conflict without per-field timestamps in the JSON files; snapshots converge
by construction.

Two rules make a round safe to interrupt (`syncclient.Run`): a snapshot is
skipped while the journal still holds entries for that project, and the
device cursor advances only after every snapshot has been applied.

### Stale saves (implemented, slice 3)

The TUI edits in memory and saves whole snapshots, so a project changed on
disk by a pull (or by the CLI) would be silently overwritten by its next
save — and with sync on, the overwrite would journal the reversal. The store
therefore refuses a snapshot save over a file version this process did not
load. Read-modify-write under the flock (`WithProjectLocked`) is never stale;
listing projects deliberately does not move the baseline.

### Server (implemented, slice 3)

`phasionary-server`, a separate binary (`cmd/phasionary-server`,
`internal/server`) so the local-only TUI never links a server or a SQLite
driver. Pure-Go SQLite (`modernc.org/sqlite`) keeps the arm64 release build
toolchain-free.

- **State model:** one `entities` table; every project, category and task is
  a row of merged fields keyed by the domain JSON tags plus a timestamp per
  field. Order lists and a task's `category_id` are fields like any other,
  so every op kind is the same operation: set fields, or tombstone.
- **Merge:** per field, last-writer-wins on the op timestamp, later arrival
  winning a tie. A tombstone, once set, wins over anything after it. Creates
  are upserts, so enrolling with a copy of an already-synced data directory
  converges. An order list that lost the race drops the other side's
  additions from the sequence, not from existence: the snapshot appends
  every live entity the winning list misses, by `created_at`. A task whose
  category was deleted goes with it.
- **Pull:** each project row carries the server cursor at the last request
  that touched it; a pull returns projects with version > the device cursor.
- **Auth:** per-device bearer token minted at enrollment, stored hashed.
  Enrollment codes are minted by `phasionary-server enroll` into the same
  database (WAL, so it runs beside the service), single use, ten-minute
  life; a failed attempt costs a second. Retries after a lost response are
  idempotent through the per-device acked seq.
- Plain HTTP; deployment guidance is private network or TLS reverse proxy.
  The NixOS module (`nix/module.nix`, `services.phasionary-server`) runs it
  hardened under its own user. The web app's static files and any read API
  arrive with the web slice.

### Mobile & web app (next slice)

One TypeScript codebase, built twice: a static web app served by the server,
and the same app wrapped with **Capacitor** for Android (iOS later for free).
Framework: pick in that slice; the constraints that matter are a local store
(IndexedDB/SQLite via Capacitor plugin) implementing the same journal +
snapshot model, and an offline-first capture path. The TUI's dense monospace
aesthetic is the design language.

## Slice plan from here

1. **(done)** Android app, `internal/api`, `phasionary serve`,
   serve token and NixOS serve module removed; project files schema-versioned;
   this design recorded.
2. **(done)** **Journal + device identity** in the Go core
   (`internal/journal`, `data.ChangeRecorder`, `sync status`).
3. **(done)** **Server**, client sync commands, the stale-save guard, NixOS
   module.
4. **Web app** against the server, then **Capacitor** wrap for Android.
   Candidates for the same slice: automatic sync from the TUI (on start and
   after saves, best effort) and a Host-header allowlist on the server once
   a browser is a client.
