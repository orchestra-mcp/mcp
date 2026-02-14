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
