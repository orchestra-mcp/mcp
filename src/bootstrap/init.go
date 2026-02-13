package bootstrap

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed resources/skills
var bundledSkills embed.FS

//go:embed resources/agents
var bundledAgents embed.FS

//go:embed resources/hooks/orchestra-mcp-hook.sh
var bundledHookScript string

//go:embed resources/docs/CLAUDE.md
var bundledClaudeMD string

//go:embed resources/docs/AGENTS.md
var bundledAgentsMD string

//go:embed resources/docs/CONTEXT.md
var bundledContextMD string

// Run initializes Orchestra MCP in the given workspace.
func Run(workspaceRoot string) error {
	abs, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return fmt.Errorf("resolve workspace: %w", err)
	}

	projectName := detectProjectName(abs)
	projectType := detectProjectType(abs)

	if err := writeMcpJSON(abs); err != nil {
		return fmt.Errorf("write .mcp.json: %w", err)
	}

	projDir := filepath.Join(abs, ".projects", projectName)
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		return fmt.Errorf("create .projects: %w", err)
	}

	// Create events directory for hook events
	eventsDir := filepath.Join(abs, ".projects", ".events")
	_ = os.MkdirAll(eventsDir, 0o755)

	claudeDir := filepath.Join(abs, ".claude")

	// Install all bundled skills
	skillCount, err := installEmbedDir(bundledSkills, "resources/skills", filepath.Join(claudeDir, "skills"))
	if err != nil {
		return fmt.Errorf("install skills: %w", err)
	}

	// Install all bundled agents
	agentCount, err := installEmbedDir(bundledAgents, "resources/agents", filepath.Join(claudeDir, "agents"))
	if err != nil {
		return fmt.Errorf("install agents: %w", err)
	}

	// Install hook script
	if err := installHookScript(claudeDir); err != nil {
		return fmt.Errorf("install hooks: %w", err)
	}

	// Merge hooks config into settings.json
	if err := mergeHooksConfig(claudeDir); err != nil {
		return fmt.Errorf("merge hooks config: %w", err)
	}

	// Install project docs (CLAUDE.md, AGENTS.md, CONTEXT.md)
	docCount := installDocs(abs)

	fmt.Printf("Orchestra MCP initialized!\n")
	fmt.Printf("  Project:  %s\n", projectName)
	fmt.Printf("  Type:     %s\n", projectType)
	fmt.Printf("  Root:     %s\n", abs)
	fmt.Printf("  Config:   .mcp.json\n")
	fmt.Printf("  Data:     .projects/%s/\n", projectName)
	fmt.Printf("  Skills:   .claude/skills/ (%d installed)\n", skillCount)
	fmt.Printf("  Agents:   .claude/agents/ (%d installed)\n", agentCount)
	fmt.Printf("  Hooks:    .claude/hooks/orchestra-mcp-hook.sh\n")
	fmt.Printf("  Docs:     CLAUDE.md, AGENTS.md, CONTEXT.md (%d installed)\n", docCount)
	return nil
}

// InstallSkills installs bundled skills to the target directory.
func InstallSkills(target string) (int, error) {
	return installEmbedDir(bundledSkills, "resources/skills", target)
}

// InstallAgents installs bundled agents to the target directory.
func InstallAgents(target string) (int, error) {
	return installEmbedDir(bundledAgents, "resources/agents", target)
}

// InstallDocs installs bundled docs (CLAUDE.md, AGENTS.md, CONTEXT.md) to the workspace root.
// Only writes files that don't already exist to avoid overwriting user customizations.
func InstallDocs(root string) int {
	return installDocs(root)
}

func installDocs(root string) int {
	docs := map[string]string{
		"CLAUDE.md":  bundledClaudeMD,
		"AGENTS.md":  bundledAgentsMD,
		"CONTEXT.md": bundledContextMD,
	}
	count := 0
	for name, content := range docs {
		dest := filepath.Join(root, name)
		if _, err := os.Stat(dest); err == nil {
			continue // don't overwrite existing
		}
		if os.WriteFile(dest, []byte(content), 0o644) == nil {
			count++
		}
	}
	return count
}

// ListBundledSkills returns the names of all bundled skills.
func ListBundledSkills() []string {
	var names []string
	entries, _ := fs.ReadDir(bundledSkills, "resources/skills")
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// ListBundledAgents returns the names of all bundled agents.
func ListBundledAgents() []string {
	var names []string
	entries, _ := fs.ReadDir(bundledAgents, "resources/agents")
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	return names
}

// installEmbedDir walks an embed.FS and copies all files to the target directory.
func installEmbedDir(fsys embed.FS, root, target string) (int, error) {
	count := 0
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		dest := filepath.Join(target, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := fsys.ReadFile(path)
		if err != nil {
			return err
		}
		count++
		return os.WriteFile(dest, data, 0o644)
	})
	return count, err
}

func installHookScript(claudeDir string) error {
	hookDir := filepath.Join(claudeDir, "hooks")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return err
	}
	hookPath := filepath.Join(hookDir, "orchestra-mcp-hook.sh")
	if err := os.WriteFile(hookPath, []byte(bundledHookScript), 0o755); err != nil {
		return err
	}
	return nil
}

func mergeHooksConfig(claudeDir string) error {
	settingsPath := filepath.Join(claudeDir, "settings.json")
	config := map[string]any{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &config)
	}
	if config["hooks"] == nil {
		config["hooks"] = map[string]any{}
	}
	hooks, ok := config["hooks"].(map[string]any)
	if !ok {
		hooks = map[string]any{}
		config["hooks"] = hooks
	}
	hookCmd := ".claude/hooks/orchestra-mcp-hook.sh"
	hookEntry := []any{map[string]any{
		"matcher": "",
		"hooks": []any{map[string]any{
			"type": "command", "command": hookCmd, "async": true,
		}},
	}}
	events := []string{
		"PostToolUse", "Notification",
		"SubagentStart", "SubagentStop",
		"Stop", "SessionStart",
	}
	for _, event := range events {
		if _, exists := hooks[event]; !exists {
			hooks[event] = hookEntry
		}
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, append(data, '\n'), 0o644)
}

func writeMcpJSON(root string) error {
	mcpPath := filepath.Join(root, ".mcp.json")
	config := map[string]any{}
	if data, err := os.ReadFile(mcpPath); err == nil {
		_ = json.Unmarshal(data, &config)
	}
	if config["mcpServers"] == nil {
		config["mcpServers"] = map[string]any{}
	}
	servers := config["mcpServers"].(map[string]any)
	servers["orchestra-mcp"] = map[string]any{
		"command": "orchestra-mcp",
		"args":    []string{"--workspace", root},
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(mcpPath, append(data, '\n'), 0o644)
}

func detectProjectName(root string) string {
	if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		var pkg map[string]any
		if json.Unmarshal(data, &pkg) == nil {
			if name, ok := pkg["name"].(string); ok && name != "" {
				return name
			}
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "module ") {
				parts := strings.Split(strings.TrimPrefix(line, "module "), "/")
				return parts[len(parts)-1]
			}
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "Cargo.toml")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "name") && strings.Contains(line, "=") {
				val := strings.TrimSpace(strings.SplitN(line, "=", 2)[1])
				return strings.Trim(val, "\"'")
			}
		}
	}
	return filepath.Base(root)
}

func detectProjectType(root string) string {
	checks := []struct {
		file     string
		projType string
	}{
		{"package.json", "Node.js"},
		{"go.mod", "Go"},
		{"Cargo.toml", "Rust"},
		{"composer.json", "PHP"},
		{"pyproject.toml", "Python"},
		{"requirements.txt", "Python"},
		{"Gemfile", "Ruby"},
	}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(root, c.file)); err == nil {
			return c.projType
		}
	}
	return "Unknown"
}
