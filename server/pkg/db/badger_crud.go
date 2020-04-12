package db

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
)

const (
	seqBandwidth = 10
)

func (b *BadgerContext) newSequence(kind string, cache uint64) error {
	seq, err := b.db.GetSequence([]byte(kind), cache)
	if err != nil {
		return err
	}

	// start with 1
	seq.Next()

	b.seqList[kind] = seq

	return nil
}

// GetSequence :
func (b *BadgerContext) GetSequence(kind string, cache uint64) (uint64, error) {
	var seq *badger.Sequence

	seqType := kind
	if seqType == "" {
		seqType = "GLOBAL"
	}

	if b.seqList != nil {
		if _, ok := b.seqList[seqType]; !ok {
			if err := b.newSequence(seqType, cache); err != nil {
				return 0, err
			}

		}
	}

	seq, _ = b.seqList[seqType]

	if seq != nil {
		return seq.Next()
	}
	return 0, fmt.Errorf("Failed to get sequence")
}

// InsertIndex : data insert index
func (b *BadgerContext) InsertIndex(key string) error {

	if b.KeyExists(key) {
		return nil
	}

	var body []byte
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), body)
	})
}

// Insert : data insert routine
func (b *BadgerContext) Insert(key string, value interface{}) error {

	if b.KeyExists(key) {
		return fmt.Errorf("key already exists")
	}

	k := []byte(key)

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, jsonValue)
	})

	return err
}

// InsertAndTrack : Insert and Track with a trigger
func (b *BadgerContext) InsertAndTrack(key string, value interface{}, templates ...interface{}) error {

	err := b.Insert(key, value)
	if err != nil {
		return err
	}

	if b.listener != nil && templates != nil {
		for _, template := range templates {
			b.listener.TrackEvent(value, template)
		}
	}

	return nil
}

// List :
func (b *BadgerContext) List(prefix []byte) ([][]byte, error) {
	var rows [][]byte
	// disable full row scans
	if prefix == nil {
		return rows, fmt.Errorf("Must choose a prefix when listing rows")
	}

	err := b.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()

		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()
			data, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			rows = append(rows, data)
		}
		return nil
	})

	return rows, err
}

// ListKeys :
func (b *BadgerContext) ListKeys(prefix []byte) ([][]byte, error) {
	var keys [][]byte
	// disable full row scans
	if prefix == nil {
		return keys, fmt.Errorf("Must choose a prefix when listing rows")
	}

	err := b.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()

		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()
			key := item.KeyCopy(nil)

			keys = append(keys, key)
		}
		return nil
	})

	return keys, err
}

// Get : Data getter
func (b *BadgerContext) Get(key string) ([]byte, error) {

	var data []byte
	var err error
	k := []byte(key)

	err = b.db.View(func(tx *badger.Txn) error {

		item, err := tx.Get(k)
		if err != nil {
			return ErrRecNotFound{}
		}

		if item == nil {
			return ErrRecNotFound{}
		}

		data, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return data, fmt.Errorf("invalid key: does not exist")
	}

	return data, err
}

// Upsert : Data row merge
func (b *BadgerContext) Upsert(key string, value interface{}) error {

	if b.KeyExists(key) {
		return b.Update(key, value)
	}

	return b.Insert(key, value)

}

// Update : data row update
func (b *BadgerContext) Update(key string, value interface{}) error {
	k := []byte(key)

	if !b.KeyExists(key) {
		return ErrRecNotFound{}
	}

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *badger.Txn) error {
		err := tx.Set(k, jsonValue)
		fmt.Println("err:", err)
		return err
	})

}

// UpdateAndTrack : Updates with a trigger
func (b *BadgerContext) UpdateAndTrack(key string, value interface{}, templates ...interface{}) error {

	err := b.Update(key, value)
	if err != nil {
		return err
	}

	if b.listener != nil && templates != nil {
		for _, template := range templates {
			b.listener.TrackEvent(value, template)
		}
	}

	return nil
}

// KeyExists : check if a key exists
func (b *BadgerContext) KeyExists(key string) bool {
	var exsts bool = false
	k := []byte(key)

	err := b.db.View(func(txn *badger.Txn) error {
		v, err := txn.Get(k)

		if err != nil {
			return err
		}

		if v != nil {
			exsts = true
			return nil
		}

		exsts = false
		return nil
	})

	if err != nil {
		return false
	}

	return exsts
}

// DeleteIfExists : Delete a row if it exists
func (b *BadgerContext) DeleteIfExists(key string) error {

	k := []byte(key)

	return b.db.Update(func(tx *badger.Txn) error {
		err := tx.Delete(k)
		return err
	})
}

// SoftDelete : Changes key to deleted: so it is not found when actual
// key search occurs but exists in DB
func (b *BadgerContext) SoftDelete(key string) error {
	var err error
	k := []byte(key)

	txn := b.db.NewTransaction(true)
	defer txn.Discard()

	item, err := txn.Get(k)
	if err != nil {
		return fmt.Errorf("failed to get value for key %s, err: %v", key, err)
	}

	v, err := item.ValueCopy(nil)
	if err != nil {
		return err
	}

	err = txn.Set([]byte("deleted: "+key), v)
	if err != nil {
		return err
	}

	err = txn.Delete(k)

	return txn.Commit(func(error) {})
}

// Delete : Delete only if key exists. should be soft delete?
func (b *BadgerContext) Delete(key string) error {
	return b.DeleteIfExists(key)
}

// GetListener : Get Listener for row changes
func (b *BadgerContext) GetListener() ChangeListener {
	return b.listener
}

// Close : close the db
func (b *BadgerContext) Close() {
	if b.listener != nil {
		b.listener.Close()
	}
	defer b.db.Close()
}
