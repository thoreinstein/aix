package toolperm

import "fmt"

// ToolPermError represents an error in tool permission syntax.
type ToolPermError struct {
	Token   string // The problematic token
	Message string // Description of the error
}

func (e *ToolPermError) Error() string {
	if e.Token == "" {
		return "tool permission error: " + e.Message
	}
	return fmt.Sprintf("invalid tool permission %q: %s", e.Token, e.Message)
}
