@echo off
REM Podium Service Uninstallation Script for Windows
REM Run as Administrator

echo ==========================================
echo   Podium Service Uninstallation
echo ==========================================
echo.

set BIN_DIR=%~dp0bin
set BINARY_PATH=%BIN_DIR%\podium.exe

if not exist "%BINARY_PATH%" (
    echo Error: podium.exe not found in bin\ directory!
    pause
    exit /b 1
)

echo Uninstalling Podium service...
cd /d "%~dp0"
"%BINARY_PATH%" -service uninstall

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Service uninstalled successfully!
    echo.
) else (
    echo.
    echo Failed to uninstall service!
    echo Make sure you are running as Administrator.
    echo.
)

pause
