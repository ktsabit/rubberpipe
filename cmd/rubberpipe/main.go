package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rubberpipe/rubberpipe/internal"

	_ "github.com/rubberpipe/rubberpipe/adapters/destinations"
	_ "github.com/rubberpipe/rubberpipe/adapters/sources"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rubberpipe <command> [args...]")
		fmt.Println("Commands: backup, restore, list, config")

		return
	}

	db, err := sql.Open("sqlite3", "./rubberpipe.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS backups (
		id INTEGER PRIMARY KEY,
		source TEXT,
		destination TEXT,
		filename TEXT,
		timestamp DATETIME,
		status TEXT,
		error_msg TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS adapter_configs (
		name TEXT PRIMARY KEY,    -- unique adapter handle
		type TEXT,                -- "postgres", "local", etc.
		config_json TEXT          -- JSON blob of adapter-specific settings
	)`)
	if err != nil {
		log.Fatal(err)
	}

	hub, err := internal.NewHub(db)
	if err != nil {
		fmt.Println(err)
	}
	cmd := os.Args[1]

	switch cmd {
	case "backup":
		if len(os.Args) < 4 {
			fmt.Println("Usage: rubberpipe backup <source> <destination>")
			return
		}
		source := os.Args[2]
		dest := os.Args[3]

		file, err := hub.Backup(source, dest)
		if err != nil {
			_ = internal.LogBackup(db, source, dest, file, "failed", err.Error())
			fmt.Println("Backup failed:", err)
		} else {
			_ = internal.LogBackup(db, source, dest, file, "success", "")
			fmt.Println("Backup successful:", file)
		}

	case "restore":
		if len(os.Args) < 3 {
			fmt.Println("Usage: rubberpipe restore <backup_id>")
			return
		}
		backupIdString := os.Args[2]

		backupId, err := strconv.Atoi(backupIdString)
		if err != nil {
			fmt.Printf("Invalid backup_id: %v\n", err)
			return
		}

		err = hub.Restore(backupId, db)

		if err != nil {
			fmt.Println("Restore failed:", err)
		}

	case "list":
		rows, err := db.Query(`SELECT source, destination, filename, timestamp, status, error_msg 
                               FROM backups ORDER BY timestamp DESC`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		fmt.Println("Backup history:")
		for rows.Next() {
			var src, dest, file, status, errMsg string
			var ts string
			_ = rows.Scan(&src, &dest, &file, &ts, &status, &errMsg)
			fmt.Printf("[%s] %s -> %s | %s | %s\n", ts, src, dest, status, errMsg)
		}

	case "config":
		if len(os.Args) < 3 {
			fmt.Println("Usage: rubberpipe config <list|add|remove>")
			return
		}
		sub := os.Args[2]

		switch sub {
		case "list":
			rows, err := db.Query(`SELECT name, type, config_json FROM adapter_configs ORDER BY name`)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			fmt.Println("Adapter configs:")
			for rows.Next() {
				var name, typ, cfgJSON string
				_ = rows.Scan(&name, &typ, &cfgJSON)
				var prettyCfg map[string]interface{}
				json.Unmarshal([]byte(cfgJSON), &prettyCfg)
				// mask password if exists
				if typ == "postgres" {
					if _, ok := prettyCfg["password"]; ok {
						prettyCfg["password"] = "*****"
					}
				}
				prettyBytes, _ := json.MarshalIndent(prettyCfg, "    ", "  ")
				fmt.Printf("- %s (%s):\n%s\n", name, typ, string(prettyBytes))
			}

		case "add":
			if len(os.Args) < 6 {
				fmt.Println("Usage: rubberpipe config add <name> <type> <json>")
				return
			}
			name := os.Args[3]
			typ := os.Args[4]
			cfgJSON := os.Args[5]
			_, err := db.Exec(`INSERT INTO adapter_configs (name, type, config_json) VALUES (?, ?, ?)`, name, typ, cfgJSON)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Adapter config '%s' added.\n", name)

		case "remove":
			if len(os.Args) < 4 {
				fmt.Println("Usage: rubberpipe config remove <name>")
				return
			}
			name := os.Args[3]
			_, err := db.Exec(`DELETE FROM adapter_configs WHERE name = ?`, name)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Adapter config '%s' removed.\n", name)

		default:
			fmt.Println("Unknown config subcommand. Use list, add, or remove.")
		}

	default:
		fmt.Println("Unknown command:", cmd)
	}
}
