# Phasionary — Android client

A native Android (Kotlin + Jetpack Compose) client for the Phasionary
`/api/v1` JSON API. Thin client, online-only, meant to be reached over
Tailscale. It carries the TUI's dense, monospace, sharp-edged aesthetic into a
touch UI.

**Scope (this build):** project picker (most recently updated first),
category-grouped task list with server-shared collapse state, tap a status
marker to cycle status optimistically, tap a row to open a full-page editor
(title / description / status / priority / estimate, plus delete), and inline
capture rows for adding tasks and categories. Separators are editable too —
insert one below a task from the editor, tap it to relabel or delete it in
place. Swipe gestures and task reordering land in a later pass.

Everything is deliberately animation-free: navigation cuts between screens with
no transition, and Material's ripple is replaced by an instant flat press
highlight (`ui/theme/Indication.kt`), so the app feels as immediate as the TUI.

### API surface it uses

The client drives these `/api/v1` endpoints (all served by `phasionary serve`):

| Method | Path | Used for |
| --- | --- | --- |
| `GET` | `/projects` | picker; sorted by `updated_at`, newest first |
| `GET` | `/projects/{pid}` | the task list, and the editor's task lookup |
| `GET` `PUT` | `/projects/{pid}/folds` | collapsed categories, shared with the TUI |
| `POST` | `/projects/{pid}/categories` | add category |
| `POST` | `/projects/{pid}/categories/{cid}/tasks` | inline task capture; `kind: "separator"` + `insert_after` for a divider |
| `PATCH` | `/projects/{pid}/categories/{cid}/tasks/{tid}` | editor save (only changed fields) |
| `DELETE` | `/projects/{pid}/categories/{cid}/tasks/{tid}` | editor delete |
| `POST` | `/projects/{pid}/categories/{cid}/tasks/{tid}/status` | status marker tap |

Folds are stored in the server's `state.json`, the same file the TUI reads, so
collapsing a category on the phone collapses it in the TUI as well — the TUI
picks it up the next time it opens the project.

**Separators** are `Task`s with `kind: "separator"`, and the API enforces what
they are: a divider carries a label (which may be empty — that renders as a
plain rule) and nothing else. A `PATCH` or a status change that would put a
status, priority, estimate or description on one is rejected with a 400 rather
than silently turning it into a half-task. The tallies skip them on both ends.

## Toolchain

This project builds the same way as the sibling `kilorep/android`: a
**Nix-patched Android SDK** (prebuilt Google binaries don't run on NixOS) and
the **committed Gradle wrapper**. Verified building with:

- **Nix** (flakes enabled) — provides the SDK via `nix/sdk.nix` (platform +
  build-tools **35**).
- **JDK 17+** — JDK 21 works (system `java`).
- **`just`** — the task runner (recipes in `justfile`).
- **`adb`** — to install onto a phone.
- Toolchain versions (all in `gradle/libs.versions.toml`): AGP 8.7.3, Kotlin
  2.1.0, Compose BOM 2024.12.01, Gradle 8.11.1, `minSdk 26` / `compileSdk 35`.

No Android Studio required (though `File ▸ Open` on this folder works too).

## Setup (once)

```bash
cd android
just sdk        # builds the Nix Android SDK, writes local.properties → sdk.dir
```

`just sdk` is idempotent; the SDK derivation is shared with `kilorep/android`,
so if you built it there it's already in the Nix store (no rebuild).

## Run on your phone

Plug in a phone with USB debugging on (`adb devices` should list it), then:

```bash
just run        # installDebug + adb reverse + launch on the phone
```

`just run` also runs `adb reverse tcp:7777 tcp:7777`, which tunnels the phone's
`localhost:7777` to your laptop. That makes the tightest dev loop: run the
server on the laptop and point the app at `localhost` (see below). Other
recipes:

```bash
just install    # build + install the debug APK, no launch
just apk        # just build the APK → app/build/outputs/apk/debug/app-debug.apk
just logs       # tail logcat for this app only
just test       # JVM unit tests (no device needed)
```

## Connect the app to the server

`phasionary serve` generates a token on first run, saves it to `config.json` and
prints it once; pass `--token` only if you want to supply your own. Set **Base
URL** + **Token** in the app's **⚙ Settings**. Two ways to reach it:

- **Desk loop (via `just run`'s `adb reverse`):** run the server bound to
  localhost, and use `localhost` in the app:

  ```bash
  phasionary serve --host 127.0.0.1 --port 7777 --token <TOKEN>
  ```
  App Base URL → `http://localhost:7777`.

- **Real use (over Tailscale):** bind to the tailnet address and use the
  tailnet IP / MagicDNS name in the app:

  ```bash
  phasionary serve --host 0.0.0.0 --port 7777
  # or PHASIONARY_SERVE_TOKEN=<TOKEN> phasionary serve ...
  ```
  App Base URL → `http://100.x.y.z:7777`.

  A tailnet **IP** works as-is. To use a **MagicDNS name** instead, permit it —
  the server accepts IP addresses and `localhost` by default and refuses other
  names, which is what stops a web page from reaching it by re-pointing its own
  domain at your machine:

  ```bash
  phasionary serve --host 0.0.0.0 --allowed-host phasionary.tail1a2b.ts.net
  ```

Include `http://` in the Base URL. The app allows cleartext HTTP
(`usesCleartextTraffic=true`) because the server speaks plain HTTP over a
private network.

## Testing

Unit tests run on the JVM (no device/emulator) — `just test` (or
`./gradlew :app:testDebugUnitTest`). They cover the three seams:

- **Serialization** (`ModelSerializationTest`) — wire JSON ⇆ models, omitted /
  unknown fields, request-body omission.
- **Repository** (`RemotePhasionaryRepositoryTest`) — Ktor `MockEngine`: URL +
  bearer header, success parsing, status-code → typed-error mapping, the PATCH
  body's omitted-vs-explicitly-empty distinction, and the 204 on delete.
- **ViewModels** (`TaskListViewModelTest`, `TaskEditViewModelTest`) — load,
  optimistic status cycle and fold write with revert-on-failure, inline add,
  and the editor's dirty tracking / partial-save diff.

## Release APK

Every GitHub release (tag push) attaches a signed `phasionary-android.apk`,
built by the `build-android` job in `.github/workflows/release.yml`. How it
fits together:

- **Version:** `versionName`/`versionCode` derive from the repo-root `VERSION`
  file (the same value as the release tag), so upgrades always install in
  place. Nothing to bump here.
- **Signing:** `android/release.keystore` + `android/keystore.properties`,
  both gitignored. CI reconstructs them from the repo secrets
  `ANDROID_KEYSTORE_BASE64` and `ANDROID_KEYSTORE_PASSWORD` (alias
  `phasionary`).
- **Back up the keystore.** GitHub secrets are write-only; the local
  `release.keystore` + `keystore.properties` are the only recoverable copy.
  Losing them means future releases get a new signature and phones refuse the
  upgrade until the app is uninstalled.
- **Local signed build:** with those two files present, `just release-apk`
  produces the same signed APK. Without them the release build is unsigned
  (debug builds are unaffected).

## Layout

```
app/src/main/java/com/phasionary/app/
  data/
    model/        @Serializable wire models + request bodies (mirror the Go domain)
    net/          Ktor client config (shared by app + tests) + JSON
    settings/     ServerConfig + DataStore-backed SettingsRepository
    repo/         PhasionaryRepository interface + Remote (Ktor) impl
    ApiException  typed failures
  di/             AppContainer (manual DI, no Hilt in v1)
  ui/
    theme/        TUI design system: color tokens, monospace type, flat theme,
                  FlatIndication (the no-animation press effect)
    components/    TaskRow, StatusChip, CategoryHeader, InlineInputRow, state views
    projects/     project picker screen + ViewModel
    tasks/        task list + task editor screens & ViewModels
    settings/     settings screen + ViewModel
    navigation/    NavHost + routes (transitions disabled)
  MainActivity, PhasionaryApp
nix/sdk.nix       Nix-patched Android SDK definition
justfile          task runner (sdk / run / install / apk / logs / test)
```

## Notes / v1 tradeoffs

- **Token storage:** plaintext DataStore — acceptable for a single-user, private
  (Tailscale-only) setup. Move to an encrypted store before any wider use.
  Backups are disabled (`allowBackup="false"` plus `data_extraction_rules.xml`)
  so that plaintext token is never copied off the device — neither to the user's
  Drive by Auto Backup, nor via `adb backup` on the Android 8–11 devices
  `minSdk 26` still supports. Nothing here is worth restoring anyway: the app's
  whole persisted state is a URL and a token.
- **Task order:** rendered as the server returns it (matches the TUI); no
  client-side re-sort. Projects are sorted server-side by `updated_at` for this
  client only — the TUI picker keeps its own manual order.
- **Freshness:** the task list re-reads the project when the screen resumes.
  There is no live push, so changes made in the TUI while the list is already on
  screen appear on the next resume, not instantly.
- **Font:** system monospace (`FontFamily.Monospace`). Drop a JetBrains Mono
  `.ttf` into `app/src/main/res/font` and point `ui/theme/Type.kt` at it to
  match the desktop TUI exactly.
- **Offline:** none yet. The `PhasionaryRepository` interface is the seam where
  a local cache / Syncthing path can be added later without touching the UI.
