@echo off
REM Podium Build Script for Windows
REM Builds binaries for all supported platforms

setlocal enabledelayedexpansion

set VERSION=1.0.0
set BUILD_DIR=bin
set APP_NAME=podium

echo ==========================================
echo   Building Podium v%VERSION%
echo ==========================================
echo.

REM Create bin directory if it doesn't exist
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"

REM Get build time
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "BUILD_TIME=%dt:~0,4%-%dt:~4,2%-%dt:~6,2%_%dt:~8,2%:%dt:~10,2%:%dt:~12,2%"

REM Parse command line arguments
set BUILD_ALL=0
set BUILD_CURRENT=0
set BUILD_LINUX=0
set BUILD_MACOS=0
set BUILD_WINDOWS=0

if "%1"=="" (
    set BUILD_CURRENT=1
) else if "%1"=="all" (
    set BUILD_ALL=1
) else if "%1"=="current" (
    set BUILD_CURRENT=1
) else if "%1"=="linux" (
    set BUILD_LINUX=1
) else if "%1"=="macos" (
    set BUILD_MACOS=1
) else if "%1"=="windows" (
    set BUILD_WINDOWS=1
) else if "%1"=="clean" (
    echo Cleaning build directory...
    rmdir /s /q "%BUILD_DIR%" 2>nul
    mkdir "%BUILD_DIR%"
    echo Cleaned
    goto :end
) else (
    echo Unknown option: %1
    echo.
    echo Usage: %0 [all^|current^|linux^|macos^|windows^|clean]
    echo.
    echo Options:
    echo   all       - Build for all platforms
    echo   current   - Build for current platform only (default)
    echo   linux     - Build for Linux (amd64)
    echo   macos     - Build for macOS (amd64 and arm64)
    echo   windows   - Build for Windows (amd64)
    echo   clean     - Remove all built binaries
    echo.
    goto :end
)

REM Build based on options
if %BUILD_ALL%==1 (
    call :build_platform linux amd64 %APP_NAME%-linux-amd64
    call :build_platform linux arm64 %APP_NAME%-linux-arm64
    call :build_platform darwin amd64 %APP_NAME%-darwin-amd64
    call :build_platform darwin arm64 %APP_NAME%-darwin-arm64
    call :build_platform windows amd64 %APP_NAME%-windows-amd64.exe
) else if %BUILD_CURRENT%==1 (
    call :build_platform windows amd64 %APP_NAME%.exe
) else (
    if %BUILD_LINUX%==1 (
        call :build_platform linux amd64 %APP_NAME%-linux-amd64
        call :build_platform linux arm64 %APP_NAME%-linux-arm64
    )
    if %BUILD_MACOS%==1 (
        call :build_platform darwin amd64 %APP_NAME%-darwin-amd64
        call :build_platform darwin arm64 %APP_NAME%-darwin-arm64
    )
    if %BUILD_WINDOWS%==1 (
        call :build_platform windows amd64 %APP_NAME%-windows-amd64.exe
    )
)

echo ==========================================
echo   Build Complete!
echo ==========================================
echo.
echo Binaries are in the '%BUILD_DIR%' directory
dir /b "%BUILD_DIR%"
echo.
goto :end

:build_platform
set GOOS=%1
set GOARCH=%2
set OUTPUT=%3

echo Building for %GOOS%/%GOARCH%...

set CGO_ENABLED=0
go build -o "%BUILD_DIR%\%OUTPUT%" .

if %ERRORLEVEL% EQU 0 (
    echo [OK] Successfully built: %BUILD_DIR%\%OUTPUT%
    dir "%BUILD_DIR%\%OUTPUT%" | findstr /V "Volume Serial Directory"
) else (
    echo [FAIL] Failed to build for %GOOS%/%GOARCH%
)
echo.
goto :eof

:end
endlocal
