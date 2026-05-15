# QodeAssist Updater

Utility for installing and updating the QodeAssist plugin for Qt Creator with checksum verification.

## Installation

```bash
# From source
go mod download
go build -o qodeassist-updater

# Or download pre-built binary
# https://github.com/Palm1r/QodeAssistUpdater/releases
```

## Usage

The updater is fully driven by command-line flags. There is no configuration file — you
explicitly pass the Qt Creator version and the directory where the plugin should be
installed.

```bash
./qodeassist-updater --version        # Show version information
./qodeassist-updater --help           # Show help
./qodeassist-updater --list-versions  # List all available versions (>= 0.5.9)
```

### Options

| Option                   | Description                                                            |
| ------------------------ | ---------------------------------------------------------------------- |
| `--qtc-version <ver>`    | Qt Creator version, required for install/update (e.g. `19.0.1`)        |
| `--plugin-dir <path>`    | Directory to install into / remove from (required, except list)       |
| `--plugin-version <ver>` | Install a specific plugin version instead of the latest                |
| `--checksum <hash>`      | Expected SHA256 of the downloaded archive (optional)                   |
| `--wait-pid <pid>`       | Wait for this process to exit before touching files                   |
| `--wait-timeout <sec>`   | Timeout for `--wait-pid` (default `120`, `0` = infinite)               |
| `--log-file <path>`      | Mirror all output into a log file (color codes stripped)               |
| `--relaunch <path>`      | Launch this application after a successful install/update              |
| `-y`, `--yes`            | Assume "yes" for all prompts (non-interactive mode)                    |

### Install

```bash
./qodeassist-updater --install --qtc-version 19.0.1 --plugin-dir ~/QtPlugins
```

- `--qtc-version` — your Qt Creator version, used to pick the matching release asset.
- `--plugin-dir` — directory the plugin archive is extracted into (created if missing).

### Update

`--update` works like `--install`, but removes old QodeAssist files from the target
directory before extracting the new ones:

```bash
./qodeassist-updater --update --qtc-version 19.0.1 --plugin-dir ~/QtPlugins
./qodeassist-updater --update --qtc-version 19.0.1 --plugin-dir ~/QtPlugins --yes
```

### Remove

```bash
./qodeassist-updater --remove --plugin-dir ~/QtPlugins
./qodeassist-updater --remove --plugin-dir ~/QtPlugins --yes
```

### Install a specific version

```bash
./qodeassist-updater --install --qtc-version 19.0.1 --plugin-dir ~/QtPlugins --plugin-version 0.8.1
# Or with 'v' prefix:
./qodeassist-updater --install --qtc-version 19.0.1 --plugin-dir ~/QtPlugins --plugin-version v0.8.1
```

### Checksum verification (optional)

```bash
./qodeassist-updater --install --qtc-version 19.0.1 --plugin-dir ~/QtPlugins --checksum abc123...
```

### Non-interactive mode

For automated scripts and CI/CD pipelines, use `--yes` / `-y` to skip confirmation prompts:

```bash
./qodeassist-updater --update --qtc-version 19.0.1 --plugin-dir ~/QtPlugins --yes
./qodeassist-updater --remove --plugin-dir ~/QtPlugins -y
```

### Detached update (calling from Qt Creator)

The plugin library is locked while Qt Creator is running, so an in-app updater
must spawn `qodeassist-updater` as a detached process, then quit. These options
make that flow safe:

- `--wait-pid <pid>` — wait for that process (the running Qt Creator) to fully
  exit before touching any files.
- `--wait-timeout <sec>` — give up waiting after N seconds (default `120`,
  `0` = wait forever).
- `--log-file <path>` — mirror all output into a file, since a detached process
  has no visible console. Color codes are stripped from the file.
- `--relaunch <path>` — start the given application again after a successful
  install/update.

```bash
qodeassist-updater --update --qtc-version 19.0.1 --plugin-dir ~/QtPlugins \
    --yes \
    --wait-pid 12345 \
    --log-file ~/qodeassist-update.log \
    --relaunch "/Applications/Qt Creator.app"
```

Always pass `--yes` in this mode — a detached process cannot answer prompts.
The exit code is `0` on success and `1` on failure; check `--log-file` for
details since stdout/stderr are not visible.

### Typical plugin directories

- **Linux**: `~/.local/share/QtProject/qtcreator/plugins/<qtc_version>/petrmironychev.qodeassist`
- **macOS**: `~/Library/Application Support/QtProject/Qt Creator/plugins/<qtc_version>/petrmironychev.qodeassist`
- **Windows**: `%LOCALAPPDATA%\QtProject\qtcreator\plugins\<qtc_version>\petrmironychev.qodeassist`

## Notes

- Close Qt Creator before updating
- Restart Qt Creator after installation/update

## License

GPLv3
