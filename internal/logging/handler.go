package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/fatih/color"

	"github.com/thoreinstein/aix/internal/doctor"
)

// Handler implements slog.Handler for TTY-optimized text output.
// It provides colorized output when the writer supports it.
type Handler struct {
	opts   slog.HandlerOptions
	out    io.Writer
	mu     *sync.Mutex
	attrs  []slog.Attr
	groups []string

	// Colors
	timeColor  *color.Color
	debugColor *color.Color
	infoColor  *color.Color
	warnColor  *color.Color
	errorColor *color.Color
	keyColor   *color.Color
}

// NewHandler creates a new TTY-optimized text handler.
func NewHandler(out io.Writer, opts *slog.HandlerOptions) *Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	h := &Handler{
		opts: *opts,
		out:  out,
		mu:   &sync.Mutex{},
	}

	// Only initialize colors if the writer supports them
	if SupportsColor(out) {
		h.timeColor = color.New(color.FgHiBlack)
		h.debugColor = color.New(color.FgMagenta)
		h.infoColor = color.New(color.FgGreen)
		h.warnColor = color.New(color.FgYellow)
		h.errorColor = color.New(color.FgRed, color.Bold)
		h.keyColor = color.New(color.FgCyan)
	}

	return h
}

// Enabled reports whether the handler handles records at the given level.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle handles the Record.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. Time
	if !r.Time.IsZero() {
		t := r.Time.Format(time.Kitchen)
		if h.timeColor != nil {
			t = h.timeColor.Sprint(t)
		}
		fmt.Fprintf(h.out, "%s ", t)
	}

	// 2. Level
	levelStr := r.Level.String()
	if h.timeColor != nil { // use timeColor as proxy for "useColor"
		switch {
		case r.Level >= slog.LevelError:
			levelStr = h.errorColor.Sprint(levelStr)
		case r.Level >= slog.LevelWarn:
			levelStr = h.warnColor.Sprint(levelStr)
		case r.Level >= slog.LevelInfo:
			levelStr = h.infoColor.Sprint(levelStr)
		default:
			levelStr = h.debugColor.Sprint(levelStr)
		}
	}
	fmt.Fprintf(h.out, "%-5s ", levelStr)

	// 3. Message
	fmt.Fprintf(h.out, "%s", r.Message)

	// 4. Attributes (from WithAttrs)
	for _, a := range h.attrs {
		h.appendAttr(a)
	}

	// 5. Attributes (from Record)
	r.Attrs(func(a slog.Attr) bool {
		h.appendAttr(a)
		return true
	})

	fmt.Fprintln(h.out)

	return nil
}

func (h *Handler) appendAttr(a slog.Attr) {
	key := a.Key
	if h.keyColor != nil {
		key = h.keyColor.Sprint(key)
	}

	value := a.Value.Any()

	// Redact sensitive values
	if doctor.ShouldMask(a.Key) {
		value = doctor.MaskValue(fmt.Sprint(value))
	} else if strVal, ok := value.(string); ok {
		if doctor.ContainsTokenPrefix(strVal) {
			value = doctor.MaskValue(strVal)
		}
	}

	fmt.Fprintf(h.out, " %s=%v", key, value)
}

// WithAttrs returns a new Handler with the given attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newH := *h
	// Shallow copy is fine because we don't modify the slices in place,
	// but we should create a new slice to avoid side effects if multiple
	// loggers are derived from the same handler.
	newH.attrs = make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newH.attrs, h.attrs)
	copy(newH.attrs[len(h.attrs):], attrs)
	return &newH
}

// WithGroup returns a new Handler with the given group name.
// Currently groups are implemented by prefixing keys.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newH := *h
	newH.groups = make([]string, len(h.groups)+1)
	copy(newH.groups, h.groups)
	newH.groups[len(h.groups)] = name
	return &newH
}
