@echo off
REM GitHub Package Manager (ghpm) Windows Installer (Batch Fallback)
REM This is a fallback for users with PowerShell execution policy issues

setlocal enabledelayedexpansion

color 0A
title GitHub Package Manager - Windows Installer

echo.
echo ==========================================
echo GitHub Package Manager (ghpm) Installer
echo ==========================================
echo.

REM Check Go
echo Checking for Go...
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go is not installed or not in PATH
    echo Download from: https://golang.org/dl/
    pause
    exit /b 1
)
echo [OK] Go is installed

REM Check Git
echo Checking for Git...
git --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Git is not installed or not in PATH
    echo Download from: https://git-scm.com/download/win
    pause
    exit /b 1
)
echo [OK] Git is installed

echo.
echo Building ghpm...
go build -o ghpm.exe -v
if errorlevel 1 (
    echo [ERROR] Build failed
    pause
    exit /b 1
)
echo [OK] Build completed

REM Setup installation directory
set "INSTALL_DIR=%USERPROFILE%\AppData\Local\ghpm"
echo.
echo Setting up installation directory: %INSTALL_DIR%
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
)
copy ghpm.exe "%INSTALL_DIR%\ghpm.exe" /Y >nul
echo [OK] Copied ghpm.exe to %INSTALL_DIR%

echo.
echo Adding to PATH...
setx PATH "%PATH%;%INSTALL_DIR%"
echo [OK] Added to user PATH

echo.
echo Verifying installation...
"%INSTALL_DIR%\ghpm.exe" 2>&1 | find "Usage: ghpm" >nul
if errorlevel 1 (
    echo [WARN] Could not verify installation
) else (
    echo [OK] ghpm is working correctly!
)

echo.
echo ==========================================
echo Installation complete!
echo.
echo Usage:
echo   ghpm install owner/repo    - Install from GitHub
echo   ghpm list                  - List installed packages
echo   ghpm remove repo-name      - Remove a package
echo.
echo Note: Please close and reopen your terminal for PATH changes to take effect.
echo ==========================================
echo.
pause
