// Copyright 2024, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package wavebase

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const WaveDBVersion = 1
const InitTimeout = 10 * time.Second

var globalDB *sql.DB
var globalDBOnce sync.Once
var globalDBErr error

// GetDB returns the global database connection, initializing it if necessary.
func GetDB(ctx context.Context) (*sql.DB, error) {
	globalDBOnce.Do(func() {
		dbPath, err := GetWaveDBPath()
		if err != nil {
			globalDBErr = fmt.Errorf("error getting wave db path: %w", err)
			return
		}
		// Added _foreign_keys=on to enforce referential integrity at the SQLite level.
		db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
		if err != nil {
			globalDBErr = fmt.Errorf("error opening wave db: %w", err)
			return
		}
		db.SetMaxOpenConns(1)
		globalDB = db
	})
	if globalDBErr != nil {
		return nil, globalDBErr
	}
	return globalDB, nil
}

// InitDB initializes the database schema if it does not already exist.
func InitDB(ctx context.Context) error {
	db, err := GetDB(ctx)
	if err != nil {
		return err
	}
	initCtx, cancel := context.WithTimeout(ctx, InitTimeout)
	defer cancel()

	tx, err := db.BeginTx(initCtx, nil)
	if err != nil {
		return fmt.Errorf("error beginning init transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(initCtx, `
		CREATE TABLE IF NOT EXISTS db_meta (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("error creating db_meta table: %w", err)
	}

	_, err = tx.ExecContext(initCtx, `
		INSERT OR IGNORE INTO db_meta (key, value)
		VALUES ('version', ?);
	`, WaveDBVersion)
	if err != nil {
		return fmt.Errorf("error inserting db version: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing init transaction: %w", err)
	}

	log.Printf("wavedb initialized at version %d\n", WaveDBVersion)
	return nil
}

// CloseDB closes the global database connection.
func CloseDB() {
	if globalDB != nil {
		globalDB.Close()
		globalDB = nil
	}
}
