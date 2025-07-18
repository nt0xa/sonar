package lark

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// slogAdapter is our implementation that wraps *slog.Logger.
type slogAdapter struct {
	*slog.Logger
}

// newSlogAdapter creates a new adapter for the given slog.Logger.
func newSlogAdapter(logger *slog.Logger) *slogAdapter {
	return &slogAdapter{logger}
}

// The core logic for parsing variadic arguments.
// It handles the message and converts key-value pairs to slog.Attr.
func (a *slogAdapter) buildMessage(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	// Use a strings.Builder for efficient string concatenation.
	var b strings.Builder
	for i, arg := range args {
		if i > 0 {
			b.WriteString(" ") // Add a space between arguments.
		}
		// fmt.Fprint is a good way to write the string representation
		// of any type to a writer.
		fmt.Fprint(&b, arg)
	}
	return b.String()
}

// Debug implements the Logger interface.
func (a *slogAdapter) Debug(ctx context.Context, args ...any) {
	msg := a.buildMessage(args...)
	if msg == "" {
		return
	}
	a.LogAttrs(ctx, slog.LevelDebug, msg)
}

// Info implements the Logger interface.
func (a *slogAdapter) Info(ctx context.Context, args ...any) {
	msg := a.buildMessage(args...)
	if msg == "" {
		return
	}
	a.LogAttrs(ctx, slog.LevelInfo, msg)
}

// Warn implements the Logger interface.
func (a *slogAdapter) Warn(ctx context.Context, args ...any) {
	msg := a.buildMessage(args...)
	if msg == "" {
		return
	}
	a.LogAttrs(ctx, slog.LevelWarn, msg)
}

// Error implements the Logger interface.
func (a *slogAdapter) Error(ctx context.Context, args ...any) {
	msg := a.buildMessage(args...)
	if msg == "" {
		return
	}
	a.LogAttrs(ctx, slog.LevelError, msg)
}
