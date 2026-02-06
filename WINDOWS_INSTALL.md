# Windows Installation Guide

## Quick Start (Recommended)

### Option 1: PowerShell (Recommended)

1. Open PowerShell and navigate to the project directory:
```powershell
cd path\to\ghpm
```

2. Run the install script (may need to allow script execution):
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser -Force
.\install.ps1
```

3. Done! Open a new terminal and use `ghpm`:
```powershell
ghpm list
```

### Option 2: Batch File (Fallback)

If PowerShell gives you issues, double-click `install.bat` or run:
```cmd
install.bat
```

## What the Installer Does

✅ Checks for Go and Git installations  
✅ Builds the ghpm binary  
✅ Installs to `%USERPROFILE%\AppData\Local\ghpm`  
✅ Adds ghpm to your PATH automatically  
✅ Verifies the installation works  

## Requirements

- **Go 1.22+** - Download from https://golang.org/dl/
- **Git** - Download from https://git-scm.com/download/win

## Troubleshooting

### "ghpm command not found"
Close and reopen your terminal window. PATH changes require a new terminal session.

### PowerShell Execution Policy Error
Run this in PowerShell as admin:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Build Fails
Make sure Go is properly installed:
```
go version
```

## Manual Installation

If the scripts don't work, you can install manually:

1. Build the binary:
```
go build -o ghpm.exe
```

2. Move to a directory in your PATH or create a new one:
```
mkdir %USERPROFILE%\AppData\Local\ghpm
move ghpm.exe %USERPROFILE%\AppData\Local\ghpm\
```

3. Add to PATH manually via System Properties → Environment Variables

## Usage

After installation, use ghpm:

```
ghpm install owner/repo      # Install a GitHub repository
ghpm list                    # List installed packages
ghpm remove repo-name        # Remove an installed package
```

## Uninstalling

Simply delete the ghpm folder from `%USERPROFILE%\AppData\Local\ghpm` and remove it from your PATH.
