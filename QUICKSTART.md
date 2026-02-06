# 🚀 GHPM - Windows Quick Start

## Installation (30 seconds)

### All-in-One Setup

1. **Double-click** `setup.bat`
   - That's it! The script handles everything.

OR

2. From Command Prompt run:
```
setup.bat
```

### What happens:
✅ Checks for Go and Git  
✅ Builds ghpm  
✅ Installs to your computer  
✅ Adds to PATH so you can use it anywhere  

### Then what?
Close and reopen your terminal, then use:

```cmd
ghpm install owner/repo    # Example: ghpm install golang/go
ghpm list                  # See what you installed
ghpm remove repo-name      # Remove a package
```

---

## Troubleshooting

### ❌ Setup fails or commands not found
1. Make sure you have **Go** and **Git** installed
2. Download Go: https://golang.org/dl/
3. Download Git: https://git-scm.com/download/win
4. Close/reopen Terminal
5. Run `setup.bat` again

### ❌ "ghpm command not found" after installation
- Close and reopen your terminal window completely
- PATH changes require a new terminal session

### ❌ PowerShell script execution issue
Use `setup.bat` instead (it's the Batch version)

---

## That's it!

Your GitHub package manager is ready to use. 🎉
