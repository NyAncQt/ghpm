package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	// Ensure base, packages and manifests directories exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ghpm <command> [args]")
		fmt.Println("Commands: install, remove, list, search")
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
		repoArg := os.Args[2]
		if strings.Contains(repoArg, "/") {
			installRepo(repoArg)
		} else {
			// If user provided only a name (e.g. "btop"), search GitHub and prompt
			searchAndPrompt(repoArg)
		}
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghpm remove repo-name")
			return
		}
		removeRepo(os.Args[2])
	case "list":
		listRepos()
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghpm search <query>")
			return
		}
		searchAndPrompt(os.Args[2])
	default:
		fmt.Println("Unknown command:", command)
	}
}

// GitHub API structs
type ghSearchResult struct {
	TotalCount int          `json:"total_count"`
	Items      []ghRepoItem `json:"items"`
}

type ghRepoItem struct {
	FullName        string  `json:"full_name"`
	HTMLURL         string  `json:"html_url"`
	Description     string  `json:"description"`
	StargazersCount int     `json:"stargazers_count"`
	Language        *string `json:"language"`
}

func searchAndPrompt(query string) {
	results, err := searchRepos(query, 10)
	if err != nil {
		fmt.Println("Search failed:", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No results found for:", query)
		return
	}

	// Print results
	for i, r := range results {
		lang := "Unknown"
		if r.Language != nil && *r.Language != "" {
			lang = *r.Language
		}
		fmt.Printf("%d) %s  ★%d  %s\n", i+1, r.FullName, r.StargazersCount, lang)
	}

	// Prompt user
	fmt.Print("Select a number: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Println("Failed to read input:", err)
		return
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println("No selection made")
		return
	}
	idx, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid selection")
		return
	}
	if idx < 1 || idx > len(results) {
		fmt.Println("Selection out of range")
		return
	}

	chosen := results[idx-1]
	installRepo(chosen.FullName)
}

func searchRepos(query string, perPage int) ([]ghRepoItem, error) {
	if perPage <= 0 || perPage > 50 {
		perPage = 10
	}
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&per_page=%d", query, perPage)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghpm-cli")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var sr ghSearchResult
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&sr); err != nil {
		return nil, err
	}

	return sr.Items, nil
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
