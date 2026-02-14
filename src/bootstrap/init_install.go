package bootstrap

import (
	"embed"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
)

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
