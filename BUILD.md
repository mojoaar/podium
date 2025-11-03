# Building Podium

This guide explains how to build Podium for different platforms.

## Quick Start

### Using Makefile (Recommended)

```bash
make build      # Build for current platform
make build-all  # Build for all platforms
make clean      # Clean build artifacts
```

### Using Build Scripts

**Linux/macOS:**

```bash
./build.sh              # Build for current platform
./build.sh all          # Build for all platforms
./build.sh linux        # Build for Linux only
./build.sh macos        # Build for macOS only
./build.sh windows      # Build for Windows only
./build.sh clean        # Clean build artifacts
```

**Windows:**

```cmd
build.bat              REM Build for current platform
build.bat all          REM Build for all platforms
build.bat clean        REM Clean build artifacts
```

## Build Output

All binaries are created in the `bin/` directory:

```
bin/
├── podium                    # Current platform
├── podium-linux-amd64        # Linux 64-bit (Intel/AMD)
├── podium-linux-arm64        # Linux ARM64
├── podium-darwin-amd64       # macOS Intel
├── podium-darwin-arm64       # macOS Apple Silicon (M1/M2/M3)
└── podium-windows-amd64.exe  # Windows 64-bit
```

## Platform-Specific Builds

### Linux

Build for both amd64 and arm64:

```bash
./build.sh linux
# or
make build-linux
```

Produces:

- `bin/podium-linux-amd64` - For x86_64 systems
- `bin/podium-linux-arm64` - For ARM64 systems (e.g., Raspberry Pi 4+)

### macOS

Build for both Intel and Apple Silicon:

```bash
./build.sh macos
# or
make build-macos
```

Produces:

- `bin/podium-darwin-amd64` - For Intel Macs
- `bin/podium-darwin-arm64` - For Apple Silicon (M1/M2/M3)

### Windows

Build for Windows 64-bit:

```bash
./build.sh windows
# or
make build-windows
```

Produces:

- `bin/podium-windows-amd64.exe` - For 64-bit Windows

## Cross-Compilation

### Manual Cross-Compilation

Go makes it easy to cross-compile for different platforms:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/podium-linux-amd64 .

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o bin/podium-linux-arm64 .

# macOS AMD64 (Intel)
GOOS=darwin GOARCH=amd64 go build -o bin/podium-darwin-amd64 .

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/podium-darwin-arm64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o bin/podium-windows-amd64.exe .
```

### Supported Platforms

Podium can be built for any platform supported by Go. The build scripts include:

| OS      | Architecture | Build Output             |
| ------- | ------------ | ------------------------ |
| Linux   | amd64        | podium-linux-amd64       |
| Linux   | arm64        | podium-linux-arm64       |
| macOS   | amd64        | podium-darwin-amd64      |
| macOS   | arm64        | podium-darwin-arm64      |
| Windows | amd64        | podium-windows-amd64.exe |

## Build Requirements

- Go 1.16 or higher
- Make (optional, for Makefile)
- Git (optional, for version info)

## Build Flags

The build script automatically includes build metadata:

- **Version**: Set via `VERSION` environment variable (default: 1.0.0)
- **Git Commit**: Automatically detected from git repository
- **Build Time**: UTC timestamp of the build

Example with custom version:

```bash
VERSION=2.0.0 ./build.sh all
```

## Development Builds

For quick development builds without cross-compilation:

```bash
# Just run directly
go run main.go

# Or build for current platform only
go build -o bin/podium .
# or
make build
# or
./build.sh current
```

## Production Builds

For production deployments, use the build scripts or Makefile to ensure:

1. Binaries are placed in the correct location (`bin/`)
2. Build metadata is included
3. Proper naming conventions are followed

```bash
# Build all platforms for distribution
./build.sh all

# Or build specific platform
./build.sh linux    # For Linux servers
./build.sh windows  # For Windows servers
```

## Cleaning Build Artifacts

Remove all built binaries:

```bash
make clean
# or
./build.sh clean
# or (Windows)
build.bat clean
```

This removes:

- The entire `bin/` directory
- Any binaries in the root directory
- All cross-compiled artifacts

## Troubleshooting

### "command not found" errors

Make sure the build script is executable:

```bash
chmod +x build.sh
```

### Cross-compilation issues

If you encounter issues cross-compiling, ensure:

1. Go is properly installed: `go version`
2. CGO is disabled for pure Go builds: `export CGO_ENABLED=0`
3. Target platform is supported: `go tool dist list`

### Large binary sizes

Go binaries include the Go runtime. To reduce size:

```bash
# Strip debug symbols
go build -ldflags="-s -w" -o bin/podium .

# Use UPX compression (optional)
upx --best --lzma bin/podium
```

### Build fails on Windows

If building on Windows fails:

1. Ensure Go is in your PATH
2. Try running `build.bat` as Administrator
3. Check that no antivirus is blocking the compilation

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: "1.21"
      - run: ./build.sh all
      - uses: actions/upload-artifact@v2
        with:
          name: binaries
          path: bin/
```

### GitLab CI Example

```yaml
build:
  image: golang:1.21
  script:
    - ./build.sh all
  artifacts:
    paths:
      - bin/
```

## Distribution

After building, distribute the appropriate binary for each platform:

- **Linux servers**: `podium-linux-amd64` or `podium-linux-arm64`
- **macOS Intel**: `podium-darwin-amd64`
- **macOS Apple Silicon**: `podium-darwin-arm64`
- **Windows**: `podium-windows-amd64.exe`

Users can then install and run the binary following the installation instructions in the main README.
