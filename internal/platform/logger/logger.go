package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// New devuelve un logger con colores en desarrollo (legible en terminal)
// y JSON en cualquier otro entorno (consumible por agregadores de logs).
func New(level, env string) *slog.Logger {
	var l slog.Level
	_ = l.UnmarshalText([]byte(level))
	if env == "development" {
		return slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: l, TimeFormat: time.TimeOnly}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}
