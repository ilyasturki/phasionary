# Phasionary — Android client

A native Android (Kotlin + Jetpack Compose) client for the Phasionary
`/api/v1` JSON API. Thin client, online-only, meant to be reached over
Tailscale. It carries the TUI's dense, monospace, sharp-edged aesthetic into a
touch UI.

**v1 scope (this build):** the read path + status changes — project picker,
category-grouped collapsible task list, tap a row to reveal its description, tap
a status chip to cycle status (todo → in progress → completed → cancelled →
todo, applied optimistically). Quick-capture (FAB) and swipe gestures land in
the next pass.

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

Start the server with a token, then set **Base URL** + **Token** in the app's
**⚙ Settings**. Two ways to reach it:

- **Desk loop (via `just run`'s `adb reverse`):** run the server bound to
  localhost, and use `localhost` in the app:

  ```bash
  phasionary serve --host 127.0.0.1 --port 7777 --token <TOKEN>
  ```
  App Base URL → `http://localhost:7777`.

- **Real use (over Tailscale):** bind to the tailnet address and use the
  tailnet IP / MagicDNS name in the app:

  ```bash
  phasionary serve --host 0.0.0.0 --port 7777 --token <TOKEN>
  # or PHASIONARY_SERVE_TOKEN=<TOKEN> phasionary serve ...
  ```
  App Base URL → `http://100.x.y.z:7777`.

Include `http://` in the Base URL. The app allows cleartext HTTP
(`usesCleartextTraffic=true`) because the server speaks plain HTTP over a
private network.

## Testing

Unit tests run on the JVM (no device/emulator) — `just test` (or
`./gradlew :app:testDebugUnitTest`). They cover the three seams:

- **Serialization** (`ModelSerializationTest`) — wire JSON ⇆ models, omitted /
  unknown fields, request-body omission.
- **Repository** (`RemotePhasionaryRepositoryTest`) — Ktor `MockEngine`: URL +
  bearer header, success parsing, and status-code → typed-error mapping.
- **ViewModel** (`TaskListViewModelTest`) — load, optimistic status cycle,
  revert-on-failure, category collapse.

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
    theme/        TUI design system: color tokens, monospace type, flat theme
    components/    TaskRow, StatusChip, CategoryHeader, state views
    projects/     project picker screen + ViewModel
    tasks/        task list screen + ViewModel
    settings/     settings screen + ViewModel
    navigation/    NavHost + routes
  MainActivity, PhasionaryApp
nix/sdk.nix       Nix-patched Android SDK definition
justfile          task runner (sdk / run / install / apk / logs / test)
```

## Notes / v1 tradeoffs

- **Token storage:** plaintext DataStore — acceptable for a single-user, private
  (Tailscale-only) setup. Move to an encrypted store before any wider use.
- **Task order:** rendered as the server returns it (matches the web UI); no
  client-side re-sort in v1.
- **Font:** system monospace (`FontFamily.Monospace`). Drop a JetBrains Mono
  `.ttf` into `app/src/main/res/font` and point `ui/theme/Type.kt` at it to
  match the desktop TUI exactly.
- **Offline:** none yet. The `PhasionaryRepository` interface is the seam where
  a local cache / Syncthing path can be added later without touching the UI.
