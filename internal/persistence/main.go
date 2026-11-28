package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/markusressel/system-control/internal/configuration"
)

var (
	BaseDir = path.Join(configuration.BaseDir, "persistence")
)

// keyFile returns the file backing a key. Keys must be plain file names.
func keyFile(key string) (string, error) {
	if key == "" || key == "." || key == ".." || strings.ContainsAny(key, "/\x00") {
		return "", fmt.Errorf("invalid persistence key %q", key)
	}
	return path.Join(BaseDir, key+".sav"), nil
}

func write(key string, content []byte) error {
	file, err := keyFile(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(BaseDir, 0755); err != nil {
		return err
	}
	return writeFileAtomic(file, content, 0644)
}

func read(key string) ([]byte, error) {
	file, err := keyFile(key)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(file)
}

func SaveInt(key string, value int) error {
	return write(key, []byte(strconv.Itoa(value)))
}

func ReadInt(key string) (int64, error) {
	content, err := read(key)
	if err != nil {
		return -1, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(content)), 0, 64)
}

func SaveFloat(key string, value float64) error {
	return write(key, []byte(strconv.FormatFloat(value, 'f', -1, 64)))
}

func ReadFloat(key string) (float64, error) {
	content, err := read(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(content)), 64)
}

func SaveStruct(key string, value interface{}) error {
	jsonString, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return write(key, jsonString)
}

func ReadStruct(key string, target interface{}) error {
	content, err := read(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, target)
}

// writeFileAtomic replaces the file at path with data, so that concurrent readers see either
// the old or the new content, never a partial one.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	_, err = tmp.Write(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Chmod(tmpName, perm)
	}
	if err == nil {
		err = os.Rename(tmpName, path)
	}
	if err != nil {
		_ = os.Remove(tmpName)
	}
	return err
}
