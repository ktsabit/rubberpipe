package internal

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func LogBackup(db *sql.DB, source, dest, file string, status string, errorMsg string) error {
	_, err := db.Exec(`INSERT INTO backups (source, destination, filename, timestamp, status, error_msg)
                       VALUES (?, ?, ?, ?, ?, ?)`,
		source, dest, file, time.Now(), status, errorMsg)
	return err
}
