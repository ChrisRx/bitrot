package hashdb

import (
	"bytes"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
)

type DB struct {
	*badger.DB
}

func Open(dir string) (_ *DB, err error) {
	opts := badger.DefaultOptions(dir).
		WithLogger(&nullLogger{})
	bdb, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &DB{DB: bdb}, nil
}

func (db *DB) Get(path string) (*File, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get([]byte(path))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, errors.Errorf("cannot find file: %q", path)
		}
		return nil, err
	}
	data, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	file := &File{}
	if err := file.Unmarshal(data); err != nil {
		return nil, err
	}
	return file, nil
}

func (db *DB) Save(path string, file File) error {
	txn := db.NewTransaction(true)
	defer txn.Discard()

	item, err := txn.Get([]byte(path))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			data, err := file.Marshal()
			if err != nil {
				return err
			}
			if err := txn.SetEntry(badger.NewEntry([]byte(path), data)); err != nil {
				return err
			}
			return txn.Commit()
		}
		return err
	}
	return item.Value(func(val []byte) error {
		old := &File{}
		if err := old.Unmarshal(val); err != nil {
			return err
		}
		if !bytes.Equal(old.Hash, file.Hash) {
			if file.ModTime > old.ModTime {
				fmt.Printf("%s: updating hash %x -> %x\n", path, old.Hash, file.Hash)
				data, err := file.Marshal()
				if err != nil {
					return err
				}
				if err := txn.SetEntry(badger.NewEntry([]byte(path), data)); err != nil {
					return err
				}
				return txn.Commit()
			}
			return errors.Errorf("BIT ROT DETECTED: %s is set with %x, but hash is %x", path, old.Hash, file.Hash)
		}
		return nil
	})
}
