package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Manifest struct {
	Name        string    `json:"name"`
	Repo        string    `json:"repo"`
	URL         string    `json:"url"`
	InstalledAt time.Time `json:"installed_at"`
	Commit      string    `json:"commit,omitempty"`
	Version     string    `json:"version,omitempty"`
}

var baseDir, packagesDir, manifestsDir string

func initDirs() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	baseDir = filepath.Join(home, ".ghpm")
	packagesDir = filepath.Join(baseDir, "packages")
	manifestsDir = filepath.Join(baseDir, "manifests")

	return os.MkdirAll(manifestsDir, 0755)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ghpm <command> [args]")
		fmt.Println("Commands: install, remove, list, update")
		return
	}

	if err := initDirs(); err != nil {
		fmt.Println("Failed to init directories:", err)
		return
	}

	command := os.Args[1]

	switch command {
	case "install":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghpm install owner/repo")
			return
		}
		installRepo(os.Args[2])
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghpm remove repo-name")
			return
		}
		removeRepo(os.Args[2])
	case "list":
		listRepos()
	default:
		fmt.Println("Unknown command:", command)
	}
}

func installRepo(repo string) {
	if !strings.Contains(repo, "/") {
		fmt.Println("Invalid repo format. Use owner/repo")
		return
	}

	repoName := strings.Split(repo, "/")[1]
	dest := filepath.Join(packagesDir, repoName)
	if _, err := os.Stat(dest); err == nil {
		fmt.Println("Already installed:", repoName)
		return
	}

	url := "https://github.com/" + repo + ".git"
	fmt.Println("Cloning", url, "to", dest)

	cmd := exec.Command("git", "clone", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Git clone failed:", err)
		return
	}

	manifest := Manifest{
		Name:        repoName,
		Repo:        repo,
		URL:         url,
		InstalledAt: time.Now(),
	}
	saveManifest(manifest)
	fmt.Println("Installed", repoName)
}

func removeRepo(name string) {
	manifestPath := filepath.Join(manifestsDir, name+".json")
	pkgPath := filepath.Join(packagesDir, name)

	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		fmt.Println("Repo not installed:", name)
		return
	}

	os.RemoveAll(pkgPath)
	os.Remove(manifestPath)
	fmt.Println("Removed", name)
}

func listRepos() {
	files, err := os.ReadDir(manifestsDir)
	if err != nil {
		fmt.Println("Failed to read manifests:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No installed packages.")
		return
	}

	fmt.Println("Installed packages:")
	for _, f := range files {
		var m Manifest
		data, _ := os.ReadFile(filepath.Join(manifestsDir, f.Name()))
		json.Unmarshal(data, &m)
		fmt.Printf("- %s (%s)\n", m.Name, m.Repo)
	}
}

func saveManifest(m Manifest) {
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(manifestsDir, m.Name+".json"), data, 0644)
}
