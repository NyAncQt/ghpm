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

:: Add ghpm directory to PowerShell profile(s) so future shells include it
set "PS_DIR=%USERPROFILE%\Documents\PowerShell"
set "PS_PROFILE=%PS_DIR%\Microsoft.PowerShell_profile.ps1"
set "WPS_DIR=%USERPROFILE%\Documents\WindowsPowerShell"
set "WPS_PROFILE=%WPS_DIR%\Microsoft.PowerShell_profile.ps1"

echo.
echo Ensuring PowerShell profiles will include ghpm path...

if exist "%PS_PROFILE%" (
    findstr /C:"# ghpm path" "%PS_PROFILE%" >nul 2>&1 || (
        echo # ghpm path>>"%PS_PROFILE%"
        echo $ghpm = "%USERPROFILE%\AppData\Local\ghpm" >>"%PS_PROFILE%"
        echo if (-not ($env:Path -like "*$ghpm*")) { $env:Path = $env:Path + ";" + $ghpm } >>"%PS_PROFILE%"
        echo [OK] Updated %PS_PROFILE%
    )
) else (
    if not exist "%PS_DIR%" mkdir "%PS_DIR%"
    echo # ghpm path>"%PS_PROFILE%"
    echo $ghpm = "%USERPROFILE%\AppData\Local\ghpm" >>"%PS_PROFILE%"
    echo if (-not ($env:Path -like "*$ghpm*")) { $env:Path = $env:Path + ";" + $ghpm } >>"%PS_PROFILE%"
    echo [OK] Created and added ghpm path to %PS_PROFILE%
)

if exist "%WPS_PROFILE%" (
    findstr /C:"# ghpm path" "%WPS_PROFILE%" >nul 2>&1 || (
        echo # ghpm path>>"%WPS_PROFILE%"
        echo $ghpm = "%USERPROFILE%\AppData\Local\ghpm" >>"%WPS_PROFILE%"
        echo if (-not ($env:Path -like "*$ghpm*")) { $env:Path = $env:Path + ";" + $ghpm } >>"%WPS_PROFILE%"
        echo [OK] Updated %WPS_PROFILE%
    )
) else (
    if not exist "%WPS_DIR%" mkdir "%WPS_DIR%"
    echo # ghpm path>"%WPS_PROFILE%"
    echo $ghpm = "%USERPROFILE%\AppData\Local\ghpm" >>"%WPS_PROFILE%"
    echo if (-not ($env:Path -like "*$ghpm*")) { $env:Path = $env:Path + ";" + $ghpm } >>"%WPS_PROFILE%"
    echo [OK] Created and added ghpm path to %WPS_PROFILE%
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
echo To refresh PATH in the current PowerShell session, run this command now:
echo   $env:PATH += ";%INSTALL_DIR%"
echo Or restart your terminal.
echo ==========================================
echo.
pause
