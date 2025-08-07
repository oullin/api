package llogs

import (
	"fmt"
	"github.com/oullin/metal/env"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type FilesLogs struct {
	path   string
	file   *os.File
	logger *slog.Logger
	env    *env.Environment
}

func MakeFilesLogs(env *env.Environment) (Driver, error) {
	manager := FilesLogs{}
	manager.env = env

	manager.path = manager.DefaultPath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(manager.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return FilesLogs{}, fmt.Errorf("failed to create log directory: %w", err)
	}

	resource, err := os.OpenFile(manager.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return FilesLogs{}, err
	}

	handler := slog.New(slog.NewTextHandler(resource, nil))
	slog.SetDefault(handler)

	manager.file = resource
	manager.logger = handler

	return manager, nil
}

func (manager FilesLogs) DefaultPath() string {
	logsEnvironment := manager.env.Logs

	return fmt.Sprintf(
		logsEnvironment.Dir,
		time.Now().UTC().Format(logsEnvironment.DateFormat),
	)
}

func (manager FilesLogs) Close() bool {
	if err := manager.file.Close(); err != nil {
		manager.logger.Error("error closing file: " + err.Error())

		return false
	}

	return true
}
