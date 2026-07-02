//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dcvix/dcvix-director/internal/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupLogger(cfg config.LogConfig) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Parse log level
	logLevel, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}

	// Set up log rotation
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Directory, "dcvix-director.log"),
		MaxSize:    100,          // megabytes
		MaxBackups: cfg.Rotation, // keep at most 10 rotated files
	}

	// Configure logrus
	logrus.SetLevel(logLevel)
	if logLevel <= logrus.DebugLevel {
		logrus.SetReportCaller(true)
	}
	logrus.SetFormatter(&dcvixFormatter{})
	mw := io.MultiWriter(os.Stdout, logFile)
	logrus.SetOutput(mw)

	return nil
}
