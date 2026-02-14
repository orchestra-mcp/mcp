package helpers

import "fmt"

// ValidateArgs checks that required fields are present and types match.
func ValidateArgs(args, props map[string]any, required []string) error {
	for _, req := range required {
		if _, ok := args[req]; !ok {
			return fmt.Errorf("missing required parameter: %s", req)
		}
	}
	for name, prop := range props {
		val, ok := args[name]
		if !ok {
			continue
		}
		propMap, ok := prop.(map[string]any)
		if !ok {
			continue
		}
		expected, _ := propMap["type"].(string)
		if err := checkType(name, val, expected); err != nil {
			return err
		}
	}
	return nil
}

func checkType(name string, val any, expected string) error {
	switch expected {
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("parameter %s must be a string", name)
		}
	case "number":
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("parameter %s must be a number", name)
		}
	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("parameter %s must be a boolean", name)
		}
	case "array":
		if _, ok := val.([]any); !ok {
			return fmt.Errorf("parameter %s must be an array", name)
		}
	case "object":
		if _, ok := val.(map[string]any); !ok {
			return fmt.Errorf("parameter %s must be an object", name)
		}
	}
	return nil
}
