package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ghpm install owner/repo")
		return
	}

	command := os.Args[1]
	target := os.Args[2]

	switch command {
	case "install":
		installRepo(target)
	default:
		fmt.Println("Unknown command:", command)
	}
}

func installRepo(repo string) {
	if !strings.Contains(repo, "/") {
		fmt.Println("Invalid repo format. Use owner/repo")
		return
	}

	url := "https://github.com/" + repo + ".git"

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Cannot find home directory")
		return
	}

	installDir := filepath.Join(home, ".ghpm", "packages")
	err = os.MkdirAll(installDir, 0755)
	if err != nil {
		fmt.Println("Failed to create ghpm directory")
		return
	}

	repoName := strings.Split(repo, "/")[1]
	dest := filepath.Join(installDir, repoName)

	fmt.Println("Cloning", url, "to", dest)

	cmd := exec.Command("git", "clone", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Git clone failed:", err)
		return
	}

	fmt.Println("Installed", repoName)
}
