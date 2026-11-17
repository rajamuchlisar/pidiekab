package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrPathTraversal = errors.New("directory traversal detected")
	ErrFileNotFound  = errors.New("file not found in storage")
)

type Storage interface {
	Save(ctx context.Context, ext string, r io.Reader) (storageKey string, bytesWritten int64, err error)
	Open(ctx context.Context, storageKey string) (io.ReadCloser, error)
	Delete(ctx context.Context, storageKey string) error
	ResolvePath(storageKey string) (string, error)
}

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storage base dir: %w", err)
	}

	if err := os.MkdirAll(absDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create storage base dir: %w", err)
	}

	return &LocalStorage{baseDir: absDir}, nil
}

// GenerateKey produces a safe non-colliding key: "YYYY/MM/<random-32-hex>.<ext>"
func GenerateKey(ext string) (string, error) {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	if ext == "" {
		ext = "bin"
	}

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	randomHex := hex.EncodeToString(randomBytes)

	now := time.Now().UTC()
	return fmt.Sprintf("%04d/%02d/%s.%s", now.Year(), int(now.Month()), randomHex, ext), nil
}

func (s *LocalStorage) ResolvePath(storageKey string) (string, error) {
	// Reject leading slashes and drive letters across platforms
	trimmed := strings.TrimSpace(storageKey)
	if strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "\\") {
		return "", ErrPathTraversal
	}
	if len(trimmed) >= 2 && trimmed[1] == ':' {
		return "", ErrPathTraversal
	}

	cleanKey := filepath.Clean(filepath.FromSlash(trimmed))
	if filepath.IsAbs(cleanKey) || strings.HasPrefix(cleanKey, "..") {
		return "", ErrPathTraversal
	}

	target := filepath.Join(s.baseDir, cleanKey)
	rel, err := filepath.Rel(s.baseDir, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", ErrPathTraversal
	}

	return target, nil
}

func (s *LocalStorage) Save(ctx context.Context, ext string, r io.Reader) (string, int64, error) {
	key, err := GenerateKey(ext)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate storage key: %w", err)
	}

	targetPath, err := s.ResolvePath(key)
	if err != nil {
		return "", 0, err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
		return "", 0, fmt.Errorf("failed to create directory: %w", err)
	}

	outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0640)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create file on disk: %w", err)
	}
	defer outFile.Close()

	n, err := io.Copy(outFile, r)
	if err != nil {
		_ = os.Remove(targetPath)
		return "", 0, fmt.Errorf("failed to write data: %w", err)
	}

	return filepath.ToSlash(key), n, nil
}

func (s *LocalStorage) Open(ctx context.Context, storageKey string) (io.ReadCloser, error) {
	targetPath, err := s.ResolvePath(storageKey)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return f, nil
}

func (s *LocalStorage) Delete(ctx context.Context, storageKey string) error {
	targetPath, err := s.ResolvePath(storageKey)
	if err != nil {
		return err
	}

	err = os.Remove(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
