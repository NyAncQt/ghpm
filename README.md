# GHPM - GitHub Package Manager (Prototype)

GHPM is a simple CLI tool written in Go that allows you to install GitHub repositories directly to your local machine.  
This project is meant to be a learning tool and a starting point for a real GitHub package manager.

---

## Installation

### Windows (Recommended)

Just run the setup script! It handles everything for you:

**Option 1 (Best):** Double-click `setup.bat` or run it from Command Prompt:
```cmd
setup.bat
```

**Option 2:** Use PowerShell:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser -Force
.\install.ps1
```

### Requirements
- **Go 1.22+** - Download from [golang.org](https://golang.org/dl/)
- **Git** - Download from [git-scm.com](https://git-scm.com/download/win)

After running the setup script, open a new terminal and you're ready to use `ghpm`!

### Linux/macOS
```bash
go build -o ghpm
sudo mv ghpm /usr/local/bin/
```

---

## Features

- Install any public GitHub repository
- List installed packages
- Remove installed packages
- Manifest tracking with JSON storage

---

## Usage

### Commands

**Install a repository:**
```bash
ghpm install owner/repo
```
Example: `ghpm install golang/go`

This will clone the repository to `~/.ghpm/packages/repo-name` and create a manifest file.

**List installed packages:**
```bash
ghpm list
```
Shows all installed packages with their repository names.

**Remove a package:**
```bash
ghpm remove repo-name
```
Example: `ghpm remove go`

Deletes the package from `~/.ghpm/packages/` and removes its manifest.

### Package Location

All packages are installed to: `~/.ghpm/packages/`

Manifest files are stored at: `~/.ghpm/manifests/`

### Example Workflow

```bash
ghpm install kubernetes/kubernetes
ghpm list
ghpm remove kubernetes
```
