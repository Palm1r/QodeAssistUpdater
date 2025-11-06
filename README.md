# QodeAssistUpdater

Utility for automatically updating the QodeAssist plugin for Qt Creator.

## Features

- ✅ Automatic Qt Creator version detection
- ✅ Installed plugin version check (via Qt Creator API or file search)
- ✅ Download latest plugin version from GitHub
- ✅ Automatic archive extraction and installation (ZIP and 7z - fully embedded in binary)
- ✅ Smart release matching for your Qt Creator version
- ✅ Configuration and path validation
- ✅ Download progress indicator
- ✅ ZipSlip vulnerability protection

## Requirements

- Go 1.21 or higher (only for building from source)
- For using pre-built binary: no additional dependencies required
- Support for `.zip` and `.7z` archives is embedded in the binary

## Installation

### From Source

```bash
go mod download
go build -o qodeassist-updater
```

### From GitHub Release

1. Go to the [releases page](https://github.com/Palm1r/QodeAssistUpdater/releases)
2. Download the binary for your platform
3. Make it executable (Unix systems): `chmod +x qodeassist-updater-*`
4. Move to a directory in your PATH or use directly

## Usage

1. Edit `config.yaml`:
   ```yaml
   qtcreator_path: "/usr/bin/qtcreator"  # Path to Qt Creator executable
   plugin_path: "/home/user/.local/share/QtProject/qtcreator/plugins"  # Plugin installation directory
   ```

2. Run:
```bash
./qodeassist-updater
```

Or with a custom config file:
```bash
./qodeassist-updater -config /path/to/config.yaml
```

## How It Works

1. **Qt Creator Version Detection**: The utility runs `qtcreator -version` and parses the version
2. **Installed Plugin Check**: 
   - Uses `qtcreator --version`, which outputs a list of all plugins with their versions
   - Parses the output and finds the QodeAssist plugin line (case-insensitive)
   - If version cannot be determined, attempts to find it in filenames in the plugin directory
3. **Release Search**: Searches for the latest release on GitHub and matches the appropriate file for your platform and Qt Creator version
4. **Download and Install**: Downloads the archive, automatically detects format (.zip or .7z) and extracts it to the plugin directory (all formats are supported natively, without external dependencies)

## GitHub Release Format

The utility expects the following release file name format:
```
PluginName-vX.Y.Z-QtCMM.mm.pp-Platform-Arch.ext
```

Where:
- `vX.Y.Z` - plugin version (e.g., `v0.8.1`)
- `QtCMM.mm.pp` - Qt Creator version (e.g., `QtC16.0.2`)
- `Platform` - platform: `Windows`, `Linux`, `macOS`
- `Arch` - architecture: `x64` for Linux/Windows, `universal` for macOS
- `ext` - file extension: `.zip` or `.7z` (both formats are supported natively)

Examples:
- `QodeAssist-v0.8.1-QtC16.0.2-Linux-x64.zip`
- `QodeAssist-v0.8.1-QtC18.0.0-macOS-universal.zip`
- `QodeAssist-v0.8.1-QtC16.0.2-Windows-x64.zip`
- `QodeAssist-v0.8.1-QtC16.0.2-Linux-x64.7z`

The utility first tries to find a file with an exact Qt Creator version match (e.g., `QtC16.0.2`), and if not found, searches by major.minor version (e.g., `QtC16.0`).

## Typical Plugin Installation Paths

- **Linux**: `~/.local/share/QtProject/qtcreator/plugins`
- **macOS**: `~/Library/Application Support/QtProject/qtcreator/plugins`
- **Windows**: `%APPDATA%\QtProject\qtcreator\plugins`

## Development

### GitHub Actions

The project uses GitHub Actions for automation:

- **CI** (`.github/workflows/ci.yml`):
  - Automatic build on commits to `main`
  - Automatic build on pull requests
  - Testing on different platforms (Linux, macOS, Windows)
  - Checks on different Go versions (1.21, 1.22, 1.23)

- **Release** (`.github/workflows/release.yml`):
  - Automatic release creation when a tag matching `v*` is created (e.g., `v1.0.0`)
  - Build binaries for all platforms:
    - Linux x64 and ARM64
    - macOS Intel and Apple Silicon
    - Windows x64
  - Automatic GitHub Release creation with binaries and checksums

### Creating a Release

To create a new release:

```bash
# 1. Update version in code (if needed)
# 2. Create and push a tag
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will automatically build binaries for all platforms and create a release.

## Notes

- Make sure Qt Creator is closed before updating the plugin
- After updating, restart Qt Creator to load the new plugin version
- If the utility cannot determine the installed plugin version, it assumes the plugin is not installed and will install the latest version
