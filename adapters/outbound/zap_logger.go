package outbound

import (
	"time"

	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	log *zap.Logger
}

// initLogger initializes a zap logger with the given debug flag and RFC 3339 time format
func initLogger(debug bool) *zap.Logger {
	var cfg zap.Config = zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	if debug {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

type ZapLoggerConfig struct {
	Debug bool
}

func NewZapLogger(cfg ZapLoggerConfig) *ZapLogger {
	log := initLogger(cfg.Debug)

	return &ZapLogger{
		log: log,
	}
}

// method to convert "field ...interface{} to "fields ...zap.Field"
func (z *ZapLogger) convertFields(fields ...outbound.LogField) []zap.Field {
	var zapFields []zap.Field
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}

func (z *ZapLogger) Debug(msg string, fields ...outbound.LogField) {
	z.log.Debug(msg, z.convertFields(fields...)...)
}

func (z *ZapLogger) Info(msg string, fields ...outbound.LogField) {
	z.log.Info(msg, z.convertFields(fields...)...)
}

func (z *ZapLogger) Error(msg string, fields ...outbound.LogField) {
	z.log.Error(msg, z.convertFields(fields...)...)
}
