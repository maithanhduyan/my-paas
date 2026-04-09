package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
)

const version = "4.0.0"

// Config stores CLI authentication state.
type Config struct {
	Server       string `json:"server"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "version", "--version", "-v":
		fmt.Printf("mypaas-cli v%s\n", version)
	case "help", "--help", "-h":
		if len(rest) > 0 {
			printCommandHelp(rest[0])
		} else {
			printUsage()
		}
	case "login":
		cmdLogin(rest)
	case "logout":
		cmdLogout()
	case "status":
		cmdStatus()
	case "projects", "ps":
		cmdProjects(rest)
	case "create":
		cmdCreate(rest)
	case "deploy":
		cmdDeploy(rest)
	case "logs":
		cmdLogs(rest)
	case "env":
		cmdEnv(rest)
	case "services", "svc":
		cmdServices(rest)
	case "domains":
		cmdDomains(rest)
	case "orgs":
		cmdOrgs(rest)
	case "keys":
		cmdKeys(rest)
	case "health":
		cmdHealth()
	case "info":
		cmdInfo()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'mypaas help' for usage.\n", cmd)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`My PaaS CLI v` + version + `

Usage: mypaas <command> [options]

Commands:
  login              Login to My PaaS server
  logout             Clear stored credentials
  status             Show current auth status
  health             Server health check
  info               Server info & version

  projects (ps)      List projects
  create             Create a new project
  deploy <id>        Deploy a project
  logs <id>          Stream project logs
  env <id>           Manage environment variables

  services (svc)     List managed services
  domains <id>       List project domains

  orgs               List organizations
  keys               Manage API keys

  version            Show CLI version
  help [command]     Show help for a command

Examples:
  mypaas login --server https://paas.company.com --username admin --password secret
  mypaas login --server https://paas.company.com --api-key mpk_live_xxx
  mypaas ps
  mypaas deploy abc123
  mypaas logs abc123
  mypaas env abc123 set KEY=VALUE
`)
}

func printCommandHelp(cmd string) {
	switch cmd {
	case "login":
		fmt.Print(`Usage: mypaas login [options]

Options:
  --server URL        Server URL (required)
  --username NAME     Username
  --password PASS     Password
  --api-key KEY       API key (alternative to username/password)

Examples:
  mypaas login --server http://localhost:8080 --username admin --password admin
  mypaas login --server https://paas.company.com --api-key mpk_live_xxxxx
`)
	case "deploy":
		fmt.Print(`Usage: mypaas deploy <project-id>

Triggers a new deployment for the specified project.
`)
	case "logs":
		fmt.Print(`Usage: mypaas logs <project-id> [--tail N]

Streams live logs from a project container.
`)
	case "env":
		fmt.Print(`Usage: mypaas env <project-id> [subcommand]

Subcommands:
  (none)              List environment variables
  set KEY=VALUE ...   Set environment variables
  delete KEY          Delete an environment variable
`)
	default:
		fmt.Fprintf(os.Stderr, "No detailed help for '%s'.\n", cmd)
	}
}

// --- Config management ---

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mypaas", "config.json")
}

func loadConfig() *Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return nil
	}
	var cfg Config
	json.Unmarshal(data, &cfg)
	return &cfg
}

func saveConfig(cfg *Config) {
	dir := filepath.Dir(configPath())
	os.MkdirAll(dir, 0o700)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(configPath(), data, 0o600)
}

func clearConfig() {
	os.Remove(configPath())
}

func requireAuth() (*Config, *http.Client) {
	cfg := loadConfig()
	if cfg == nil || (cfg.AccessToken == "" && cfg.RefreshToken == "") {
		fmt.Fprintln(os.Stderr, "Not logged in. Run 'mypaas login' first.")
		os.Exit(1)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	return cfg, client
}

// --- HTTP helpers ---

func apiRequest(cfg *Config, client *http.Client, method, path string, body interface{}) (map[string]interface{}, int, error) {
	url := cfg.Server + "/api" + path

	var bodyReader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(respData, &result)
	return result, resp.StatusCode, nil
}

func apiRequestList(cfg *Config, client *http.Client, path string) ([]interface{}, error) {
	url := cfg.Server + "/api" + path
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("unauthorized — run 'mypaas login' to refresh credentials")
	}

	var list []interface{}
	if err := json.Unmarshal(data, &list); err != nil {
		// Check if it's an error object
		var errResp map[string]interface{}
		if json.Unmarshal(data, &errResp) == nil {
			if msg, ok := errResp["error"]; ok {
				return nil, fmt.Errorf("%v", msg)
			}
		}
		return nil, fmt.Errorf("unexpected response: %s", string(data))
	}
	return list, nil
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func parseFlag(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, flag+"=") {
			return a[len(flag)+1:]
		}
	}
	return ""
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// --- Commands ---

func cmdLogin(args []string) {
	server := parseFlag(args, "--server")
	username := parseFlag(args, "--username")
	password := parseFlag(args, "--password")
	apiKey := parseFlag(args, "--api-key")

	if server == "" {
		fmt.Fprintln(os.Stderr, "Error: --server is required")
		os.Exit(1)
	}
	server = strings.TrimRight(server, "/")

	if apiKey != "" {
		// API key auth — just test it
		cfg := &Config{Server: server, AccessToken: apiKey}
		client := &http.Client{Timeout: 10 * time.Second}
		_, status, err := apiRequest(cfg, client, "GET", "/health", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to %s: %v\n", server, err)
			os.Exit(1)
		}
		if status != 200 {
			fmt.Fprintln(os.Stderr, "Error: server returned non-200 status")
			os.Exit(1)
		}
		saveConfig(cfg)
		fmt.Printf("Logged in to %s with API key\n", server)
		return
	}

	if username == "" || password == "" {
		fmt.Fprintln(os.Stderr, "Error: --username and --password are required (or use --api-key)")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	body := map[string]string{"username": username, "password": password}
	data, _ := json.Marshal(body)

	resp, err := client.Post(server+"/api/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respData, &result)

	if resp.StatusCode != 200 {
		errMsg := getStr(result, "error")
		fmt.Fprintf(os.Stderr, "Login failed: %s\n", errMsg)
		os.Exit(1)
	}

	cfg := &Config{
		Server:       server,
		AccessToken:  getStr(result, "access_token"),
		RefreshToken: getStr(result, "refresh_token"),
	}
	// Fall back to session token if no JWT
	if cfg.AccessToken == "" {
		cfg.AccessToken = getStr(result, "token")
	}

	saveConfig(cfg)
	if user, ok := result["user"].(map[string]interface{}); ok {
		fmt.Printf("Logged in to %s as %s (%s)\n", server, getStr(user, "username"), getStr(user, "role"))
	} else {
		fmt.Printf("Logged in to %s\n", server)
	}
}

func cmdLogout() {
	clearConfig()
	fmt.Println("Logged out. Credentials cleared.")
}

func cmdStatus() {
	cfg := loadConfig()
	if cfg == nil {
		fmt.Println("Not logged in.")
		return
	}
	fmt.Printf("Server:  %s\n", cfg.Server)
	if strings.HasPrefix(cfg.AccessToken, "mpk_") {
		fmt.Printf("Auth:    API key (%s...)\n", cfg.AccessToken[:16])
	} else if cfg.AccessToken != "" {
		fmt.Println("Auth:    JWT token")
	}
}

func cmdHealth() {
	cfg := loadConfig()
	if cfg == nil {
		fmt.Fprintln(os.Stderr, "Not logged in. Run 'mypaas login' first.")
		os.Exit(1)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	result, status, err := apiRequest(cfg, client, "GET", "/health", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status != 200 {
		fmt.Fprintln(os.Stderr, "Server unhealthy")
		os.Exit(1)
	}
	fmt.Printf("Status:  %s\nDocker:  %s\nGo:      %s\n",
		getStr(result, "status"), getStr(result, "docker"), getStr(result, "go"))
}

func cmdInfo() {
	cfg := loadConfig()
	if cfg == nil {
		fmt.Fprintln(os.Stderr, "Not logged in. Run 'mypaas login' first.")
		os.Exit(1)
	}
	fmt.Printf("CLI Version: v%s\nServer:      %s\n", version, cfg.Server)
	client := &http.Client{Timeout: 10 * time.Second}
	result, _, err := apiRequest(cfg, client, "GET", "/health", nil)
	if err == nil {
		fmt.Printf("Go:          %s\nDocker:      %s\n", getStr(result, "go"), getStr(result, "docker"))
	}
}

func cmdProjects(args []string) {
	cfg, client := requireAuth()
	list, err := apiRequestList(cfg, client, "/projects")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(list) == 0 {
		fmt.Println("No projects found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSTATUS\tPROVIDER\tBRANCH")
	for _, item := range list {
		p, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			getStr(p, "id"), getStr(p, "name"), getStr(p, "status"),
			getStr(p, "provider"), getStr(p, "branch"))
	}
	w.Flush()
}

func cmdCreate(args []string) {
	name := parseFlag(args, "--name")
	gitURL := parseFlag(args, "--git-url")
	branch := parseFlag(args, "--branch")

	if name == "" {
		fmt.Fprintln(os.Stderr, "Error: --name is required")
		os.Exit(1)
	}

	cfg, client := requireAuth()
	body := map[string]string{"name": name}
	if gitURL != "" {
		body["git_url"] = gitURL
	}
	if branch != "" {
		body["branch"] = branch
	}

	result, status, err := apiRequest(cfg, client, "POST", "/projects", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "Error: %s\n", getStr(result, "error"))
		os.Exit(1)
	}
	fmt.Printf("Project created: %s (id: %s)\n", getStr(result, "name"), getStr(result, "id"))
}

func cmdDeploy(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: mypaas deploy <project-id>")
		os.Exit(1)
	}
	projectID := args[0]

	cfg, client := requireAuth()
	result, status, err := apiRequest(cfg, client, "POST", "/projects/"+projectID+"/deploy", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "Error: %s\n", getStr(result, "error"))
		os.Exit(1)
	}
	fmt.Printf("Deployment queued: %s (status: %s)\n", getStr(result, "id"), getStr(result, "status"))
}

func cmdLogs(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: mypaas logs <project-id>")
		os.Exit(1)
	}
	projectID := args[0]
	tail := parseFlag(args, "--tail")
	if tail == "" {
		tail = "100"
	}

	cfg, _ := requireAuth()
	url := cfg.Server + "/api/projects/" + projectID + "/logs?tail=" + tail
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 0} // No timeout for streaming
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Error (%d): %s\n", resp.StatusCode, string(data))
		os.Exit(1)
	}

	// Stream SSE lines to stdout
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			line := string(buf[:n])
			// Parse SSE data lines
			for _, l := range strings.Split(line, "\n") {
				l = strings.TrimSpace(l)
				if strings.HasPrefix(l, "data:") {
					data := strings.TrimPrefix(l, "data:")
					data = strings.TrimSpace(data)
					fmt.Println(data)
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "\nStream ended: %v\n", err)
			}
			break
		}
	}
}

func cmdEnv(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: mypaas env <project-id> [set KEY=VALUE | delete KEY]")
		os.Exit(1)
	}
	projectID := args[0]
	rest := args[1:]

	cfg, client := requireAuth()

	if len(rest) == 0 {
		// List env vars
		list, err := apiRequestList(cfg, client, "/projects/"+projectID+"/env")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(list) == 0 {
			fmt.Println("No environment variables set.")
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tVALUE\tSECRET")
		for _, item := range list {
			e, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			secret := ""
			if s, ok := e["is_secret"].(bool); ok && s {
				secret = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", getStr(e, "key"), getStr(e, "value"), secret)
		}
		w.Flush()
		return
	}

	subcmd := rest[0]
	switch subcmd {
	case "set":
		vars := []map[string]interface{}{}
		for _, kv := range rest[1:] {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Invalid format: %s (use KEY=VALUE)\n", kv)
				os.Exit(1)
			}
			vars = append(vars, map[string]interface{}{
				"key":       parts[0],
				"value":     parts[1],
				"is_secret": false,
			})
		}
		if len(vars) == 0 {
			fmt.Fprintln(os.Stderr, "Usage: mypaas env <project-id> set KEY=VALUE ...")
			os.Exit(1)
		}
		body := map[string]interface{}{"vars": vars}
		result, status, err := apiRequest(cfg, client, "PUT", "/projects/"+projectID+"/env", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if status >= 400 {
			fmt.Fprintf(os.Stderr, "Error: %s\n", getStr(result, "error"))
			os.Exit(1)
		}
		fmt.Println("Environment variables updated.")

	case "delete", "rm":
		if len(rest) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: mypaas env <project-id> delete KEY")
			os.Exit(1)
		}
		key := rest[1]
		result, status, err := apiRequest(cfg, client, "DELETE", "/projects/"+projectID+"/env/"+key, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if status >= 400 {
			fmt.Fprintf(os.Stderr, "Error: %s\n", getStr(result, "error"))
			os.Exit(1)
		}
		fmt.Printf("Deleted env var: %s\n", key)

	default:
		fmt.Fprintf(os.Stderr, "Unknown env subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}

func cmdServices(args []string) {
	cfg, client := requireAuth()
	list, err := apiRequestList(cfg, client, "/services")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if len(list) == 0 {
		fmt.Println("No services found.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tSTATUS\tIMAGE")
	for _, item := range list {
		s, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			getStr(s, "id"), getStr(s, "name"), getStr(s, "type"),
			getStr(s, "status"), getStr(s, "image"))
	}
	w.Flush()
}

func cmdDomains(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: mypaas domains <project-id>")
		os.Exit(1)
	}
	projectID := args[0]
	cfg, client := requireAuth()
	list, err := apiRequestList(cfg, client, "/projects/"+projectID+"/domains")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if len(list) == 0 {
		fmt.Println("No domains configured.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDOMAIN\tSSL")
	for _, item := range list {
		d, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		ssl := "no"
		if s, ok := d["ssl_auto"].(bool); ok && s {
			ssl = "auto"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", getStr(d, "id"), getStr(d, "domain"), ssl)
	}
	w.Flush()
}

func cmdOrgs(args []string) {
	cfg, client := requireAuth()
	list, err := apiRequestList(cfg, client, "/organizations")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if len(list) == 0 {
		fmt.Println("No organizations found.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSLUG\tPROJECTS\tSERVICES\tDEPLOYS")
	for _, item := range list {
		o, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			getStr(o, "id"), getStr(o, "name"), getStr(o, "slug"),
			getStr(o, "max_projects"), getStr(o, "max_services"), getStr(o, "max_deployments"))
	}
	w.Flush()
}

func cmdKeys(args []string) {
	cfg, client := requireAuth()

	if len(args) > 0 && args[0] == "create" {
		name := parseFlag(args, "--name")
		scopes := parseFlag(args, "--scopes")
		if name == "" {
			fmt.Fprintln(os.Stderr, "Usage: mypaas keys create --name <name> [--scopes <scopes>]")
			os.Exit(1)
		}
		body := map[string]string{"name": name}
		if scopes != "" {
			body["scopes"] = scopes
		}
		result, status, err := apiRequest(cfg, client, "POST", "/api-keys", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if status >= 400 {
			fmt.Fprintf(os.Stderr, "Error: %s\n", getStr(result, "error"))
			os.Exit(1)
		}
		fmt.Printf("API Key created: %s\n", getStr(result, "name"))
		fmt.Printf("Key:    %s\n", getStr(result, "key"))
		fmt.Printf("Scopes: %s\n", getStr(result, "scopes"))
		fmt.Println("\nSave this key now — it will not be shown again!")
		return
	}

	if len(args) > 0 && args[0] == "delete" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: mypaas keys delete <key-id>")
			os.Exit(1)
		}
		result, status, err := apiRequest(cfg, client, "DELETE", "/api-keys/"+args[1], nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if status >= 400 {
			fmt.Fprintf(os.Stderr, "Error: %s\n", getStr(result, "error"))
			os.Exit(1)
		}
		fmt.Println("API key deleted.")
		return
	}

	// List keys
	list, err := apiRequestList(cfg, client, "/api-keys")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if len(list) == 0 {
		fmt.Println("No API keys found.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPREFIX\tSCOPES\tLAST USED")
	for _, item := range list {
		k, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		lastUsed := getStr(k, "last_used")
		if lastUsed == "" || lastUsed == "<nil>" {
			lastUsed = "never"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			getStr(k, "id"), getStr(k, "name"), getStr(k, "key_prefix"),
			getStr(k, "scopes"), lastUsed)
	}
	w.Flush()
}
