# GitHub Package Manager (ghpm) Windows Installer
# This script sets up ghpm for Windows with all dependencies

param(
    [switch]$SkipDependencyCheck = $false
)

$ErrorActionPreference = "Stop"

Write-Host "🚀 GitHub Package Manager (ghpm) Installer" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Color helpers
function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor Blue
}

# Check if running as administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Check Go installation
function Test-Go {
    try {
        $output = go version 2>&1
        Write-Success "Go is installed: $output"
        return $true
    }
    catch {
        Write-Error-Custom "Go is not installed or not in PATH"
        Write-Host "Download from: https://golang.org/dl/" -ForegroundColor Yellow
        return $false
    }
}

# Check Git installation
function Test-Git {
    try {
        $output = git --version 2>&1
        Write-Success "Git is installed: $output"
        return $true
    }
    catch {
        Write-Error-Custom "Git is not installed or not in PATH"
        Write-Host "Download from: https://git-scm.com/download/win" -ForegroundColor Yellow
        return $false
    }
}

# Build the project
function Build-Project {
    Write-Info "Building ghpm..."
    try {
        & go build -o ghpm.exe -v
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Build completed successfully"
            return $true
        }
        else {
            Write-Error-Custom "Build failed with exit code $LASTEXITCODE"
            return $false
        }
    }
    catch {
        Write-Error-Custom "Build error: $_"
        return $false
    }
}

# Setup binary location
function Setup-BinaryLocation {
    $installDir = "$env:USERPROFILE\AppData\Local\ghpm"
    
    Write-Info "Setting up installation directory: $installDir"
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        Write-Success "Created directory: $installDir"
    }
    
    # Copy binary
    Copy-Item -Path ".\ghpm.exe" -Destination "$installDir\ghpm.exe" -Force
    Write-Success "Copied ghpm.exe to $installDir"
    
    return $installDir
}

# Add to PATH
function Add-ToPath {
    param([string]$BinaryPath)
    
    Write-Info "Adding ghpm to PATH..."
    
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($userPath -like "*$BinaryPath*") {
        Write-Success "ghpm is already in PATH"
        return $true
    }
    
    try {
        $newPath = "$userPath;$BinaryPath"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Success "Added ghpm to user PATH"
        Write-Info "PATH changes will take effect in new terminal windows"
        return $true
    }
    catch {
        Write-Error-Custom "Failed to update PATH: $_"
        Write-Info "You can manually add '$BinaryPath' to your PATH"
        return $false
    }
}

# Verify installation
function Verify-Installation {
    Write-Info "Verifying installation..."
    
    # Refresh PATH for current session
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
    
    try {
        $output = & ghpm 2>&1
        if ($output -like "*Usage: ghpm*") {
            Write-Success "ghpm is working correctly!"
            Write-Host ""
            Write-Host "Usage examples:" -ForegroundColor Cyan
            Write-Host "  ghpm install owner/repo    - Install from GitHub"
            Write-Host "  ghpm list                  - List installed packages"
            Write-Host "  ghpm remove repo-name      - Remove a package"
            return $true
        }
    }
    catch {
        Write-Error-Custom "Verification failed: $_"
        return $false
    }
}

# Main installation flow
function Main {
    Write-Host ""
    
    # Check dependencies
    if (-not $SkipDependencyCheck) {
        Write-Info "Checking dependencies..."
        Write-Host ""
        
        $goInstalled = Test-Go
        Write-Host ""
        $gitInstalled = Test-Git
        Write-Host ""
        
        if (-not $goInstalled -or -not $gitInstalled) {
            Write-Host ""
            Write-Error-Custom "Required dependencies are missing. Please install them first."
            Write-Host "After installing Go and Git, run this script again." -ForegroundColor Yellow
            exit 1
        }
    }
    
    Write-Host ""
    
    # Build
    Write-Info "Building project..."
    Write-Host ""
    if (-not (Build-Project)) {
        exit 1
    }
    
    Write-Host ""
    
    # Setup installation
    Write-Info "Setting up installation..."
    Write-Host ""
    $installDir = Setup-BinaryLocation
    
    Write-Host ""
    
    # Add to PATH
    Write-Info "Configuring PATH..."
    Write-Host ""
    Add-ToPath $installDir
    
    Write-Host ""
    
    # Verify
    Verify-Installation
    
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Success "Installation complete! 🎉"
    Write-Host "✨ You can now use 'ghpm' from any terminal window" -ForegroundColor Cyan
    Write-Host ""
}

# Run main installation
Main
