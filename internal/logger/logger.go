package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

// Config — настройки логгера
type Config struct {
	Level    string // debug, info, warn, error
	Pretty   bool   // человекочитаемый вывод (для dev)
	FilePath string // путь к файлу лога (опционально)
	Service  string // имя сервиса для поля "service"
}

// Init инициализирует глобальный логгер zerolog
func Init(cfg Config) error {
	// Уровень логирования
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Формат вывода
	var output io.Writer = os.Stdout
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Если указан файл — пишем туда + в stdout
	if cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		output = io.MultiWriter(output, file)
	}

	// Глобальные настройки
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack // стектрейсы для ошибок
	zerolog.TimeFieldFormat = time.RFC3339

	// Инициализация логгера с контекстом
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Str("service", cfg.Service).
		Logger()

	return nil
}

// With — добавляет поля к логгеру (для контекста)
func With(fields map[string]interface{}) zerolog.Logger {
	l := log.Logger
	for k, v := range fields {
		l = l.With().Interface(k, v).Logger()
	}
	return l
}

// FromContext — получает логгер из context (если там есть)
// Можно расширить, если будешь хранить logger в context
func FromContext(ctx zerolog.Context) zerolog.Logger { //nolint:gocritic // zerolog.Context не является указателем — это API библиотеки
	return ctx.Logger()
}
