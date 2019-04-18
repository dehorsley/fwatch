package db

import (
	"time"

	"github.com/boltdb/bolt"
)

var bucket = []byte("uploadedtime")

// Database stores the uploaded times
type Database struct {
	b *bolt.DB
}

// Open opens a new database
func Open(path string) (*Database, error) {
	b, err := bolt.Open(path, 0660, nil)
	if err != nil {
		return nil, err
	}
	return &Database{b}, nil

}

// Close closes the database connection
func (db *Database) Close() error {
	return db.b.Close()

}

// ForEach iterates over each file in the database calling
// the suplied function with the path and uploaded time
func (db *Database) ForEach(f func(string, time.Time) error) error {
	return db.b.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			t, err := time.Parse(time.RFC3339, string(v))
			if err != nil {
				return err
			}
			return f(string(k), t)
		})
	})
}

// Last returns the last uploaded time of file
func (db *Database) Last(file string) time.Time {
	var t time.Time
	db.b.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		v := b.Get([]byte(file))

		t, _ = time.Parse(time.RFC3339, string(v))
		return nil
	})
	return t
}

// Remove deletes the upload record from the database
func (db *Database) Remove(keys ...string) {
	db.b.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}

		for _, key := range keys {
			b.Delete([]byte(key))
		}
		return nil
	})
}

// Update stores the upload time of the file in the datbase
func (db *Database) Update(t time.Time, keys ...string) {
	tb := []byte(t.Format(time.RFC3339))
	db.b.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists(bucket)

		for _, key := range keys {
			b.Put([]byte(key), tb)
		}
		return nil
	})
}
