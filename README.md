# QodeAssist Updater

Utility for automatically updating the QodeAssist plugin for Qt Creator with automatic version detection and checksum verification.

## Installation

```bash
# From source
go mod download
go build -o qodeassist-updater

# Or download pre-built binary
# https://github.com/Palm1r/QodeAssistUpdater/releases
```

## Usage

```bash
./qodeassist-updater --version        # Show version information
./qodeassist-updater --status         # Check status and available updates
./qodeassist-updater --install        # Install the plugin (latest version)
./qodeassist-updater --update         # Update to latest version
./qodeassist-updater --remove         # Remove the plugin
./qodeassist-updater --list-versions  # List all available versions (>= 0.5.9)
./qodeassist-updater --help           # Show help
```

### List available versions

You can list all available plugin versions starting from 0.5.9:

```bash
./qodeassist-updater --list-versions
```

### Install specific version

You can install or update to a specific plugin version:

```bash
./qodeassist-updater --install --plugin-version 1.2.3
./qodeassist-updater --update --plugin-version 1.2.3
# Or with 'v' prefix:
./qodeassist-updater --install --plugin-version v1.2.3
```

### Checksum verification (optional)

```bash
./qodeassist-updater --install --checksum abc123...
./qodeassist-updater --update --checksum abc123...
# With specific version:
./qodeassist-updater --install --plugin-version 1.2.3 --checksum abc123...
```

### Custom config path

```bash
./qodeassist-updater --config /path/to/config.yaml --update
```

## Configuration

On first run, `config.yaml` is created:

```yaml
qtcreator_path: "/path/to/Qt Creator"
plugin_path: "/path/to/plugins/{qtc_version}/petrmironychev.qodeassist"
```

**Important**: Use the `{qtc_version}` variable in `plugin_path` — it will be automatically replaced with your Qt Creator version.

### Typical plugin paths

- **Linux**: `~/.local/share/QtProject/qtcreator/plugins/{qtc_version}/petrmironychev.qodeassist`
- **macOS**: `~/Library/Application Support/QtProject/Qt Creator/plugins/{qtc_version}/petrmironychev.qodeassist`
- **Windows**: `%APPDATA%\QtProject\qtcreator\plugins\{qtc_version}\petrmironychev.qodeassist`

## Notes

- Close Qt Creator before updating
- Restart Qt Creator after installation/update

## License

GPLv3
