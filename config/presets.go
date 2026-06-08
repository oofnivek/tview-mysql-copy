package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

type Preset struct {
	ID            string `json:"id"`
	SrcConnection string `json:"src_connection"`
	SrcDatabase   string `json:"src_database"`
	SrcTable      string `json:"src_table"`
	DstConnection string `json:"dst_connection"`
	DstDatabase   string `json:"dst_database"`
	DstTable      string `json:"dst_table"`
}

func NewPresetID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type PresetConfig struct {
	Presets []Preset
}

func presetPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".mysql-copy")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "presets.json"), nil
}

func LoadPresets() (*PresetConfig, error) {
	path, err := presetPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &PresetConfig{}, nil
	}
	if err != nil {
		return nil, err
	}
	var presets []Preset
	if err := json.Unmarshal(data, &presets); err != nil {
		return nil, err
	}
	return &PresetConfig{Presets: presets}, nil
}

func SavePresets(pc *PresetConfig) error {
	path, err := presetPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(pc.Presets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
