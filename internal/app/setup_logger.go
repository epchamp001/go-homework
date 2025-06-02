package app

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"pvz-cli/internal/config"
	"pvz-cli/pkg/logger"
)

// SetupLogger настраивает и возвращает экземпляр логгера на основе конфигурации.
func SetupLogger(cfg config.LoggingConfig) (logger.Logger, error) {
	lvl, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}

	var sampling *zap.SamplingConfig
	if cfg.Sampling != nil {
		sampling = &zap.SamplingConfig{
			Initial:    cfg.Sampling.Initial,
			Thereafter: cfg.Sampling.Thereafter,
		}
	}

	opts := []logger.Option{
		logger.WithMode(cfg.Mode),
		logger.WithLevel(lvl),
		logger.WithEncoding(cfg.Encoding),
		logger.WithDisableCaller(cfg.DisableCaller),
		logger.WithDisableStacktrace(cfg.DisableStacktrace),
		logger.WithOutputPaths(cfg.OutputPaths...),
		logger.WithErrorOutputPaths(cfg.ErrorOutputPaths...),
		logger.WithEncoderConfig(func(ec *zapcore.EncoderConfig) {
			ec.TimeKey = cfg.TimestampKey
			if cfg.CapitalizeLevel {
				ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
			}
		}),
	}

	if sampling != nil {
		opts = append(opts, logger.WithSampling(sampling))
	}

	if len(cfg.InitialFields) > 0 {
		opts = append(opts, logger.WithInitialFields(cfg.InitialFields))
	}

	return logger.NewLogger(opts...)
}
