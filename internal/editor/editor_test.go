package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDetectEditor_EnvEditor(t *testing.T) {
	t.Setenv("EDITOR", "nvim")
	t.Setenv("VISUAL", "code")

	got := detectEditor()
	if got != "nvim" {
		t.Errorf("detectEditor() = %q, want %q", got, "nvim")
	}
}

func TestDetectEditor_EnvVisual(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "code")

	got := detectEditor()
	if got != "code" {
		t.Errorf("detectEditor() = %q, want %q", got, "code")
	}
}

func TestDetectEditor_FallbackNano(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")

	got := detectEditor()

	// Should be nano if available, otherwise vi
	if _, err := exec.LookPath("nano"); err == nil {
		if got != "nano" {
			t.Errorf("detectEditor() = %q, want %q (nano available)", got, "nano")
		}
	} else {
		if got != "vi" {
			t.Errorf("detectEditor() = %q, want %q (nano not available)", got, "vi")
		}
	}
}

func TestDetectEditor_EmptyEnvTreatedAsUnset(t *testing.T) {
	// Empty string should be treated as unset
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "vscode")

	got := detectEditor()
	if got != "vscode" {
		t.Errorf("detectEditor() = %q, want %q (empty EDITOR should fall through)", got, "vscode")
	}
}

func TestOpen_Integration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping integration test on windows (uses shell script mock)")
	}

	tmpDir := t.TempDir()
	mockEditor := filepath.Join(tmpDir, "mock-editor.sh")
	outputFile := filepath.Join(tmpDir, "output.txt")

	// Create a mock editor that writes its arguments to a file
	script := "#!/bin/sh\necho \"$@\" > " + outputFile + "\n"
	if err := os.WriteFile(mockEditor, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("EDITOR", mockEditor)

	targetFile := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run Open
	if err := Open(targetFile); err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	// Verify mock editor was called with the right argument
	got, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(got), targetFile) {
		t.Errorf("mock editor output = %q, want it to contain %q", string(got), targetFile)
	}
}

func TestOpen_NoEditor(t *testing.T) {
	// This is hard to test because detectEditor always falls back to "vi"
	// but we can try to force a failure by setting EDITOR to a non-existent binary
	t.Setenv("EDITOR", "non-existent-binary-12345")
	t.Setenv("VISUAL", "")

	err := Open("test.txt")
	if err == nil {
		t.Error("expected error for non-existent editor, got nil")
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		want    []string
		wantErr bool
	}{
		{
			name: "simple command",
			cmd:  "vim",
			want: []string{"vim"},
		},
		{
			name: "command with arguments",
			cmd:  "vim --wait -n",
			want: []string{"vim", "--wait", "-n"},
		},
		{
			name: "double quoted path with spaces",
			cmd:  `"/path/to/My Editor" --wait`,
			want: []string{"/path/to/My Editor", "--wait"},
		},
		{
			name: "single quoted path with spaces",
			cmd:  `'/path/to/My Editor' --wait`,
			want: []string{"/path/to/My Editor", "--wait"},
		},
		{
			name: "mixed quotes",
			cmd:  `"code" '-w' --new-window`,
			want: []string{"code", "-w", "--new-window"},
		},
		{
			name: "escaped quote in double quotes",
			cmd:  `"path with \"quote"`,
			want: []string{`path with "quote`},
		},
		{
			name: "multiple spaces between args",
			cmd:  "vim    --wait    -n",
			want: []string{"vim", "--wait", "-n"},
		},
		{
			name: "tabs as separators",
			cmd:  "vim\t--wait\t-n",
			want: []string{"vim", "--wait", "-n"},
		},
		{
			name:    "unbalanced double quotes",
			cmd:     `"unbalanced`,
			wantErr: true,
		},
		{
			name:    "unbalanced single quotes",
			cmd:     `'unbalanced`,
			wantErr: true,
		},
		{
			name: "empty string",
			cmd:  "",
			want: nil,
		},
		{
			name: "only whitespace",
			cmd:  "   \t  ",
			want: nil,
		},
		{
			name: "VS Code style",
			cmd:  `"/Applications/Visual Studio Code.app/Contents/MacOS/Electron" --wait`,
			want: []string{"/Applications/Visual Studio Code.app/Contents/MacOS/Electron", "--wait"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitCommand(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("splitCommand() = %v (len=%d), want %v (len=%d)", got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitCommand()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
