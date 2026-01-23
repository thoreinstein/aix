package prompt

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/resource"
)

func TestSelectResource_EmptyList(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := NewSelectorWithIO(strings.NewReader(""), &buf)

	_, err := s.SelectResource("test", nil)
	if err == nil {
		t.Fatal("expected error for empty list")
	}
	if !strings.Contains(err.Error(), "no resources") {
		t.Errorf("expected ErrNoResources, got: %v", err)
	}
}

func TestSelectResource_SingleItem(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := NewSelectorWithIO(strings.NewReader(""), &buf)

	resources := []resource.Resource{
		{Name: "deploy", RepoName: "official-skills"},
	}

	result, err := s.SelectResource("deploy", resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "deploy" {
		t.Errorf("expected 'deploy', got %q", result.Name)
	}
	// Should not prompt for single item
	if buf.Len() > 0 {
		t.Errorf("expected no output for single item, got: %s", buf.String())
	}
}

func TestSelectResource_ValidSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantIdx  int
		wantName string
	}{
		{
			name:     "explicit first",
			input:    "1\n",
			wantIdx:  0,
			wantName: "deploy",
		},
		{
			name:     "explicit second",
			input:    "2\n",
			wantIdx:  1,
			wantName: "deploy",
		},
		{
			name:     "default on empty",
			input:    "\n",
			wantIdx:  0,
			wantName: "deploy",
		},
		{
			name:     "whitespace trimmed",
			input:    "  2  \n",
			wantIdx:  1,
			wantName: "deploy",
		},
	}

	resources := []resource.Resource{
		{Name: "deploy", RepoName: "official-skills"},
		{Name: "deploy", RepoName: "my-custom-repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			s := NewSelectorWithIO(strings.NewReader(tt.input), &buf)

			result, err := s.SelectResource("deploy", resources)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, result.Name)
			}
			if result.RepoName != resources[tt.wantIdx].RepoName {
				t.Errorf("expected repo %q, got %q", resources[tt.wantIdx].RepoName, result.RepoName)
			}
		})
	}
}

func TestSelectResource_InvalidSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "too low",
			input:   "0\n",
			wantErr: "out of range",
		},
		{
			name:    "too high",
			input:   "3\n",
			wantErr: "out of range",
		},
		{
			name:    "negative",
			input:   "-1\n",
			wantErr: "out of range",
		},
		{
			name:    "not a number",
			input:   "abc\n",
			wantErr: "not a number",
		},
	}

	resources := []resource.Resource{
		{Name: "deploy", RepoName: "official-skills"},
		{Name: "deploy", RepoName: "my-custom-repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			s := NewSelectorWithIO(strings.NewReader(tt.input), &buf)

			_, err := s.SelectResource("deploy", resources)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestSelectResource_Cancelled(t *testing.T) {
	t.Parallel()

	// Empty reader simulates EOF (Ctrl+D)
	var buf bytes.Buffer
	r := &eofReader{}
	s := NewSelectorWithIO(r, &buf)

	resources := []resource.Resource{
		{Name: "deploy", RepoName: "official-skills"},
		{Name: "deploy", RepoName: "my-custom-repo"},
	}

	_, err := s.SelectResource("deploy", resources)
	if err == nil {
		t.Fatal("expected error for EOF")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected ErrSelectionCancelled, got: %v", err)
	}
}

func TestSelectResource_OutputFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	s := NewSelectorWithIO(strings.NewReader("1\n"), &buf)

	resources := []resource.Resource{
		{Name: "deploy", RepoName: "official-skills"},
		{Name: "deploy", RepoName: "my-custom-repo"},
	}

	_, err := s.SelectResource("deploy", resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify output format
	if !strings.Contains(output, `Multiple resources found for "deploy":`) {
		t.Errorf("missing header in output: %s", output)
	}
	if !strings.Contains(output, "[1] deploy (official-skills)") {
		t.Errorf("missing first option in output: %s", output)
	}
	if !strings.Contains(output, "[2] deploy (my-custom-repo)") {
		t.Errorf("missing second option in output: %s", output)
	}
	if !strings.Contains(output, "Select [1]:") {
		t.Errorf("missing prompt in output: %s", output)
	}
}

// eofReader simulates immediate EOF (like Ctrl+D).
type eofReader struct{}

func (r *eofReader) Read(_ []byte) (int, error) {
	return 0, io.EOF
}
