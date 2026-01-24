package logging

import (
	"os"
	"testing"
)

func TestSupportsColor(t *testing.T) {
	tests := []struct {
		name  string
		env   map[string]string
		isTTY bool
		want  bool
	}{
		{
			name:  "NO_COLOR prevents color",
			env:   map[string]string{"NO_COLOR": "1"},
			isTTY: true,
			want:  false,
		},
		{
			name:  "TERM=dumb prevents color",
			env:   map[string]string{"TERM": "dumb"},
			isTTY: true,
			want:  false,
		},
		{
			name:  "non-TTY prevents color",
			env:   map[string]string{},
			isTTY: false,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant env vars
			os.Unsetenv("NO_COLOR")
			os.Unsetenv("TERM")

			// Set test env vars
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			// We use the internal function to test env logic independently of real TTY detection.
			var w mockWriter
			got := supportsColor(&w, tt.isTTY)
			if got != tt.want {
				t.Errorf("supportsColor() = %v, want %v (env=%v, isTTY=%v)", got, tt.want, tt.env, tt.isTTY)
			}
		})
	}
}

func TestIsTTY_NonFile(t *testing.T) {
	var w mockWriter
	if IsTTY(&w) != false {
		t.Error("IsTTY should return false for mockWriter")
	}
}

type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
