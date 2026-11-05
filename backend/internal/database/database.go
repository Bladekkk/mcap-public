package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	_path string
}

func New(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	return &DB{DB: db, _path: dbPath}, nil
}

func (d *DB) Path() string {
	return d._path
}

func (d *DB) Close() error {
	return d.DB.Close()
}

// Migrations
const migrationVersion = 1

func (d *DB) Migrate() error {
	// Проверка версии
	var v int
	if err := d.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		return err
	}

	if v == migrationVersion {
		return nil
	}

	tx, err := d.Begin()
	if err != nil {
		return err
	}

	if err := tx.Exec(`CREATE TABLE IF NOT EXISTS servers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		jar_path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'stopped',
		port INTEGER DEFAULT 25565,
		java_version TEXT,
		xms_mb INTEGER DEFAULT 1024,
		xmx_mb INTEGER DEFAULT 2048,
		affinity TEXT,
		autostart BOOLEAN DEFAULT 0,
		eula_accepted BOOLEAN DEFAULT 0
	)`); err != nil {
		tx.Rollback()
		return fmt.Errorf("create servers table: %w", err)
	}

	if err := tx.Exec(`CREATE TABLE IF NOT EXISTS jdk_versions (
		id TEXT PRIMARY KEY,
		version TEXT NOT NULL,
		arch TEXT NOT NULL,
		os TEXT NOT NULL,
		install_path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		tx.Rollback()
		return fmt.Errorf("create jdk_versions table: %w", err)
	}

	if err := tx.Exec(`CREATE TABLE IF NOT EXISTS downloaded_jars (
		id TEXT PRIMARY KEY,
		server_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		url TEXT NOT NULL,
		checksum TEXT,
		size_bytes INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
	)`); err != nil {
		tx.Rollback()
		return fmt.Errorf("create downloaded_jars table: %w", err)
	}

	if err := tx.Exec(`CREATE TABLE IF NOT EXISTS metrics_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		tps REAL,
		mspt REAL,
		online_players INTEGER,
		ram_mb INTEGER,
		cpu_percent REAL,
	FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
	)`); err != nil {
		tx.Rollback()
		return fmt.Errorf("create metrics_history table: %w", err)
	}

	if err := tx.Exec(`CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT,
		action TEXT NOT NULL,
		user TEXT,
		details TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE SET NULL
	)`); err != nil {
		tx.Rollback()
		return fmt.Errorf("create audit_log table: %w", err)
	}

	if err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", migrationVersion)); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
