# fileorbit

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A simple web application for sharing and storing files online. Single binary, no CGO, SQLite-backed, with built-in analytics.

## Features

## Getting Started

### Prerequisites

- Go 1.25.2+

### Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/jere-mie/fileorbit.git
   cd fileorbit
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Create your config:
   ```sh
   cp example.env .env
   ```
   Edit `.env` to set your admin password, port, and other options.

4. Run the application:
   ```sh
   go run .
   ```
5. Visit [http://localhost:8080](http://localhost:8080) (or whatever port you specified in `.env`).

### CLI Commands

```sh
# Run database migrations manually (also runs automatically on startup)
./fileorbit migrate

# Print the current version
./fileorbit version
```

### Building

Build for your current platform:

```sh
go build -o bin/fileorbit .
```

Cross-compile for all supported platforms:

```powershell
# PowerShell
./scripts/build.ps1
```

```sh
# Bash
./scripts/build.sh
```

Binaries are output to the `bin/` directory.

### Releasing

The release scripts use the [GitHub CLI](https://cli.github.com/) to create a GitHub release from the version in `version.txt` and upload all binaries from `bin/`:

```powershell
# PowerShell
./scripts/release.ps1
```

```sh
# Bash
./scripts/release.sh
```

### Development with Air

You can use [Air](https://github.com/air-verse/air) for live reloading during development:

```sh
go install github.com/air-verse/air@latest
air
```

## Configuration

All configuration is done via environment variables (or a `.env` file):

| Variable | Default | Description |
|---|---|---|
| `ADMIN_PASSWORD` | `admin` | Password for the admin dashboard |
| `PORT` | `8080` | Server port |
| `HOST` | `localhost` | Server bind address |
| `DATABASE_PATH` | `openly.db` | Path to SQLite database file |

## License

MIT - see [LICENSE](LICENSE) for details.

## Download a Release Binary

You can download a prebuilt binary directly from GitHub Releases without cloning the repo.

### Linux (amd64)

```sh
curl -Lo fileorbit https://github.com/jere-mie/fileorbit/releases/latest/download/fileorbit_linux_amd64
chmod +x fileorbit
```

### Linux (arm64)

```sh
curl -Lo fileorbit https://github.com/jere-mie/fileorbit/releases/latest/download/fileorbit_linux_arm64
chmod +x fileorbit
```

### macOS (Apple Silicon)

```sh
curl -Lo fileorbit https://github.com/jere-mie/fileorbit/releases/latest/download/fileorbit_darwin_arm64
chmod +x fileorbit
```

### macOS (Intel)

```sh
curl -Lo fileorbit https://github.com/jere-mie/fileorbit/releases/latest/download/fileorbit_darwin_amd64
chmod +x fileorbit
```

### Windows (PowerShell)

```powershell
Invoke-WebRequest -Uri "https://github.com/jere-mie/fileorbit/releases/latest/download/fileorbit_windows_amd64.exe" -OutFile "fileorbit.exe"
```

### Available Binaries

| Platform | Architecture | Filename |
|---|---|---|
| Linux | amd64 | `fileorbit_linux_amd64` |
| Linux | 386 | `fileorbit_linux_386` |
| Linux | arm64 | `fileorbit_linux_arm64` |
| Linux | arm | `fileorbit_linux_arm` |
| macOS | amd64 | `fileorbit_darwin_amd64` |
| macOS | arm64 | `fileorbit_darwin_arm64` |
| Windows | amd64 | `fileorbit_windows_amd64.exe` |
| Windows | 386 | `fileorbit_windows_386.exe` |
| Windows | arm64 | `fileorbit_windows_arm64.exe` |
