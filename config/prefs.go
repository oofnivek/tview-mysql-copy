package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const DefaultTheme = "dark"

type Preferences struct {
	Theme string `json:"theme"` // "dark" or "light"
}

func prefsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".mysql-copy")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "prefs.json"), nil
}

func LoadPrefs() (*Preferences, error) {
	path, err := prefsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Preferences{Theme: DefaultTheme}, nil
	}
	if err != nil {
		return nil, err
	}
	var p Preferences
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.Theme == "" {
		p.Theme = DefaultTheme
	}
	return &p, nil
}

func SavePrefs(p *Preferences) error {
	path, err := prefsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
