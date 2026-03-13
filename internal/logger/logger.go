package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

type Config struct {
	Level    string
	Pretty   bool
	FilePath string
	Service  string
}

func Init(cfg Config) error {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	if cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		output = io.MultiWriter(output, file)
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339

	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Str("service", cfg.Service).
		Logger()

	return nil
}

func With(fields map[string]interface{}) zerolog.Logger {
	l := log.Logger
	for k, v := range fields {
		l = l.With().Interface(k, v).Logger()
	}
	return l
}

func FromContext(ctx *zerolog.Context) zerolog.Logger {
	return ctx.Logger()
}
