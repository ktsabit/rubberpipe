// internal/hub.go
package internal

import (
	"database/sql"
	"fmt"
)

type SourceAdapter interface {
	Backup() (string, error)
	Validate() error
	Restore(string) error
}

type DestinationAdapter interface {
	Store(srcPath string) (string, error)
	Retrieve(string) (string, error)
}

// Function signatures for creating a Adapter from config JSON
type SourceFactory func(configJSON string) (SourceAdapter, error)
type DestinationFactory func(configJSON string) (DestinationAdapter, error)

var sourceFactories = make(map[string]SourceFactory)
var destinationFactories = make(map[string]DestinationFactory)

func RegisterSourceAdapter(typ string, factory SourceFactory) {
	if _, exists := sourceFactories[typ]; exists {
		panic(fmt.Sprintf("Source adapter type '%s' already registered", typ))
	}
	sourceFactories[typ] = factory
}

func RegisterDestinationAdapter(typ string, factory DestinationFactory) {
	if _, exists := destinationFactories[typ]; exists {
		panic(fmt.Sprintf("Destination adapter type '%s' already registered", typ))
	}
	destinationFactories[typ] = factory
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

		// 1. Look up the factory in the Source registry
		if factory, ok := sourceFactories[typ]; ok {
			adapter, err := factory(cfgJSON)
			if err != nil {
				return nil, fmt.Errorf("failed to load source adapter '%s' of type '%s': %w", name, typ, err)
			}
			hub.Sources[name] = adapter
			continue
		}

		// 2. Look up the factory in the Destination registry
		if factory, ok := destinationFactories[typ]; ok {
			adapter, err := factory(cfgJSON)
			if err != nil {
				return nil, fmt.Errorf("failed to load destination adapter '%s' of type '%s': %w", name, typ, err)
			}
			hub.Destinations[name] = adapter
			continue
		}

		// If execution reaches here, the type is unknown
		return nil, fmt.Errorf("unknown adapter type '%s' for config '%s'. Has the package been imported?", typ, name)
	}

	return hub, nil
}

// Backup from source adapter to a temporary file to be then stored with dest adapter
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

	return storedPath, nil
}

func (h *Hub) Restore(backupID int, db *sql.DB) error {
	// Query backup record
	var srcName, destName, file string
	err := db.QueryRow("SELECT source, destination, filename FROM backups WHERE id = ?", backupID).
		Scan(&srcName, &destName, &file)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	src, ok := h.Sources[srcName]
	if !ok {
		return fmt.Errorf("source adapter %s not found", srcName)
	}

	dest, ok := h.Destinations[destName]
	if !ok {
		return fmt.Errorf("destination adapter %s not found", destName)
	}

	localFile, err := dest.Retrieve(file)
	if err != nil {
		return fmt.Errorf("failed to retrieve backup: %w", err)
	}

	if err := src.Restore(localFile); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Println("Restore successful:", backupID)
	return nil
}
