@echo off
REM Podium Service Installation Script for Windows
REM Run as Administrator

echo ==========================================
echo   Podium Service Installation
echo ==========================================
echo.

set BIN_DIR=%~dp0bin
set BINARY_PATH=%BIN_DIR%\podium.exe

REM Check if binary exists
if not exist "%BINARY_PATH%" (
    echo Binary not found in bin\. Building Podium binary...
    cd /d "%~dp0"
    if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"
    go build -o "%BINARY_PATH%"
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to build binary!
        pause
        exit /b 1
    )
    echo Binary built successfully
    echo.
)

echo Installing Podium as a Windows service...
cd /d "%~dp0"
"%BINARY_PATH%" -service install

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Service installed successfully!
    echo.
    echo To start the service:
    echo   %BINARY_PATH% -service start
    echo.
    echo Or use Windows Services Manager:
    echo   1. Press Win+R, type 'services.msc', and press Enter
    echo   2. Find 'Podium Web Server' in the list
    echo   3. Right-click and select 'Start'
    echo.
    echo To access your website, visit: http://localhost:8080
    echo.
) else (
    echo.
    echo Failed to install service!
    echo Make sure you are running as Administrator.
    echo.
)

pause
