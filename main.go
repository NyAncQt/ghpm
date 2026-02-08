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
	Language    string    `json:"language,omitempty"`
	Built       bool      `json:"built,omitempty"`
	BuildCmd    string    `json:"build_cmd,omitempty"`
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
		fmt.Println("Commands: install, remove, list, search, update, info")
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
	case "update":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghpm update <repo-name>")
			return
		}
		updateRepo(os.Args[2])
	case "info":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ghpm info <repo-name>")
			return
		}
		showInfo(os.Args[2])
	default:
		fmt.Println("Unknown command:", command)
	}
}

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

	for i, r := range results {
		lang := "Unknown"
		if r.Language != nil && *r.Language != "" {
			lang = *r.Language
		}
		fmt.Printf("%d) %s  ★%d  %s\n", i+1, r.FullName, r.StargazersCount, lang)
	}

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

func detectLanguage(repoPath string) string {
	checks := []struct {
		file string
		lang string
	}{
		{"go.mod", "Go"},
		{"Cargo.toml", "Rust"},
		{"package.json", "Node"},
		{"setup.py", "Python"},
		{"pyproject.toml", "Python"},
		{"requirements.txt", "Python"},
		{"Gemfile", "Ruby"},
		{"Makefile", "C/C++"},
		{"CMakeLists.txt", "C/C++"},
	}

	for _, check := range checks {
		if _, err := os.Stat(filepath.Join(repoPath, check.file)); err == nil {
			return check.lang
		}
	}

	entries, err := os.ReadDir(repoPath)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".sh") {
				if strings.Contains(strings.ToLower(e.Name()), "install") {
					return "Shell"
				}
			}
		}
	}

	if matches, _ := filepath.Glob(filepath.Join(repoPath, "*.gemspec")); len(matches) > 0 {
		return "Ruby"
	}

	return "Unknown"
}

type binLink struct {
	Name string
	Path string
}

func packageJSONBinLinks(repoPath string) []binLink {
	data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
	if err != nil {
		return nil
	}
	var raw struct {
		Name string          `json:"name"`
		Bin  json.RawMessage `json:"bin"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	if len(raw.Bin) == 0 {
		return nil
	}

	var links []binLink
	if len(raw.Bin) > 0 && raw.Bin[0] == '"' {
		var rel string
		if err := json.Unmarshal(raw.Bin, &rel); err == nil && raw.Name != "" {
			name := filepath.Base(raw.Name)
			links = append(links, binLink{Name: name, Path: filepath.Join(repoPath, rel)})
		}
	} else if len(raw.Bin) > 0 && raw.Bin[0] == '{' {
		var m map[string]string
		if err := json.Unmarshal(raw.Bin, &m); err == nil {
			for name, rel := range m {
				links = append(links, binLink{Name: name, Path: filepath.Join(repoPath, rel)})
			}
		}
	}

	return links
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func warnIfPathMissing(binDir string) {
	path := os.Getenv("PATH")
	if path == "" {
		fmt.Println("Warning: PATH is empty; you may not be able to run linked binaries.")
		return
	}
	for _, p := range strings.Split(path, ":") {
		if p == binDir {
			return
		}
	}
	fmt.Println("Warning:", binDir, "is not on your PATH. Add it to ~/.zshrc or ~/.bashrc.")
}

func autoBuildRepo(repoPath, language string) (bool, string) {
	// Check for --no-build flag
	for _, arg := range os.Args {
		if arg == "--no-build" {
			fmt.Println("Skipping build (--no-build flag)")
			return false, "skipped"
		}
	}

	if language == "Unknown" {
		fmt.Println("Could not detect language - skipping auto-build")
		fmt.Println("You may need to build/install manually. Check the repo's README.")
		return false, "unknown language"
	}

	fmt.Println("Detected language:", language)
	fmt.Println("Attempting auto-build/install...")

	var cmd *exec.Cmd
	var cmdDesc string

	switch language {
	case "Go":
		if !commandExists("go") {
			fmt.Println("Go is not installed or not on PATH.")
			return false, "missing go"
		}
		// Try go install first
		cmdDesc = "go install"
		cmd = exec.Command("go", "install")
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Fallback to go build
			fmt.Println("go install failed, trying go build...")
			cmdDesc = "go build"
			cmd = exec.Command("go", "build")
			cmd.Dir = repoPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("Build failed. You may need to build manually.")
				return false, cmdDesc
			}
		}
		fmt.Println("Build successful!")
		return true, cmdDesc

	case "Rust":
		if !commandExists("cargo") {
			fmt.Println("Rust (cargo) is not installed or not on PATH.")
			return false, "missing cargo"
		}
		// Try cargo install --path first
		cmdDesc = "cargo install --path ."
		cmd = exec.Command("cargo", "install", "--path", ".")
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Fallback to cargo build
			fmt.Println("cargo install failed, trying cargo build...")
			cmdDesc = "cargo build --release"
			cmd = exec.Command("cargo", "build", "--release")
			cmd.Dir = repoPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("Build failed. You may need to build manually.")
				return false, cmdDesc
			}
		}
		fmt.Println("Build successful!")
		return true, cmdDesc

	case "Node":
		if !commandExists("npm") {
			fmt.Println("Node (npm) is not installed or not on PATH.")
			return false, "missing npm"
		}
		cmdDesc = "npm install"
		cmd = exec.Command("npm", "install")
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("npm install failed. You may need to install manually.")
			return false, cmdDesc
		}
		fmt.Println("Dependencies installed!")
		fmt.Println("Note: This is a Node project. Check package.json for run commands.")
		return true, cmdDesc

	case "Python":
		if !commandExists("pip") && !commandExists("python") {
			fmt.Println("Python is not installed or not on PATH.")
			return false, "missing python/pip"
		}
		if commandExists("pip") {
			// Try pip install . first
			cmdDesc = "pip install ."
			cmd = exec.Command("pip", "install", ".")
			cmd.Dir = repoPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// Fallback to setup.py
				fmt.Println("pip install failed, trying setup.py...")
				if !commandExists("python") {
					fmt.Println("python is not installed or not on PATH.")
					return false, "missing python"
				}
				cmdDesc = "python setup.py install"
				cmd = exec.Command("python", "setup.py", "install")
				cmd.Dir = repoPath
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Println("Install failed. You may need to install manually.")
					return false, cmdDesc
				}
			}
			fmt.Println("Build successful!")
			return true, cmdDesc
		}

		cmdDesc = "python setup.py install"
		cmd = exec.Command("python", "setup.py", "install")
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Install failed. You may need to install manually.")
			return false, cmdDesc
		}
		fmt.Println("Build successful!")
		return true, cmdDesc

	case "Ruby":

		if !commandExists("ruby") {
			fmt.Println("Ruby is not installed or not on PATH.")
			return false, "missing ruby"
		}
		fmt.Println("Ruby project detected. Skipping auto-build; will link bin scripts if present.")
		return false, "ruby: no build required"

	case "Shell":
		if !commandExists("sh") {
			fmt.Println("sh is not available on PATH.")
			return false, "missing sh"
		}

		installScript := ""
		entries, _ := os.ReadDir(repoPath)
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Name()), "install") && strings.HasSuffix(e.Name(), ".sh") {
				installScript = e.Name()
				break
			}
		}
		if installScript == "" {
			fmt.Println("No install.sh found. Check the README for manual installation.")
			return false, "no install.sh found"
		}

		scriptPath := filepath.Join(repoPath, installScript)

		if err := os.Chmod(scriptPath, 0755); err != nil {
			fmt.Println("Failed to make install script executable")
			return false, "chmod +x " + installScript
		}

		cmdDesc = "./" + installScript
		fmt.Println("Running", cmdDesc, "...")
		cmd = exec.Command("sh", scriptPath)
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Install script failed. Check the README for manual installation.")
			return false, cmdDesc
		}
		fmt.Println("Install script completed!")
		return true, cmdDesc

	case "C/C++":
		hasMake := commandExists("make")
		hasCmake := commandExists("cmake")
		if !hasMake && !hasCmake {
			fmt.Println("Neither make nor cmake is available on PATH.")
			return false, "missing make/cmake"
		}

		cmdDesc = "make"
		if hasMake {
			_, err := os.Stat(filepath.Join(repoPath, "Makefile"))
			if err == nil {
				fmt.Println("Running make...")
				cmd = exec.Command("make")
				cmd.Dir = repoPath
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Println("make failed. You may need to build manually.")
					return false, cmdDesc
				}
				// Try make install (don't fail if it doesn't work)
				fmt.Println("Running make install...")
				cmdInstall := exec.Command("make", "install")
				cmdInstall.Dir = repoPath
				cmdInstall.Stdout = os.Stdout
				cmdInstall.Stderr = os.Stderr
				if err := cmdInstall.Run(); err != nil {
					fmt.Println("make install failed (this is sometimes expected)")
					fmt.Println("Binary may be in:", repoPath)
				}
				fmt.Println("Build successful!")
				return true, cmdDesc + " && make install"
			}
		}

		if hasCmake {
			_, err := os.Stat(filepath.Join(repoPath, "CMakeLists.txt"))
			if err == nil {
				buildDir := filepath.Join(repoPath, "build")
				os.MkdirAll(buildDir, 0755)

				cmdDesc = "cmake && make"
				fmt.Println("Running cmake...")
				cmd = exec.Command("cmake", "..")
				cmd.Dir = buildDir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Println("cmake failed. You may need to build manually.")
					return false, cmdDesc
				}

				fmt.Println("Running make...")
				cmd = exec.Command("make")
				cmd.Dir = buildDir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Println("make failed. You may need to build manually.")
					return false, cmdDesc
				}
				fmt.Println("Build successful!")
				return true, cmdDesc
			}
		}

		fmt.Println("No Makefile or CMakeLists.txt found. Check README for build instructions.")
		return false, "no build system found"

	default:
		fmt.Println("Unsupported language for auto-build")
		return false, "unsupported language"
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

	language := detectLanguage(dest)
	built, buildCmd := autoBuildRepo(dest, language)

	manifest := Manifest{
		Name:        repoName,
		Repo:        repo,
		URL:         url,
		InstalledAt: time.Now(),
		Language:    language,
		Built:       built,
		BuildCmd:    buildCmd,
	}
	saveManifest(manifest)

	linkBinaries(dest, repoName, language)

	fmt.Println("Installed", repoName)
	if !built && language != "Unknown" {
		fmt.Println("Package cloned but not built. Check", dest, "for manual build instructions.")
	}
}

func linkBinaries(repoPath, repoName, language string) {

	binDir := filepath.Join(os.Getenv("HOME"), ".local", "bin")
	os.MkdirAll(binDir, 0755)
	warnIfPathMissing(binDir)

	var binaries []string
	isExecutable := func(path string) bool {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return false
		}
		return info.Mode()&0111 != 0
	}

	switch language {
	case "Go":

		gobin := os.Getenv("GOBIN")
		if gobin == "" {
			gobin = filepath.Join(os.Getenv("HOME"), "go", "bin")
		}
		binPath := filepath.Join(gobin, repoName)
		if _, err := os.Stat(binPath); err == nil {
			binaries = append(binaries, binPath)
		} else {

			binPath = filepath.Join(repoPath, repoName)
			if _, err := os.Stat(binPath); err == nil {
				binaries = append(binaries, binPath)
			}
		}

	case "Rust":

		binPath := filepath.Join(repoPath, "target", "release", repoName)
		if _, err := os.Stat(binPath); err == nil {
			binaries = append(binaries, binPath)
		}

	case "Node":

		for _, link := range packageJSONBinLinks(repoPath) {
			if _, err := os.Stat(link.Path); err == nil {
				os.Chmod(link.Path, 0755)
				binaries = append(binaries, link.Path)
			}
		}

		entries, _ := os.ReadDir(repoPath)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == repoName || strings.Contains(strings.ToLower(name), "install") || strings.HasSuffix(name, ".js") {
				binPath := filepath.Join(repoPath, name)
				os.Chmod(binPath, 0755)
				binaries = append(binaries, binPath)
			}
		}
		binEntries, _ := os.ReadDir(filepath.Join(repoPath, "bin"))
		for _, e := range binEntries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == repoName || strings.Contains(strings.ToLower(name), "install") || strings.HasSuffix(name, ".js") {
				binPath := filepath.Join(repoPath, "bin", name)
				os.Chmod(binPath, 0755)
				binaries = append(binaries, binPath)
			}
		}

	case "Python":

		entries, _ := os.ReadDir(repoPath)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == repoName || strings.Contains(strings.ToLower(name), "install") || strings.HasSuffix(name, ".py") || strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".sh") {
				binPath := filepath.Join(repoPath, name)
				os.Chmod(binPath, 0755)
				binaries = append(binaries, binPath)
			}
		}

		binEntries, _ := os.ReadDir(filepath.Join(repoPath, "bin"))
		for _, e := range binEntries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == repoName || strings.Contains(strings.ToLower(name), "install") || strings.HasSuffix(name, ".py") || strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".sh") {
				binPath := filepath.Join(repoPath, "bin", name)
				os.Chmod(binPath, 0755)
				binaries = append(binaries, binPath)
			}
		}

	case "Shell":

		entries, _ := os.ReadDir(repoPath)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == repoName || strings.Contains(strings.ToLower(name), "install") || strings.HasSuffix(name, ".sh") {
				binPath := filepath.Join(repoPath, name)
				os.Chmod(binPath, 0755)
				binaries = append(binaries, binPath)
			}
		}

		binEntries, _ := os.ReadDir(filepath.Join(repoPath, "bin"))
		for _, e := range binEntries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == repoName || strings.Contains(strings.ToLower(name), "install") || strings.HasSuffix(name, ".sh") {
				binPath := filepath.Join(repoPath, "bin", name)
				os.Chmod(binPath, 0755)
				binaries = append(binaries, binPath)
			}
		}

	case "Ruby":

		binPath := filepath.Join(repoPath, "bin", repoName)
		if _, err := os.Stat(binPath); err == nil {
			os.Chmod(binPath, 0755)
			binaries = append(binaries, binPath)
		} else {

			binEntries, _ := os.ReadDir(filepath.Join(repoPath, "bin"))
			for _, e := range binEntries {
				if e.IsDir() {
					continue
				}
				full := filepath.Join(repoPath, "bin", e.Name())
				os.Chmod(full, 0755)
				binaries = append(binaries, full)
			}
		}

	case "C/C++":

		candidates := []string{
			filepath.Join(repoPath, repoName),
			filepath.Join(repoPath, "bin", repoName),
			filepath.Join(repoPath, "build", repoName),
		}
		for _, c := range candidates {
			if isExecutable(c) {
				binaries = append(binaries, c)
				break
			}
		}

		if len(binaries) == 0 {
			scanDirs := []string{
				repoPath,
				filepath.Join(repoPath, "bin"),
				filepath.Join(repoPath, "build"),
			}
			for _, dir := range scanDirs {
				entries, _ := os.ReadDir(dir)
				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					full := filepath.Join(dir, e.Name())
					if isExecutable(full) {
						binaries = append(binaries, full)
					}
				}
			}
		}
	}

	for _, b := range binaries {
		linkPath := filepath.Join(binDir, filepath.Base(b))

		os.Remove(linkPath)
		err := os.Symlink(b, linkPath)
		if err != nil {
			fmt.Println("Failed to link binary", b, ":", err)
			continue
		}
		fmt.Println("Linked", filepath.Base(b), "to", binDir)
	}

	if len(binaries) == 0 {
		fmt.Println("No binaries found to link for", repoName)
	}
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

		extra := ""
		if m.Language != "" && m.Language != "Unknown" {
			extra = fmt.Sprintf(" [%s]", m.Language)
			if m.Built {
				extra += " ✓"
			} else {
				extra += " ✗"
			}
		}
		fmt.Printf("- %s (%s)%s\n", m.Name, m.Repo, extra)
	}
}

func updateRepo(name string) {
	manifestPath := filepath.Join(manifestsDir, name+".json")
	pkgPath := filepath.Join(packagesDir, name)

	// Check if installed
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		fmt.Println("Package not installed:", name)
		return
	}

	// Load manifest
	var m Manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Println("Failed to read manifest:", err)
		return
	}
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Println("Failed to parse manifest:", err)
		return
	}

	fmt.Println("Updating", name, "...")

	// Git pull
	cmd := exec.Command("git", "pull")
	cmd.Dir = pkgPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Git pull failed:", err)
		return
	}

	// Re-detect language and rebuild if it was built before
	m.Language = detectLanguage(pkgPath)
	if m.Built && m.Language != "Unknown" {
		fmt.Println("Rebuilding...")
		success, buildCmd := autoBuildRepo(pkgPath, m.Language)
		m.Built = success
		m.BuildCmd = buildCmd
	}

	m.InstalledAt = time.Now()
	saveManifest(m)
	fmt.Println("Updated", name)
}

// showInfo displays detailed package information
func showInfo(name string) {
	manifestPath := filepath.Join(manifestsDir, name+".json")

	var m Manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Println("Package not found:", name)
		return
	}
	if err := json.Unmarshal(data, &m); err != nil {
		fmt.Println("Failed to parse manifest:", err)
		return
	}

	fmt.Println("Package:", m.Name)
	fmt.Println("Repository:", m.Repo)
	fmt.Println("URL:", m.URL)
	if m.Language != "" {
		fmt.Println("Language:", m.Language)
	}
	fmt.Println("Built:", m.Built)
	if m.BuildCmd != "" {
		fmt.Println("Build Command:", m.BuildCmd)
	}
	fmt.Println("Installed:", m.InstalledAt.Format("2006-01-02 15:04:05"))

	pkgPath := filepath.Join(packagesDir, name)
	if _, err := os.Stat(pkgPath); err == nil {
		fmt.Println("Location:", pkgPath)
	}
}

func saveManifest(m Manifest) {
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(manifestsDir, m.Name+".json"), data, 0644)
}
