package editor

import (
	"os/exec"
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
