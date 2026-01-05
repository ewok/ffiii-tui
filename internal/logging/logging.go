/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger with the specified configuration.
// If debug is true, uses development config with debug level logging.
// outputPaths specifies where to write logs; if empty, uses default outputs.
// Returns the logger, a cleanup function, and any error encountered.
func New(debug bool, outputPaths ...string) (*zap.Logger, func(), error) {
	var config zap.Config

	if debug {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	if len(outputPaths) > 0 {
		config.OutputPaths = outputPaths
		config.ErrorOutputPaths = outputPaths
	}

	logger, err := config.Build()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = logger.Sync()
	}

	return logger, cleanup, nil
}
