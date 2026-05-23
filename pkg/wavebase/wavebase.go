// Copyright 2024, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

// Package wavebase provides core utilities and constants for the Wave Terminal application.
package wavebase

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	// WaveVersion is the current version of Wave Terminal
	WaveVersion = "0.1.0"

	// WaveAppName is the application name used for directories and identifiers
	WaveAppName = "waveterm"

	// WaveHomeDirName is the name of the Wave home directory inside the user's home.
	// Can be overridden at runtime via the WAVETERM_HOME environment variable.
	WaveHomeDirName = ".waveterm"

	// WaveDBFileName is the name of the main database file
	WaveDBFileName = "waveterm.db"

	// WaveHomeDirMode is the permission bits used when creating the Wave home directory.
	// 0700 ensures only the owning user can read/write/execute it.
	WaveHomeDirMode = 0700
)

var (
	waveHomeDir     string
	waveHomeDirOnce sync.Once
	waveHomeDirErr  error
)

// GetWaveHomeDir returns the Wave Terminal home directory path.
// It respects the WAVETERM_HOME environment variable override.
// On first call the path is resolved and cached for subsequent calls.
func GetWaveHomeDir() (string, error) {
	waveHomeDirOnce.Do(func() {
		if override := os.Getenv("WAVETERM_HOME"); override != "" {
			waveHomeDir = override
			return
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			waveHomeDirErr = fmt.Errorf("could not determine user home directory: %w", err)
			return
		}
		waveHomeDir = filepath.Join(homeDir, WaveHomeDirName)
	})
	return waveHomeDir, waveHomeDirErr
}

// EnsureWaveHomeDir creates the Wave home directory if it does not already exist.
func EnsureWaveHomeDir() error {
	dir, err := GetWaveHomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, WaveHomeDirMode); err != nil {
		return fmt.Errorf("failed to create wave home directory %q: %w", dir, err)
	}
	return nil
}

// GetWaveDBPath returns the full path to the Wave Terminal database file.
func GetWaveDBPath() (string, error) {
	dir, err := GetWaveHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, WaveDBFileName), nil
}

// GetPlatform returns a normalised platform string: "darwin", "linux", or "windows".
func GetPlatform() string {
	return runtime.GOOS
}

// IsDevMode returns true when the WAVETERM_DEV environment variable is set to a
// non-empty value, enabling additional logging and debug features.
func IsDevMode() bool {
	return os.Getenv("WAVETERM_DEV") != ""
}
