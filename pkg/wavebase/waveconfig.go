// Copyright 2024, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package wavebase

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const ConfigFileName = "config.json"

// WaveConfig holds the application-level configuration settings.
type WaveConfig struct {
	// Telemetry controls whether anonymous usage data is sent.
	Telemetry bool `json:"telemetry"`

	// UpdateChannel controls the release channel for auto-updates (e.g. "stable", "beta").
	UpdateChannel string `json:"updateChannel,omitempty"`

	// TermFontSize is the default terminal font size in points.
	TermFontSize int `json:"termFontSize,omitempty"`

	// TermFontFamily is the default terminal font family.
	TermFontFamily string `json:"termFontFamily,omitempty"`

	// Theme is the UI color theme (e.g. "dark", "light").
	Theme string `json:"theme,omitempty"`
}

var defaultConfig = WaveConfig{
	Telemetry:      true,
	UpdateChannel:  "stable",
	TermFontSize:   12,
	TermFontFamily: "monospace",
	Theme:          "dark",
}

var (
	configMu     sync.Mutex
	cachedConfig *WaveConfig
)

// GetConfigPath returns the path to the config file inside the Wave home directory.
func GetConfigPath() string {
	return filepath.Join(GetWaveHomeDir(), ConfigFileName)
}

// ReadConfig reads the configuration from disk. If the file does not exist,
// it returns the default configuration without writing anything to disk.
func ReadConfig() (*WaveConfig, error) {
	configMu.Lock()
	defer configMu.Unlock()

	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := defaultConfig
		return &cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	cfg := defaultConfig // start from defaults so missing keys keep their defaults
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	cachedConfig = &cfg
	return &cfg, nil
}

// WriteConfig persists the given configuration to disk, creating the Wave home
// directory if necessary.
func WriteConfig(cfg *WaveConfig) error {
	configMu.Lock()
	defer configMu.Unlock()

	if err := EnsureWaveHomeDir(); err != nil {
		return fmt.Errorf("ensuring wave home dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}

	path := GetConfigPath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file %q: %w", path, err)
	}

	cachedConfig = cfg
	return nil
}

// GetCachedConfig returns the in-memory cached config, or reads it from disk
// if no cached version is available.
func GetCachedConfig() (*WaveConfig, error) {
	configMu.Lock()
	if cachedConfig != nil {
		cfg := *cachedConfig
		configMu.Unlock()
		return &cfg, nil
	}
	configMu.Unlock()
	return ReadConfig()
}
