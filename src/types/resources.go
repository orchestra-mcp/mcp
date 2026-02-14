package types

// ResourceDefinition describes a single MCP resource.
type ResourceDefinition struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceContent is a single content item returned from resources/read.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// ResourceHandler reads a resource by URI and returns its contents.
type ResourceHandler func(uri string) ([]ResourceContent, error)

// Resource pairs a definition with its handler.
type Resource struct {
	Definition ResourceDefinition
	Handler    ResourceHandler
}
