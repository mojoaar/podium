# Podium Service Quick Reference

## Installation

### Linux

```bash
sudo ./podium -service install
sudo ./podium -service start
sudo systemctl enable Podium  # Optional: start on boot
```

### macOS

```bash
./podium -service install
./podium -service start
```

### Windows (Run as Admin)

```cmd
podium.exe -service install
podium.exe -service start
```

## Common Commands

| Action    | Linux                              | macOS                         | Windows                         |
| --------- | ---------------------------------- | ----------------------------- | ------------------------------- |
| Install   | `sudo ./podium -service install`   | `./podium -service install`   | `podium.exe -service install`   |
| Start     | `sudo ./podium -service start`     | `./podium -service start`     | `podium.exe -service start`     |
| Stop      | `sudo ./podium -service stop`      | `./podium -service stop`      | `podium.exe -service stop`      |
| Restart   | `sudo ./podium -service restart`   | `./podium -service restart`   | `podium.exe -service restart`   |
| Uninstall | `sudo ./podium -service uninstall` | `./podium -service uninstall` | `podium.exe -service uninstall` |

## Platform-Specific Management

### Linux (systemd)

```bash
sudo systemctl status Podium    # Check status
sudo systemctl enable Podium    # Start on boot
sudo systemctl disable Podium   # Don't start on boot
sudo journalctl -u Podium -f    # View logs
```

### macOS (launchd)

```bash
launchctl list | grep Podium                      # Check status
launchctl load ~/Library/LaunchAgents/Podium.plist    # Start
launchctl unload ~/Library/LaunchAgents/Podium.plist  # Stop
tail -f ~/Library/Logs/Podium.err.log            # View logs
```

### Windows

**GUI**: Win+R → `services.msc` → Find "Podium Web Server"

**PowerShell**:

```powershell
Get-Service Podium                        # Check status
Set-Service Podium -StartupType Automatic # Start on boot
```

## Troubleshooting

### Check if service is running

```bash
# Linux
sudo systemctl status Podium

# macOS
launchctl list | grep Podium

# Windows
sc query Podium
```

### Check port 8080

```bash
# Linux/macOS
lsof -i :8080

# Windows
netstat -an | findstr 8080
```

### View logs

```bash
# Linux
sudo journalctl -u Podium -n 50

# macOS
tail -50 ~/Library/Logs/Podium.err.log

# Windows - Use Event Viewer
```

## Access Website

After starting the service, visit: **http://localhost:8080**

## Development Mode

To run without installing as a service:

```bash
go run main.go
# or
./podium
```
