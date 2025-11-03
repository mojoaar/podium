# Podium - System Service Edition

## What's New

Podium now supports installation as a system service on:

- ✅ **Windows** (Windows Service)
- ✅ **macOS** (launchd)
- ✅ **Linux** (systemd)

## Key Features

- Run automatically on system startup
- Background operation without terminal
- Cross-platform service management
- Simple installation scripts
- Platform-native service integration

## Quick Start

### 1. Build the Binary

```bash
go build -o podium
```

### 2. Install as Service

**Linux:**

```bash
sudo ./install-service.sh
```

**macOS:**

```bash
./install-service.sh
```

**Windows (as Admin):**

```cmd
install-service.bat
```

### 3. Access Your Website

Visit: **http://localhost:8080**

## Files Added

- `main.go` - Updated with service support
- `install-service.sh` - Linux/macOS installation script
- `uninstall-service.sh` - Linux/macOS uninstallation script
- `install-service.bat` - Windows installation script
- `uninstall-service.bat` - Windows uninstallation script
- `Makefile` - Build and service management commands
- `SERVICE.md` - Comprehensive service documentation
- `QUICKREF.md` - Quick reference guide

## Documentation

- **README.md** - Main documentation with service installation
- **SERVICE.md** - Detailed service management guide
- **QUICKREF.md** - Quick reference for commands

## Makefile Commands

```bash
make build          # Build binary
make install        # Install as service
make start          # Start service
make stop           # Stop service
make restart        # Restart service
make uninstall      # Uninstall service
make run            # Run in development mode
make clean          # Clean build artifacts
```

## Service Management

### Universal Commands (All Platforms)

```bash
./podium -service install    # Install service
./podium -service start      # Start service
./podium -service stop       # Stop service
./podium -service restart    # Restart service
./podium -service uninstall  # Uninstall service
```

**Note:** Use `sudo` on Linux for all service commands.

### Platform-Specific Tools

**Linux (systemd):**

```bash
sudo systemctl start Podium
sudo systemctl stop Podium
sudo systemctl status Podium
sudo systemctl enable Podium  # Start on boot
```

**macOS (launchd):**

```bash
launchctl load ~/Library/LaunchAgents/Podium.plist
launchctl unload ~/Library/LaunchAgents/Podium.plist
```

**Windows:**

- Services Manager: `services.msc`
- PowerShell: `Start-Service Podium`
- Command: `sc start Podium`

## Dependencies

The following Go package was added:

- `github.com/kardianos/service` - Cross-platform service management

## Upgrade from Previous Version

If you have an existing Podium installation:

1. Pull the latest code
2. Run: `go mod download`
3. Rebuild: `go build -o podium`
4. Install as service (if desired)

## Support

For detailed troubleshooting and advanced configuration, see:

- `SERVICE.md` - Complete service documentation
- `QUICKREF.md` - Quick command reference

---

**Ready to use!** Your Podium installation can now run as a professional system service. 🚀
