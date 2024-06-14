package logger

import (
	"context"
	"log/slog"
	"os"

	"cloud.google.com/go/logging"
	"github.com/pkg/errors"
)

var (
	SeverityDefault = slog.Level(logging.Default)
	SeverityInfo    = slog.Level(logging.Info)
	SeverityWarning = slog.Level(logging.Warning)
	SeverityError   = slog.Level(logging.Error)
)

type Logger struct {
	*slog.Logger
}

func New() *Logger {
	replacer := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.MessageKey {
			a.Key = "message"
		}

		if a.Key == slog.LevelKey {
			a.Key = "severity"
			a.Value = slog.StringValue(logging.Severity(a.Value.Any().(slog.Level)).String())
		}

		if a.Key == slog.SourceKey {
			a.Key = "logging.googleapis.com/sourceLocation"
		}

		return a
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replacer,
	})
	slog.SetDefault(slog.New(h))
	return &Logger{slog.New(h)}
}

func (l *Logger) LogErrorWithStack(ctx context.Context, err error) {
	if er, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
		l.Log(ctx, slog.Level(logging.Error), err.Error(), "stack_trace", er.StackTrace())
	} else {
		l.Log(ctx, slog.Level(logging.Error), err.Error())
	}
}
