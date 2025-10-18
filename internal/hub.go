// internal/hub.go
package internal

import (
	"fmt"

	"github.com/rubberpipe/rubberpipe/adapters/destinations"
	"github.com/rubberpipe/rubberpipe/adapters/sources"
)

type SourceAdapter interface {
	Backup() (string, error)
	Validate() error
}

type DestinationAdapter interface {
	Store(srcPath string) (string, error)
}

type Hub struct {
	Sources      map[string]SourceAdapter
	Destinations map[string]DestinationAdapter
}

func NewHub() *Hub {
	postgres := &sources.PostgresAdapter{
		Host:      "localhost",
		Port:      5432,
		User:      "postgres",
		Password:  "pass",
		DBName:    "rubberpipe_test",
		BackupDir: "/tmp/rubberpipe",
	}

	local := &destinations.LocalAdapter{
		BaseDir: "backups",
	}

	return &Hub{
		Sources: map[string]SourceAdapter{
			"postgres": postgres,
		},
		Destinations: map[string]DestinationAdapter{
			"local": local,
		},
	}
}

func (h *Hub) Backup(sourceName, destName string) (string, error) {
	src, ok := h.Sources[sourceName]
	if !ok {
		return "", fmt.Errorf("source adapter %s not found", sourceName)
	}
	dest, ok := h.Destinations[destName]
	if !ok {
		return "", fmt.Errorf("destination adapter %s not found", destName)
	}

	if err := src.Validate(); err != nil {
		return "", fmt.Errorf("source validation failed: %w", err)
	}

	backupFile, err := src.Backup()
	if err != nil {
		return "", fmt.Errorf("backup failed: %w", err)
	}

	// Store backup
	storedPath, err := dest.Store(backupFile)
	if err != nil {
		return "", fmt.Errorf("storing backup failed: %w", err)
	}

	fmt.Println("Backup successful:", storedPath)
	return storedPath, nil
}
