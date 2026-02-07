# GHPM - GitHub Package Manager (Prototype)

GHPM is a simple CLI tool written in Go that lets you install public GitHub repositories to your local machine. This repository is a prototype and learning project.

---

**Quick summary:** `ghpm install owner/repo` clones into `~/.ghpm/packages/repo-name` and creates a JSON manifest in `~/.ghpm/manifests/`.

**Repository status:** Prototype — minimal feature set, no sandboxing, and intended for local use only.

---

## Requirements

- Go 1.22+ (for building from source)
- Git (for cloning repositories)

On Linux, you also need a writable install location (either `/usr/local/bin` or `~/.local/bin`).

## Installation

### Using the bundled installer (Linux)

An install script is included: `install.sh`.

How it works:

- Builds the `ghpm` binary with `go build`.
- Installs to `/usr/local/bin` if writable, otherwise to `$HOME/.local/bin`.
- If `$HOME/.local/bin` is used the script will attempt to add that directory to a sensible shell profile (for example `~/.zshrc`, `~/.bashrc` or `~/.profile`) so the command is available in new shells.
- The script avoids `sudo` so it won't prompt for credentials; use `sudo` manually if you prefer installing to `/usr/local/bin`.

Run the installer from the project root:

```bash
./install.sh
```

If the script updated a profile file, open a new terminal or source the file, for example:

```bash
source ~/.zshrc
```

### Manual build (Linux/macOS)

```bash
go build -o ghpm
sudo mv ghpm /usr/local/bin/
```

### Windows

Use the included Windows setup scripts (`setup.bat` or `install.ps1`) as documented in the original project files.

---

## Features & Behavior

- Install any public GitHub repository by owner/repo.
- If you provide only a repository name (for example `btop`), `ghpm` will perform a GitHub search and prompt you to pick one of the matching repositories.
- Installs go into `~/.ghpm/packages/`.
- Manifest JSON files are written to `~/.ghpm/manifests/` and include fields: `name`, `repo`, `url`, `installed_at`, and optional `commit` and `version`.
- The tool creates `~/.ghpm`, `~/.ghpm/packages`, and `~/.ghpm/manifests` automatically.

Note: This tool clones repositories (git). It does not build or install software contained inside the repo — it simply downloads the repository contents into a local folder.

---

## Usage

**Install a repository (owner specified):**

```bash
ghpm install owner/repo
```

Example:

```bash
ghpm install golang/go
```

**Install by name (search):**

If you don't know the owner you can provide only the repository name and `ghpm` will search GitHub and prompt you to choose:

```bash
ghpm install btop
```

You can automate selection in scripts by piping a number into the command:

```bash
printf '1\n' | ghpm install btop
```

**List installed packages:**

```bash
ghpm list
```

**Remove a package:**

```bash
ghpm remove repo-name
```

Example workflow:

```bash
ghpm install aristocratos/btop
ghpm list
ghpm remove btop
```

---

## Troubleshooting

- If `ghpm` is not found after running `install.sh`, make sure the installer added `$HOME/.local/bin` to a shell profile and that profile has been sourced (or open a new terminal).
- To add `$HOME/.local/bin` to your current session manually:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

- If `ghpm install <name>` prints `Invalid repo format. Use owner/repo`, use the search form instead (just the repo name) — recent versions will automatically search if you provide a name without an owner.

---

## Development

Build locally:

```bash
go build -o ghpm
```

Run with `go run` for iterative development:

```bash
go run main.go list
```

---

## License

See `LICENSE` in this repository.
