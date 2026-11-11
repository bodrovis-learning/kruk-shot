# Kruk Shot

Kruk Shot is a lightweight, portable screenshot utility for Windows, written in Go. It uses DXGI Desktop Duplication and Direct3D 11 to capture screenshots of the desktop and games, including exclusive fullscreen applications.

> If you're looking for a cross-platform app, check out [kruk_screenshots](https://github.com/bodrovis-learning/kruk_screenshots).

## Features

- Global screenshot hotkey: **Ctrl+Shift+F9**.
- Support for multiple monitors.
- PNG screenshots saved in the `screens` folder next to the executable.
- Automatic recovery when switching display modes or resolutions.
- Dynamic detection of connected and disconnected monitors.
- No installation, administrator privileges, or CGO required.

## Usage

1. Download the appropriate Windows archive (AMD64 or ARM64) from [GitHub Releases](https://github.com/bodrovis-learning/kruk-shot/releases).
2. Extract the archive to a writable directory.
3. Run `kruk-shot.exe`.
4. Press **Ctrl+Shift+F9** to capture all active displays.

Each screenshot is saved as a separate PNG file with a timestamp. Keep the application running to capture more screenshots, or press **Ctrl+C** in its console to exit.

## Build from source

Requires Go 1.27.3 or later.

```bash
go build -o kruk-shot.exe .
```

Windows resources, including the application icon and DPI awareness manifest, are embedded during compilation.