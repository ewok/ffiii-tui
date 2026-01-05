/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(debug bool, outputPaths ...string) (*zap.Logger, func(), error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if debug {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
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
