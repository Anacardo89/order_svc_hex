package log

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init(lokiHook *LokiHook) {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	Log = Log.With(slog.Group("loki", slog.Any("hook", lokiHook)))
}
