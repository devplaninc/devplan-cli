package logging

import (
	"os"
	"path/filepath"

	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"gopkg.in/natefinch/lumberjack.v2"
)
import "log/slog"

func GetLogFile() (string, error) {
	configDir, err := prefs.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "cli.log"), nil
}

func Setup() error {
	fileName, err := GetLogFile()
	if err != nil {
		return err
	}
	logFile := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    10, // MB
		MaxBackups: 2,
		MaxAge:     28, // days
		Compress:   true,
	}
	level := slog.LevelInfo
	if len(os.Getenv("DEVPLAN_DEBUG")) > 0 {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return nil
}
