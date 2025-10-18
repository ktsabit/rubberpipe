// internal/hub.go
package internal

import (
	"database/sql"
	"encoding/json"
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

func NewHub(db *sql.DB) (*Hub, error) {
	hub := &Hub{
		Sources:      make(map[string]SourceAdapter),
		Destinations: make(map[string]DestinationAdapter),
	}

	rows, err := db.Query(`SELECT name, type, config_json FROM adapter_configs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, typ, cfgJSON string
		_ = rows.Scan(&name, &typ, &cfgJSON)

		switch typ {
		case "postgres":
			var cfg sources.PostgresConfig
			json.Unmarshal([]byte(cfgJSON), &cfg)
			hub.Sources[name] = sources.NewPostgresAdapter(cfg)

		case "local":
			var cfg destinations.LocalConfig
			json.Unmarshal([]byte(cfgJSON), &cfg)
			hub.Destinations[name] = destinations.NewLocalAdapter(cfg)
		}
	}

	return hub, nil
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
