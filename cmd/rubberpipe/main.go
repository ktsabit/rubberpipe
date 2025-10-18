package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/rubberpipe/rubberpipe/internal"
)

func main() {
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

	hub := internal.NewHub()

	backupFile, err := hub.Backup("postgres", "local")
	if err != nil {
		internal.LogBackup(db, "postgres", "local", backupFile, "failed", err.Error())
	} else {
		internal.LogBackup(db, "postgres", "local", backupFile, "success", "")
	}
}
