# Phasionary

A terminal-first project planner for people who'd rather press `j` than reach for a mouse. Single binary, plain JSON on disk, no accounts, no network.

![Phasionary demo](assets/demo.gif)

## Install

### Arch Linux (AUR)

```bash
yay -S phasionary       # build from source
yay -S phasionary-bin   # prebuilt binary
```

### Nix

Run without installing:

```bash
nix run github:ilyasturki/phasionary
```

Or pin it in your system flake:

```nix
inputs.phasionary.url = "github:ilyasturki/phasionary";
# environment.systemPackages = [ inputs.phasionary.packages.x86_64-linux.default ];
```

### Prebuilt binary

Grab the latest `phasionary-linux-x64` or `phasionary-linux-arm64` from the [releases page](https://github.com/ilyasturki/phasionary/releases), `chmod +x`, drop in `$PATH`.

### From source

```bash
go build -o phasionary ./cmd/phasionary
```

## Quick start

```bash
phasionary                # Launch TUI — opens the last project, or the picker
phasionary -p "Life"      # Jump straight into a project by name
```

First run seeds a default project. Inside the TUI: `a` adds a task, `Space` cycles status, `i` opens task info, `?` shows every keybinding.

## CLI

Every TUI action has a CLI counterpart, useful for scripts and quick edits. All commands accept `-j` for JSON output.

```bash
phasionary tasks -s todo -C Feature                   # Filter tasks
phasionary task add -C Feature "Build widget"         # Capture
phasionary task status <id> in_progress               # Triage
phasionary export -f json -o project.json             # Snapshot
phasionary import project.md                          # Restore from markdown
```

Run `phasionary --help` for the full surface (projects, categories, config, completions).

## Mobile

Every release also ships `phasionary-android.apk` — a native Android client, sideloaded from the [releases page](https://github.com/ilyasturki/phasionary/releases). It's a thin, online-only client for the `/api/v1` JSON API that `phasionary serve` exposes; point it at your machine over a private network (Tailscale) and the same projects, categories and fold state show up on your phone.

`phasionary serve` requires a bearer token on every request. The first run generates one, saves it to `config.json` and prints it once:

```
phasionary serve                    # prints the generated token
```

Enter that token in the app's settings alongside the server URL. To use your own instead, pass `--token` or set `PHASIONARY_SERVE_TOKEN`.

The server binds to `127.0.0.1:7777` by default; reach it from your phone over Tailscale (`--host 0.0.0.0`) or an SSH tunnel (`ssh -L 7777:localhost:7777 your-server`). Requests are accepted when the `Host` header is an IP address or `localhost`, which covers both routes — if you reach the server by a name instead, such as a Tailscale MagicDNS name, permit it with `--allowed-host`:

```
phasionary serve --host 0.0.0.0 --allowed-host phasionary.tail1a2b.ts.net
```

Authentication is not optional, even on loopback: binding to `127.0.0.1` does not keep a web page out, because a page the browser loads can re-point its own domain at the loopback address and reach the server. The token and the `Host` check are what actually close that off.

> **Upgrading from a version without tokens:** the phone will get `401`s until you enter the token that `phasionary serve` prints on its next run.

Setup, the API surface it uses, and build instructions: [`android/README.md`](android/README.md).

### Running it as a NixOS service

The flake ships `nixosModules.phasionary`. `tokenFile` is required — without it the server generates a token and prints it, which under systemd means the credential lands in the journal and stays there.

```nix
services.phasionary.serve = {
  enable = true;
  tokenFile = config.age.secrets.phasionary-token.path;  # or sops-nix, etc.
};
```

That binds `127.0.0.1:7777` and needs nothing else on a machine you reach by SSH tunnel. To reach it by a **name** — a domain, a Tailscale MagicDNS name — put TLS in front and list the name, or every request is refused with `421`:

```nix
services.phasionary.serve = {
  enable = true;
  tokenFile = config.age.secrets.phasionary-token.path;
  allowedHosts = [ "phasionary.example.com" ];
};

services.nginx.virtualHosts."phasionary.example.com" = {
  forceSSL = true;
  enableACME = true;
  locations."/".proxyPass = "http://127.0.0.1:7777";
};
```

Keep the bind on loopback and `openFirewall = false` in this shape: nginx reaches the service without crossing the firewall, and the token only stays secret because TLS terminates upstream — it is a bearer credential sent in the clear over plain HTTP.

Exposing the API to the open internet is worth thinking twice about. It is a single-user tool: one shared token, no per-client revocation, and no rate limiting on the token comparison. A private network (Tailscale, WireGuard, an SSH tunnel) fits it better than a public domain.

## Configuration

Config lives at `~/.config/phasionary/config.json`. Override with `phasionary config set <key> <value>` or edit directly.

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `status_display` | `text`, `icons` | `icons` | How task status renders in the TUI |
| `default_project` | project UUID | (none) | Project to open on launch |
| `priority_color` | `full`, `icon`, `none` | `full` | How priority is colored — full row, icon only, or off |
| `serve_token` | token string | (generated) | Bearer token `phasionary serve` requires; created on first serve |

Because it holds `serve_token`, `config.json` is kept at mode `0600` and is tightened automatically if it was created by an earlier version.

Paths can be relocated with `PHASIONARY_CONFIG_PATH` and `PHASIONARY_DATA_PATH` — both name a directory; `PHASIONARY_CONFIG_PATH` also accepts a path ending in `config.json` for convenience.

## Data

One project per JSON file under `~/.local/share/phasionary/projects/{uuid}.json`. UI state (fold state, cursor position, last project per directory) sits next to it in `state.json`. Every change is written synchronously — no undo, but nothing is ever in-flight.

## License

[MIT](LICENSE)
