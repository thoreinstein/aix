package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestExitError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ExitError
		want string
	}{
		{
			name: "with underlying error",
			err:  NewExitError(ErrNotFound, ExitGeneral),
			want: "resource not found",
		},
		{
			name: "with wrapped error",
			err:  NewExitError(fmt.Errorf("loading config: %w", ErrInvalidConfig), ExitUsage),
			want: "loading config: invalid configuration",
		},
		{
			name: "nil underlying error",
			err:  NewExitError(nil, ExitMisuse),
			want: "exit code 64",
		},
		{
			name: "success code with error",
			err:  NewExitError(errors.New("unexpected"), ExitSuccess),
			want: "unexpected",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ExitError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExitError_Unwrap(t *testing.T) {
	tests := []struct {
		name       string
		err        *ExitError
		wantTarget error
		wantIs     bool
	}{
		{
			name:       "unwrap to sentinel error",
			err:        NewExitError(ErrNotFound, ExitGeneral),
			wantTarget: ErrNotFound,
			wantIs:     true,
		},
		{
			name:       "unwrap through wrapped error",
			err:        NewExitError(fmt.Errorf("skill loading: %w", ErrMissingName), ExitUsage),
			wantTarget: ErrMissingName,
			wantIs:     true,
		},
		{
			name:       "no match for different sentinel",
			err:        NewExitError(ErrNotFound, ExitGeneral),
			wantTarget: ErrInvalidConfig,
			wantIs:     false,
		},
		{
			name:       "nil underlying error",
			err:        NewExitError(nil, ExitGeneral),
			wantTarget: ErrNotFound,
			wantIs:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.wantTarget); got != tt.wantIs {
				t.Errorf("errors.Is() = %v, want %v", got, tt.wantIs)
			}
		})
	}
}

func TestExitError_As(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantAs   bool
	}{
		{
			name:     "direct ExitError",
			err:      NewExitError(ErrNotFound, ExitGeneral),
			wantCode: ExitGeneral,
			wantAs:   true,
		},
		{
			name:     "wrapped ExitError",
			err:      fmt.Errorf("command failed: %w", NewExitError(ErrInvalidConfig, ExitUsage)),
			wantCode: ExitUsage,
			wantAs:   true,
		},
		{
			name:     "ExitMisuse code",
			err:      NewExitError(ErrInvalidToolSyntax, ExitMisuse),
			wantCode: ExitMisuse,
			wantAs:   true,
		},
		{
			name:     "non-ExitError",
			err:      ErrNotFound,
			wantCode: 0,
			wantAs:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var exitErr *ExitError
			gotAs := errors.As(tt.err, &exitErr)
			if gotAs != tt.wantAs {
				t.Errorf("errors.As() = %v, want %v", gotAs, tt.wantAs)
			}
			if gotAs && exitErr.Code != tt.wantCode {
				t.Errorf("ExitError.Code = %d, want %d", exitErr.Code, tt.wantCode)
			}
		})
	}
}

func TestNewExitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     int
		wantErr  error
		wantCode int
	}{
		{
			name:     "with sentinel error",
			err:      ErrNotFound,
			code:     ExitGeneral,
			wantErr:  ErrNotFound,
			wantCode: ExitGeneral,
		},
		{
			name:     "with nil error",
			err:      nil,
			code:     ExitSuccess,
			wantErr:  nil,
			wantCode: ExitSuccess,
		},
		{
			name:     "with custom error",
			err:      errors.New("custom error"),
			code:     ExitMisuse,
			wantErr:  errors.New("custom error"),
			wantCode: ExitMisuse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewExitError(tt.err, tt.code)
			if got.Code != tt.wantCode {
				t.Errorf("NewExitError().Code = %d, want %d", got.Code, tt.wantCode)
			}
			if tt.wantErr == nil {
				if got.Err != nil {
					t.Errorf("NewExitError().Err = %v, want nil", got.Err)
				}
			} else {
				if got.Err == nil {
					t.Errorf("NewExitError().Err = nil, want %v", tt.wantErr)
				} else if got.Err.Error() != tt.wantErr.Error() {
					t.Errorf("NewExitError().Err = %q, want %q", got.Err.Error(), tt.wantErr.Error())
				}
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "ErrMissingName",
			err:     ErrMissingName,
			wantMsg: "name is required",
		},
		{
			name:    "ErrNotFound",
			err:     ErrNotFound,
			wantMsg: "resource not found",
		},
		{
			name:    "ErrInvalidConfig",
			err:     ErrInvalidConfig,
			wantMsg: "invalid configuration",
		},
		{
			name:    "ErrInvalidToolSyntax",
			err:     ErrInvalidToolSyntax,
			wantMsg: "invalid tool syntax",
		},
		{
			name:    "ErrUnknownTool",
			err:     ErrUnknownTool,
			wantMsg: "unknown tool",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("%s.Error() = %q, want %q", tt.name, got, tt.wantMsg)
			}
		})
	}
}

func TestExitCodeConstants(t *testing.T) {
	tests := []struct {
		name string
		code int
		want int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitGeneral", ExitGeneral, 1},
		{"ExitUsage", ExitUsage, 2},
		{"ExitMisuse", ExitMisuse, 64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}

func TestErrorWrappingChain(t *testing.T) {
	// Test a realistic error wrapping scenario
	baseErr := ErrInvalidConfig
	wrappedOnce := fmt.Errorf("parsing skill file: %w", baseErr)
	wrappedTwice := fmt.Errorf("loading skill 'test': %w", wrappedOnce)
	exitErr := NewExitError(wrappedTwice, ExitUsage)

	// errors.Is should find the sentinel through the chain
	if !errors.Is(exitErr, ErrInvalidConfig) {
		t.Error("errors.Is() should find ErrInvalidConfig through wrapping chain")
	}

	// errors.As should find ExitError
	var target *ExitError
	if !errors.As(exitErr, &target) {
		t.Error("errors.As() should find ExitError")
	}
	if target.Code != ExitUsage {
		t.Errorf("ExitError.Code = %d, want %d", target.Code, ExitUsage)
	}

	// Error message should contain the full chain
	want := "loading skill 'test': parsing skill file: invalid configuration"
	if got := exitErr.Error(); got != want {
		t.Errorf("ExitError.Error() = %q, want %q", got, want)
	}
}
