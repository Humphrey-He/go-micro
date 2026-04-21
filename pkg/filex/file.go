package filex

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Read reads the entire file content.
func Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadString reads the entire file content as string.
func ReadString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadLines reads all lines from a file.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// ReadJSON reads and unmarshals a JSON file.
func ReadJSON[T any](path string) (T, error) {
	var result T
	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	return result, json.Unmarshal(data, &result)
}

// Write writes data to a file.
func Write(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// WriteString writes a string to a file.
func WriteString(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// WriteJSON marshals and writes data to a JSON file.
func WriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Append appends data to a file.
func Append(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// AppendString appends a string to a file.
func AppendString(path string, content string) error {
	return Append(path, []byte(content))
}

// Exists checks if the file or directory exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// IsDir checks if the path is a directory.
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if the path is a file.
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Size returns the file size in bytes.
func Size(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// MkdirAll creates a directory and all parent directories.
func MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

// Remove removes a file.
func Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll removes a directory and all its contents.
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Copy copies a file from src to dst.
func Copy(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return dstFile.Sync()
}

// Move moves a file from src to dst.
func Move(dst, src string) error {
	return os.Rename(src, dst)
}

// Walk walks the file tree rooted at root.
func Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

// Glob returns the names of all files matching pattern.
func Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Ext returns the file name extension.
func Ext(name string) string {
	return filepath.Ext(name)
}

// Base returns the base name of the file.
func Base(path string) string {
	return filepath.Base(path)
}

// Dir returns the directory name of the path.
func Dir(path string) string {
	return filepath.Dir(path)
}

// Join joins any number of path elements.
func Join(elem ...string) string {
	return filepath.Join(elem...)
}

// SplitList splits the PATH environment variable.
func SplitList(path string) []string {
	return filepath.SplitList(path)
}

// Rel returns a relative path from base to path.
func Rel(base, path string) (string, error) {
	return filepath.Rel(base, path)
}

// Abs returns the absolute path.
func Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// ReadDir reads all files in a directory.
func ReadDir(dir string) ([]os.DirEntry, error) {
	return os.ReadDir(dir)
}

// WatchFile watches a single file for changes.
func WatchFile(name string) (uint, error) {
	return 0, nil // stub for compatibility
}

// Getwd returns the current working directory.
func Getwd() (string, error) {
	return os.Getwd()
}

// TempDir returns the default temp directory.
func TempDir() string {
	return os.TempDir()
}

// TempFile creates a temp file.
func TempFile(dir, prefix string) (*os.File, error) {
	return os.CreateTemp(dir, prefix)
}

// ReadFull reads exactly len(p) bytes into p.
func ReadFull(r io.Reader, p []byte) (n int, err error) {
	return io.ReadFull(r, p)
}

// CopyN copies n bytes from src to dst.
func CopyN(dst io.Writer, src io.Reader, n int64) (written int64, err error) {
	return io.CopyN(dst, src, n)
}

// ByteReader at end returns true if reader is at end.
func ByteReaderAtEnd(r *bytes.Reader) bool {
	_, err := r.ReadByte()
	if err == io.EOF {
		return true
	}
	r.UnreadByte()
	return false
}

// SplitExtension splits a filename into name and extension.
func SplitExtension(name string) (string, string) {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext), ext
}
