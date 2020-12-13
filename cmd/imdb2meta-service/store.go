package main

import (
	"github.com/dgraph-io/badger/v2"
	"go.etcd.io/bbolt"
)

type metaStore struct {
	badgerDB *badger.DB
	boltDB   *bbolt.DB
}

func (s *metaStore) Get(id string) ([]byte, error) {
	var err error
	var metaBytes []byte

	if s.badgerDB != nil {
		err = s.badgerDB.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(id))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					return errNotFound
				}
				return err
			}
			metaBytes, err = item.ValueCopy(nil)
			return err
		})
	} else {
		err = s.boltDB.View(func(tx *bbolt.Tx) error {
			txBytes := tx.Bucket(imdbBytes).Get([]byte(id))
			if txBytes == nil {
				return errNotFound
			}
			copy(metaBytes, txBytes)
			return nil
		})
	}

	return metaBytes, err
}
