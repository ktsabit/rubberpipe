package destinations

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalAdapter struct {
	BaseDir string
}

type LocalConfig struct {
	BaseDir string `json:"base_dir"`
}

func NewLocalAdapter(cfg LocalConfig) *LocalAdapter {
	return &LocalAdapter{BaseDir: cfg.BaseDir}
}

func (l *LocalAdapter) Store(srcPath string) (string, error) {
	filename := filepath.Base(srcPath)
	destPath := filepath.Join(l.BaseDir, filename)

	if err := os.MkdirAll(l.BaseDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy data: %w", err)
	}

	return filename, nil
}

func (l *LocalAdapter) Retrieve(fileName string) (string, error) {
	path := filepath.Join(l.BaseDir, fileName)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}
	return path, nil
}
