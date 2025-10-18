package sources

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"time"

	_ "github.com/lib/pq"
)

type PostgresAdapter struct {
	Host      string
	Port      int
	User      string
	Password  string
	DBName    string
	BackupDir string
}

type PostgresConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBName    string `json:"dbname"`
	BackupDir string `json:"backup_dir"`
}

func NewPostgresAdapter(cfg PostgresConfig) *PostgresAdapter {
	return &PostgresAdapter{
		Host:      cfg.Host,
		Port:      cfg.Port,
		User:      cfg.User,
		Password:  cfg.Password,
		DBName:    cfg.DBName,
		BackupDir: cfg.BackupDir,
	}
}

func (p *PostgresAdapter) Backup() (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	backupFile := fmt.Sprintf("%s/%s.dump", p.BackupDir, timestamp)

	cmd := exec.Command(
		"pg_dump",
		"-h", p.Host,
		"-p", fmt.Sprintf("%d", p.Port),
		"-U", p.User,
		"-F", "c",
		"-f", backupFile,
		p.DBName,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", p.Password))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("pg_dump failed: %w", err)
	}

	return backupFile, nil
}

func (p *PostgresAdapter) Validate() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.DBName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}
