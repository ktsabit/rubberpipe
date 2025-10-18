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

	return destPath, nil
}
