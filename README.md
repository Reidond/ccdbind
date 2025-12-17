# quicksetd

This repo contains `ccd-gamed`, a systemd *user* daemon that:

- Detects CCD/L3 CPU groups from sysfs.
- When it sees a Steam/Proton game process, it pins user slices (default: `app.slice`, `background.slice`) to the OS CPUs.
- Moves game PIDs into a dedicated scope under `game.slice`, and pins that scope to the GAME CPUs.

## Build

```sh
go test ./...
go build ./cmd/ccd-gamed
```

## Install (user service)

```sh
install -Dm755 ./ccd-gamed ~/.local/bin/ccd-gamed
install -Dm644 systemd/user/ccd-gamed.service ~/.config/systemd/user/ccd-gamed.service
install -Dm644 systemd/user/game.slice ~/.config/systemd/user/game.slice
install -Dm644 ./config.example.toml ~/.config/ccd-gamed/config.toml

systemctl --user daemon-reload
systemctl --user enable --now ccd-gamed.service
```

## Config

- Config file path (default): `~/.config/ccd-gamed/config.toml`
- Optional ignore list: `~/.config/ccd-gamed/ignore.txt` (one executable basename per line, `#` comments allowed)
- State file (default): `~/.local/state/ccd-gamed/state.json`

Start from `config.example.toml`.

## CLI flags

- `--print-topology`: print detected `OS_CPUS`/`GAME_CPUS` and exit.
- `--dry-run`: log intended actions but don't mutate systemd state.
- `--dump-state`: print persisted state JSON and exit.
- `--config <path>`: config file.
- `--interval <dur>`: poll interval override (e.g. `1s`, `500ms`).

## D-Bus notes

`ccd-gamed` uses the systemd user manager D-Bus API on the user bus:

- `org.freedesktop.systemd1.Manager.StartTransientUnit` signature: `(s name, s mode, a(sv) properties, a(sa(sv)) aux)`
- `org.freedesktop.systemd1.Manager.AttachProcessesToUnit` signature: `(s unit, s subcgroup, au pids)`

In `godbus/dbus`, `a(sv)` can be passed as `[]struct{Name string; Value dbus.Variant}{ {Name: "Prop", Value: dbus.MakeVariant(value)} }`.
