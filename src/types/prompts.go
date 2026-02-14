package types

// PromptArgument describes a parameter accepted by a prompt.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptDefinition describes a single MCP prompt.
type PromptDefinition struct {
	Name        string           `json:"name"`
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptMessage is a single message in a prompt response.
type PromptMessage struct {
	Role    string       `json:"role"`
	Content ContentBlock `json:"content"`
}

// PromptHandler generates prompt messages from given arguments.
// Returns (description, messages, error).
type PromptHandler func(args map[string]string) (string, []PromptMessage, error)

// Prompt pairs a definition with its handler.
type Prompt struct {
	Definition PromptDefinition
	Handler    PromptHandler
}
