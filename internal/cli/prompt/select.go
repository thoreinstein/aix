// Package prompt provides interactive CLI prompts for user input.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/resource"
)

// Sentinel errors for resource selection.
var (
	ErrNoResources        = errors.New("no resources to select from")
	ErrInvalidSelection   = errors.New("invalid selection")
	ErrSelectionCancelled = errors.New("selection cancelled")
)

// Selector handles interactive resource selection prompts.
type Selector struct {
	reader io.Reader
	writer io.Writer
}

// NewSelector creates a new Selector using stdin and stdout.
func NewSelector() *Selector {
	return &Selector{
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

// NewSelectorWithIO creates a Selector with custom reader and writer for testing.
func NewSelectorWithIO(r io.Reader, w io.Writer) *Selector {
	return &Selector{
		reader: r,
		writer: w,
	}
}

// SelectResource prompts the user to choose from a list of resources.
//
// Returns:
//   - ErrNoResources if the list is empty
//   - The resource if only one exists (auto-selects without prompting)
//   - The selected resource based on user input
//   - ErrInvalidSelection if the selection is out of range
//   - ErrSelectionCancelled if input is EOF (e.g., Ctrl+D)
func (s *Selector) SelectResource(query string, resources []resource.Resource) (*resource.Resource, error) {
	if len(resources) == 0 {
		return nil, ErrNoResources
	}

	// Auto-select if only one resource
	if len(resources) == 1 {
		return &resources[0], nil
	}

	// Display selection prompt
	fmt.Fprintf(s.writer, "Multiple resources found for %q:\n", query)
	for i, r := range resources {
		fmt.Fprintf(s.writer, "  [%d] %s (%s)\n", i+1, r.Name, r.RepoName)
	}
	fmt.Fprintf(s.writer, "Select [1]: ")

	// Read user input
	reader := bufio.NewReader(s.reader)
	input, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, ErrSelectionCancelled
		}
		return nil, errors.Wrap(err, "reading selection")
	}

	input = strings.TrimSpace(input)

	// Default to first option if empty
	if input == "" {
		return &resources[0], nil
	}

	// Parse selection number
	selection, err := strconv.Atoi(input)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidSelection, "%q is not a number", input)
	}

	// Validate range (1-indexed)
	if selection < 1 || selection > len(resources) {
		return nil, errors.Wrapf(ErrInvalidSelection, "%d is out of range [1-%d]", selection, len(resources))
	}

	return &resources[selection-1], nil
}

// SelectResourceDefault is a convenience function that uses stdin/stdout.
func SelectResourceDefault(query string, resources []resource.Resource) (*resource.Resource, error) {
	return NewSelector().SelectResource(query, resources)
}
