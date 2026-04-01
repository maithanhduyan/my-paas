package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// App represents a source directory to analyze.
type App struct {
	Source string
}

func NewApp(path string) (*App, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &App{Source: abs}, nil
}

func (a *App) HasFile(name string) bool {
	_, err := os.Stat(filepath.Join(a.Source, name))
	return err == nil
}

func (a *App) ReadFile(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(a.Source, name))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *App) ReadJSON(name string, v any) error {
	data, err := os.ReadFile(filepath.Join(a.Source, name))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ListFiles returns file names in the root directory.
func (a *App) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(a.Source)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}
