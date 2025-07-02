package log

import (
	lcontext "app/lib/context"
	"app/lib/environment"
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

var (
	logger *slog.Logger
)

func Init() {
	var level slog.Leveler
	if environment.IsDebug() {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

func Debug(c context.Context, messageFormat string, a ...any) {
	log(c, slog.LevelDebug, fmt.Sprintf(messageFormat, a...))
}

func Info(c context.Context, messageFormat string, a ...any) {
	log(c, slog.LevelInfo, fmt.Sprintf(messageFormat, a...))
}

func Warn(c context.Context, messageFormat string, a ...any) {
	log(c, slog.LevelWarn, fmt.Sprintf(messageFormat, a...))
}

func Error(c context.Context, messageFormat string, a ...any) {
	log(c, slog.LevelError, fmt.Sprintf(messageFormat, a...))
}

func log(c context.Context, level slog.Level, message string) {
	requestID, ok := c.Value(lcontext.ContextKeyRequestID).(string)
	if !ok {
		requestID = "-"
	}

	_, file, line, _ := runtime.Caller(2)
	fileName := strings.Split(file, "/")[len(strings.Split(file, "/"))-1]

	logger.LogAttrs(context.Background(), level, message,
		slog.String(lcontext.ContextKeyRequestID.String(), requestID),
		slog.String("caller", fmt.Sprintf("%v:%v", fileName, line)),
	)
}
