package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/minio/highwayhash"
	"github.com/pkg/errors"
)

type nullLogger struct {
}

func (l nullLogger) Errorf(string, ...interface{})   {}
func (l nullLogger) Warningf(string, ...interface{}) {}
func (l nullLogger) Infof(string, ...interface{})    {}
func (l nullLogger) Debugf(string, ...interface{})   {}

type DB struct {
	*badger.DB

	key  []byte
	root string
}

func OpenHashDB(dir string, root string) (_ *DB, err error) {
	opts := badger.DefaultOptions(dir).
		WithLogger(&nullLogger{})
	bdb, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	db := &DB{
		DB:   bdb,
		root: root,
	}
	if err := db.init(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) init() error {
	const metadataKey = "_metadataKey"

	txn := db.NewTransaction(true)
	defer txn.Discard()

	item, err := txn.Get([]byte(metadataKey))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			data, err := newMeta(db.root)
			if err != nil {
				return err
			}
			if err := txn.SetEntry(badger.NewEntry([]byte(metadataKey), data)); err != nil {
				return err
			}
			return txn.Commit()
		}
		return err
	}
	var m meta
	if err := item.Value(func(val []byte) error {
		return gob.NewDecoder(bytes.NewReader(val)).Decode(&m)
	}); err != nil {
		return err
	}
	if m.Root != db.root {
		return errors.Errorf("expected root %q, found %q", db.root, m.Root)
	}
	db.key = m.Key
	return nil
}

type meta struct {
	Key     []byte
	Root    string
	Created time.Time
}

func newMeta(root string) ([]byte, error) {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	key := sha256.Sum256(data)
	m := meta{
		Key:     key[:32],
		Root:    root,
		Created: time.Now(),
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (db *DB) Hash(r io.Reader) (n int64, _ []byte, err error) {
	hh, err := highwayhash.New(db.key)
	if err != nil {
		return
	}
	n, err = io.Copy(hh, r)
	if err != nil {
		return
	}
	return n, hh.Sum(nil), nil
}

type File struct {
	ModTime int64
	Hash    []byte
}

func (db *DB) Save(path string, file File) error {
	txn := db.NewTransaction(true)
	defer txn.Discard()

	item, err := txn.Get([]byte(path))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			var buf bytes.Buffer
			if err := gob.NewEncoder(&buf).Encode(file); err != nil {
				return errors.Wrapf(err, "cannot encode")
			}
			if err := txn.SetEntry(badger.NewEntry([]byte(path), buf.Bytes())); err != nil {
				return err
			}
			return txn.Commit()
		}
		return err
	}
	return item.Value(func(val []byte) error {
		var old File
		if err := gob.NewDecoder(bytes.NewReader(val)).Decode(&old); err != nil {
			return errors.Wrapf(err, "cannot decode")
		}
		if !bytes.Equal(old.Hash, file.Hash) {
			if file.ModTime > old.ModTime {
				fmt.Printf("%s: updating hash %x -> %x\n", path, old.Hash, file.Hash)
				var buf bytes.Buffer
				if err := gob.NewEncoder(&buf).Encode(file); err != nil {
					return errors.Wrapf(err, "cannot encode")
				}
				if err := txn.SetEntry(badger.NewEntry([]byte(path), buf.Bytes())); err != nil {
					return err
				}
				return txn.Commit()
			}
			return errors.Errorf("BIT ROT DETECTED: %s is set with %x, but hash is %x\n", path, old.Hash, file.Hash)
		}
		return nil
	})
}
