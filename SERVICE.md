# Podium Service Management Guide

This guide provides detailed instructions for installing, managing, and troubleshooting Podium as a system service across different operating systems.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
  - [Linux](#linux)
  - [macOS](#macos)
  - [Windows](#windows)
- [Service Management](#service-management)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)

## Overview

Podium can run as a system service, which provides several benefits:

- **Auto-start on boot**: The service starts automatically when your system boots
- **Background operation**: Runs in the background without needing a terminal open
- **Service management**: Use standard OS tools to manage the service
- **Reliability**: Automatic restarts on failure (depending on OS configuration)

## Installation

### Linux

#### Prerequisites

- Go 1.16 or higher
- sudo access
- systemd (most modern Linux distributions)

#### Installation Steps

1. **Build the binary**:

   ```bash
   cd /path/to/podium
   go build -o podium
   ```

2. **Install the service**:

   ```bash
   sudo ./podium -service install
   ```

3. **Start the service**:

   ```bash
   sudo ./podium -service start
   ```

4. **Enable on boot** (optional):

   ```bash
   sudo systemctl enable Podium
   ```

5. **Verify the service is running**:
   ```bash
   sudo systemctl status Podium
   ```

#### Using the Installation Script

Alternatively, use the provided script:

```bash
sudo ./install-service.sh
```

#### Service Files Location

On Linux with systemd, the service file is created at:

- `/etc/systemd/system/Podium.service`

### macOS

#### Prerequisites

- Go 1.16 or higher
- macOS 10.10 or later

#### Installation Steps

1. **Build the binary**:

   ```bash
   cd /path/to/podium
   go build -o podium
   ```

2. **Install the service**:

   ```bash
   ./podium -service install
   ```

3. **Start the service**:

   ```bash
   ./podium -service start
   ```

4. **Verify the service is running**:
   ```bash
   launchctl list | grep Podium
   ```

#### Using the Installation Script

Alternatively, use the provided script:

```bash
./install-service.sh
```

#### Service Files Location

On macOS, the LaunchAgent plist file is created at:

- `~/Library/LaunchAgents/Podium.plist` (user service)
- `/Library/LaunchDaemons/Podium.plist` (system service, requires sudo)

### Windows

#### Prerequisites

- Go 1.16 or higher
- Administrator privileges

#### Installation Steps

1. **Open Command Prompt or PowerShell as Administrator**

   - Right-click on Command Prompt/PowerShell
   - Select "Run as Administrator"

2. **Navigate to Podium directory**:

   ```cmd
   cd C:\path\to\podium
   ```

3. **Build the binary**:

   ```cmd
   go build -o podium.exe
   ```

4. **Install the service**:

   ```cmd
   podium.exe -service install
   ```

5. **Start the service**:

   ```cmd
   podium.exe -service start
   ```

6. **Verify the service is running**:
   ```cmd
   sc query Podium
   ```

#### Using the Installation Script

Alternatively, **run as Administrator**:

```cmd
install-service.bat
```

#### Service Configuration

On Windows, the service is registered in the Windows Service Manager and can be configured there.

## Service Management

### Command-Line Management

All platforms support these commands:

```bash
# Install service
./podium -service install    # sudo required on Linux

# Start service
./podium -service start      # sudo required on Linux

# Stop service
./podium -service stop       # sudo required on Linux

# Restart service
./podium -service restart    # sudo required on Linux

# Uninstall service
./podium -service uninstall  # sudo required on Linux
```

### Platform-Specific Tools

#### Linux (systemd)

```bash
# Start
sudo systemctl start Podium

# Stop
sudo systemctl stop Podium

# Restart
sudo systemctl restart Podium

# Check status
sudo systemctl status Podium

# Enable on boot
sudo systemctl enable Podium

# Disable on boot
sudo systemctl disable Podium

# View logs
sudo journalctl -u Podium -f

# View recent logs
sudo journalctl -u Podium -n 50
```

#### macOS (launchd)

```bash
# Start
launchctl load ~/Library/LaunchAgents/Podium.plist

# Stop
launchctl unload ~/Library/LaunchAgents/Podium.plist

# Check if running
launchctl list | grep Podium

# View logs
tail -f ~/Library/Logs/Podium.out.log
tail -f ~/Library/Logs/Podium.err.log
```

#### Windows

**Services Manager (GUI)**:

1. Press `Win+R`, type `services.msc`, press Enter
2. Find "Podium Web Server"
3. Right-click for options: Start, Stop, Restart, Properties

**Command Prompt**:

```cmd
# Start
sc start Podium

# Stop
sc stop Podium

# Query status
sc query Podium

# Configure to start automatically
sc config Podium start= auto

# Configure to start manually
sc config Podium start= demand
```

**PowerShell**:

```powershell
# Start
Start-Service Podium

# Stop
Stop-Service Podium

# Restart
Restart-Service Podium

# Get status
Get-Service Podium

# Set to start automatically
Set-Service Podium -StartupType Automatic

# Set to manual start
Set-Service Podium -StartupType Manual
```

## Troubleshooting

### Service Won't Install

**Linux**:

- Ensure you're using `sudo`
- Check that systemd is available: `systemctl --version`
- Verify the binary has execute permissions: `chmod +x podium`

**macOS**:

- Ensure you have write permissions to `~/Library/LaunchAgents/`
- Check that the binary path is absolute

**Windows**:

- Ensure you're running as Administrator
- Check that the binary name is `podium.exe`

### Service Won't Start

**Check Port Availability**:

```bash
# Linux/macOS
lsof -i :8080
netstat -an | grep 8080

# Windows
netstat -an | findstr 8080
```

**Check File Permissions** (Linux/macOS):

```bash
ls -la podium
# Should show executable permissions (rwxr-xr-x)
```

**Check Working Directory**:
Ensure the working directory contains:

- `templates/` folder with HTML templates
- `static/` folder (can be empty)
- `posts/` folder (can be empty)
- `assets/` folder with `style.css`

**View Logs**:

Linux:

```bash
sudo journalctl -u Podium -n 100 --no-pager
```

macOS:

```bash
cat ~/Library/Logs/Podium.err.log
```

Windows:

- Open Event Viewer
- Navigate to Windows Logs → Application
- Look for Podium entries

### Service Crashes

1. **Test in development mode**:

   ```bash
   go run main.go
   # or
   ./podium
   ```

   This will show errors directly in the terminal.

2. **Check dependencies**:

   ```bash
   go mod verify
   go mod download
   ```

3. **Rebuild the binary**:
   ```bash
   go clean
   go build -o podium
   ```

### Cannot Access Website

1. **Verify service is running**:

   - Use platform-specific commands above

2. **Check firewall**:

   Linux (ufw):

   ```bash
   sudo ufw allow 8080
   ```

   Linux (firewalld):

   ```bash
   sudo firewall-cmd --permanent --add-port=8080/tcp
   sudo firewall-cmd --reload
   ```

   macOS:

   - System Preferences → Security & Privacy → Firewall → Firewall Options
   - Allow incoming connections for Podium

   Windows:

   ```powershell
   New-NetFirewallRule -DisplayName "Podium" -Direction Inbound -Port 8080 -Protocol TCP -Action Allow
   ```

3. **Test locally**:
   ```bash
   curl http://localhost:8080
   ```

### Permission Denied Errors

**Linux**:

- Use `sudo` for service commands
- Check SELinux status: `getenforce`
- If SELinux is enforcing, you may need to set proper contexts

**macOS**:

- Ensure the binary is in a user-accessible location
- Don't use `/usr/local/bin` without proper permissions

**Windows**:

- Always run service commands as Administrator
- Check that the executable isn't blocked (Properties → Unblock)

## Uninstallation

### Linux

```bash
# Stop the service
sudo ./podium -service stop

# Uninstall
sudo ./podium -service uninstall

# Or use the script
sudo ./uninstall-service.sh
```

### macOS

```bash
# Stop the service
./podium -service stop

# Uninstall
./podium -service uninstall

# Or use the script
./uninstall-service.sh
```

### Windows

**Run as Administrator**:

```cmd
podium.exe -service stop
podium.exe -service uninstall
```

Or use the batch file:

```cmd
uninstall-service.bat
```

### Manual Cleanup

If automatic uninstallation fails:

**Linux**:

```bash
sudo systemctl stop Podium
sudo systemctl disable Podium
sudo rm /etc/systemd/system/Podium.service
sudo systemctl daemon-reload
```

**macOS**:

```bash
launchctl unload ~/Library/LaunchAgents/Podium.plist
rm ~/Library/LaunchAgents/Podium.plist
```

**Windows**:

```cmd
sc stop Podium
sc delete Podium
```

## Additional Notes

### Security Considerations

- The service runs on port 8080 by default
- Consider using a reverse proxy (nginx, Apache) for production deployments
- Use HTTPS in production (can be configured through reverse proxy)
- Restrict firewall access to necessary IP ranges

### Performance

- The service runs in release mode (production mode)
- Logs are minimized for better performance
- Consider using a process supervisor for additional monitoring

### Updates

To update Podium:

1. Stop the service
2. Replace the binary with the new version
3. Restart the service

```bash
# Linux
sudo systemctl stop Podium
go build -o podium
sudo systemctl start Podium

# macOS
./podium -service stop
go build -o podium
./podium -service start

# Windows (as Administrator)
sc stop Podium
go build -o podium.exe
sc start Podium
```

## Support

For issues and questions:

- Check the main README.md
- Review the logs using platform-specific commands
- Ensure all prerequisites are met
- Test in development mode first (`go run main.go`)
