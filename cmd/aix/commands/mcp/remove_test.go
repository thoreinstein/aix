package mcp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/thoreinstein/aix/internal/cli"
	climocks "github.com/thoreinstein/aix/internal/cli/mocks"
	"github.com/thoreinstein/aix/internal/errors"
)

func TestFindPlatformsWithMCP(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(t *testing.T) []cli.Platform
		serverName string
		wantCount  int
	}{
		{
			name: "server found on one platform",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)

				m2 := climocks.NewMockPlatform(t)
				m2.EXPECT().GetMCP("github", mock.Anything).Return(nil, errors.New("not found"))

				return []cli.Platform{m1, m2}
			},
			serverName: "github",
			wantCount:  1,
		},
		{
			name: "server found on all platforms",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)

				m2 := climocks.NewMockPlatform(t)
				m2.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)

				return []cli.Platform{m1, m2}
			},
			serverName: "github",
			wantCount:  2,
		},
		{
			name: "server not found on any platform",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().GetMCP("github", mock.Anything).Return(nil, errors.New("not found"))

				m2 := climocks.NewMockPlatform(t)
				m2.EXPECT().GetMCP("github", mock.Anything).Return(nil, errors.New("not found"))

				return []cli.Platform{m1, m2}
			},
			serverName: "github",
			wantCount:  0,
		},
		{
			name:       "no platforms",
			setupMocks: func(t *testing.T) []cli.Platform { return []cli.Platform{} },
			serverName: "github",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platforms := tt.setupMocks(t)
			got := findPlatformsWithMCP(platforms, tt.serverName)
			if len(got) != tt.wantCount {
				t.Errorf("findPlatformsWithMCP() returned %d platforms, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestConfirmMCPRemoval(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		setupMock func(t *testing.T) *climocks.MockPlatform
		want      bool
	}{
		{
			name:  "yes confirms",
			input: "yes\n",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().DisplayName().Return("Claude Code")
				return m
			},
			want: true,
		},
		{
			name:  "y confirms",
			input: "y\n",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().DisplayName().Return("Claude Code")
				return m
			},
			want: true,
		},
		{
			name:  "Y confirms (case insensitive)",
			input: "Y\n",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().DisplayName().Return("Claude Code")
				return m
			},
			want: true,
		},
		{
			name:  "no rejects",
			input: "no\n",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().DisplayName().Return("Claude Code")
				return m
			},
			want: false,
		},
		{
			name:  "empty input rejects (default N)",
			input: "\n",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().DisplayName().Return("Claude Code")
				return m
			},
			want: false,
		},
		{
			name:  "random input rejects",
			input: "maybe\n",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().DisplayName().Return("Claude Code")
				return m
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			in := strings.NewReader(tt.input)

			platforms := []cli.Platform{tt.setupMock(t)}
			got := confirmRemoval(&out, in, "github", platforms)
			if got != tt.want {
				t.Errorf("confirmMCPRemoval() = %v, want %v", got, tt.want)
			}

			// Verify prompt was written
			output := out.String()
			if !strings.Contains(output, "github") {
				t.Error("prompt should contain server name")
			}
			if !strings.Contains(output, "[y/N]") {
				t.Error("prompt should contain [y/N]")
			}
		})
	}
}

func TestConfirmMCPRemoval_ListsPlatforms(t *testing.T) {
	m1 := climocks.NewMockPlatform(t)
	m1.EXPECT().DisplayName().Return("Claude Code")

	m2 := climocks.NewMockPlatform(t)
	m2.EXPECT().DisplayName().Return("OpenCode")

	platforms := []cli.Platform{m1, m2}

	var out bytes.Buffer
	in := strings.NewReader("n\n")

	confirmRemoval(&out, in, "github", platforms)

	output := out.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should list Claude Code")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Error("output should list OpenCode")
	}
}

func TestRemoveCommand_Metadata(t *testing.T) {
	if removeCmd.Use != "remove <name>" {
		t.Errorf("Use = %q, want %q", removeCmd.Use, "remove <name>")
	}

	if removeCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if removeCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}
