@echo off
REM GitHub Package Manager (ghpm) - Complete Windows Setup
REM This script handles everything: dependencies check, build, and installation

setlocal enabledelayedexpansion
title GitHub Package Manager Setup

:: Simple color output using findstr trick
echo.
echo ==========================================
echo. | findstr /r "^" > nul && (
    cls
    color 0A
)
echo GitHub Package Manager - Windows Setup
echo ==========================================
echo.

:: Check Administrator Privileges (nice to have but not required)
net session >nul 2>&1
if %errorlevel% == 0 (
    echo [INFO] Running with administrator privileges
) else (
    echo [INFO] Running without admin privileges - some features may be limited
)

echo.
echo Step 1: Checking dependencies...
echo.

:: Check Go
echo Checking Go installation...
go version >nul 2>&1
if errorlevel 1 (
    color 0C
    echo [ERROR] Go is not installed or not in PATH
    echo.
    echo Please install Go from: https://golang.org/dl/
    echo Then restart your terminal and run this script again.
    echo.
    pause
    exit /b 1
)
for /f "tokens=3" %%a in ('go version') do echo [OK] Go is installed: %%a

:: Check Git
echo Checking Git installation...
git --version >nul 2>&1
if errorlevel 1 (
    color 0C
    echo [ERROR] Git is not installed or not in PATH
    echo.
    echo Please install Git from: https://git-scm.com/download/win
    echo Then restart your terminal and run this script again.
    echo.
    pause
    exit /b 1
)
for /f "tokens=1,2,3" %%a in ('git --version') do echo [OK] %%a %%b %%c installed

echo.
echo Step 2: Building ghpm...
echo.

:: Build
go build -o ghpm.exe -v
if errorlevel 1 (
    color 0C
    echo [ERROR] Build failed!
    echo.
    pause
    exit /b 1
)
echo [OK] Build successful!

echo.
echo Step 3: Installing ghpm...
echo.

:: Create install directory
set "GHPM_DIR=%USERPROFILE%\AppData\Local\ghpm"
if not exist "%GHPM_DIR%" (
    mkdir "%GHPM_DIR%"
    echo [OK] Created directory: %GHPM_DIR%
) else (
    echo [OK] Directory exists: %GHPM_DIR%
)

:: Copy binary
copy ghpm.exe "%GHPM_DIR%\ghpm.exe" /Y >nul 2>&1
if errorlevel 1 (
    color 0C
    echo [ERROR] Failed to copy ghpm.exe
    pause
    exit /b 1
)
echo [OK] Copied ghpm.exe to %GHPM_DIR%

echo.
echo Step 4: Adding to PATH...
echo.

:: Check if already in PATH
echo %PATH% | find /i "%GHPM_DIR%">nul
if %errorlevel% == 0 (
    echo [OK] ghpm is already in PATH
) else (
    setx PATH "%PATH%;%GHPM_DIR%"
    if errorlevel 1 (
        echo [WARN] Could not update PATH automatically
        echo [INFO] Please manually add to PATH: %GHPM_DIR%
    ) else (
        echo [OK] Added ghpm to user PATH
        echo [INFO] PATH changes will take effect in new terminal windows
    )
)

echo.
echo Step 5: Verifying installation...
echo.

:: Test ghpm
call "%GHPM_DIR%\ghpm.exe" 2>&1 | find "Usage: ghpm">nul
if %errorlevel% == 0 (
    color 0A
    echo [OK] ghpm is working correctly!
    echo.
    echo ==========================================
    echo. Installation Complete!
    echo ==========================================
    echo.
    echo Quick Start:
    echo   ghpm install owner/repo    - Install from GitHub
    echo   ghpm list                  - List installed packages
    echo   ghpm remove repo-name      - Remove a package
    echo.
    echo NOTE: Close and reopen your terminal for PATH changes to take effect.
    echo.
) else (
    color 0E
    echo [WARN] Could not verify installation
    echo [INFO] Try closing and reopening your terminal
    echo.
)

pause
