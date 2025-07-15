package llogs

import (
	"fmt"
	"github.com/oullin/env"
	"log/slog"
	"os"
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

	fmt.Println("===>", manager.DefaultPath())

	manager.path = manager.DefaultPath()
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
