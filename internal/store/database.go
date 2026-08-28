package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketStudents    = []byte("students")
	bucketAssignments = []byte("assignments")
	bucketSubmissions = []byte("submissions")
	bucketSnapshots   = []byte("grade_snapshots")
	bucketMetadata    = []byte("metadata")
)

type Database struct {
	mu     sync.RWMutex
	db     *bolt.DB
	path   string
	closed bool
}

func Open(path string) (*Database, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	cleaned := filepath.Clean(path)
	parent := filepath.Dir(cleaned)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := bolt.Open(cleaned, 0o600, &bolt.Options{NoGrowSync: false})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	database := &Database{db: db, path: cleaned}
	if err := database.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return database, nil
}

func (d *Database) initialize() error {
	return d.update(func(tx *bolt.Tx) error {
		buckets := [][]byte{bucketStudents, bucketAssignments, bucketSubmissions, bucketSnapshots, bucketMetadata}
		for _, name := range buckets {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		metadata := tx.Bucket(bucketMetadata)
		if metadata.Get([]byte("schema_version")) == nil {
			if err := metadata.Put([]byte("schema_version"), []byte("1")); err != nil {
				return fmt.Errorf("write schema version: %w", err)
			}
		}
		return nil
	})
}

func (d *Database) Path() string {
	if d == nil {
		return ""
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.path
}

func (d *Database) Close() error {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	if d.db == nil {
		return nil
	}
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}

func (d *Database) IsClosed() bool {
	if d == nil {
		return true
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.closed
}

func (d *Database) view(fn func(*bolt.Tx) error) error {
	if d == nil {
		return fmt.Errorf("database is nil")
	}
	if fn == nil {
		return fmt.Errorf("view callback is nil")
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed || d.db == nil {
		return fmt.Errorf("database is closed")
	}
	if err := d.db.View(fn); err != nil {
		return fmt.Errorf("database view: %w", err)
	}
	return nil
}

func (d *Database) update(fn func(*bolt.Tx) error) error {
	if d == nil {
		return fmt.Errorf("database is nil")
	}
	if fn == nil {
		return fmt.Errorf("update callback is nil")
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed || d.db == nil {
		return fmt.Errorf("database is closed")
	}
	if err := d.db.Update(fn); err != nil {
		return fmt.Errorf("database update: %w", err)
	}
	return nil
}

func (d *Database) Backup(path string) error {
	if path == "" {
		return fmt.Errorf("backup path is required")
	}
	cleaned := filepath.Clean(path)
	if cleaned == d.Path() {
		return fmt.Errorf("backup path must differ from database path")
	}
	if err := os.MkdirAll(filepath.Dir(cleaned), 0o755); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}
	return d.view(func(tx *bolt.Tx) error {
		if err := tx.CopyFile(cleaned, 0o600); err != nil {
			return fmt.Errorf("copy backup: %w", err)
		}
		return nil
	})
}

func (d *Database) SchemaVersion() (string, error) {
	var version string
	err := d.view(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketMetadata)
		if bucket == nil {
			return errors.New("metadata bucket is missing")
		}
		value := bucket.Get([]byte("schema_version"))
		if value == nil {
			return errors.New("schema version is missing")
		}
		version = string(value)
		return nil
	})
	return version, err
}
