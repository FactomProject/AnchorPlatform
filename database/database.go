package database

// As a key value store, we are not as yet using too many of the features of any database.
// We use hashes to order entries, which are also organized into buckets, much as LevelDB
// allows.
//
//To use this DB interface, you must allocate a DB
// Then call DB.Init(int)
//
// To set a value in the database, call DB.Put(bucket string, key []byte, value []byte) error
//
// To get a value from the database, call DB.Get(bucket string, key []byte) (value []byte)_
//
// see ValAcc/types/types.go for the constants for bucket names

import (
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v2"
)

type DB struct {
	DBHome   string
	badgerDB *badger.DB
}

type BatchEntry struct {
	Key   []byte
	Value []byte
	Meta  byte
}

type Batch []*BatchEntry

var db *DB

func GetDB() *DB {
	if db != nil {
		return db
	}
	db = new(DB)
	db.InitDB()
	return db
}

// Close
// Close the underlying database
func (d *DB) Close() error {
	return d.badgerDB.Close()
}

// InitDB
// Initialize the database.  This will certainly open an existing database, but will
// also initialize an new, empty database.
func (d *DB) InitDB() {
	// Make sure the home directory exists. If it does, then use it, otherwise go find the home directory.
	if len(d.DBHome) == 0 {
		d.DBHome = fmt.Sprintf("%s%s%03d", GetHomeDir(), "/.AnchorMaker/badger", 0)
	}
	// Make sure all directories exist
	os.MkdirAll(d.DBHome, 0777)
	// Open Badger
	// Try at least three databases before we give up
	db, err := badger.Open(badger.DefaultOptions(d.DBHome))
	if err != nil { // Panic if we can't open Badger
		panic(err)
	}
	d.badgerDB = db // And all is good.
}

// GetKey
// Given a bucket and a key, return the combined key
func GetKey(bucket int, key []byte) (CKey []byte) {
	CKey = append(CKey, Uint32Bytes(uint32(bucket))...)
	CKey = append(CKey, key...)
	return CKey
}

// Get
// Look in the given bucket, and return the key found.  Returns nil if no value
// is found for the given key
func (d *DB) Get(bucket int, key []byte) (value []byte) {
	CKey := GetKey(bucket, key) // combine the bucket and the key
	// Go look up the CKey, and return any error we might find.
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(CKey)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			value = append(value, val...)
			return nil
		})
		return err
	})
	// If anything goes wrong, return nil
	if err != nil {
		return nil
	}
	// If we didn't find the value, we will return a nil here.
	return value
}

func (d *DB) GetInt32(bucket int, ikey uint32) (value []byte) {
	key := Uint32Bytes(ikey)
	return d.Get(bucket, key)
}

// Put
// Put a key/value in the database.  We return an error if there was a problem
// writing the key/value pair to the database.
func (d *DB) Put(bucket int, key []byte, value []byte) error {
	CKey := GetKey(bucket, key)

	// Update the key/value in the database
	err := d.badgerDB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(CKey), value)
		return err
	})
	return err
}

// BeginBatch
// Return a Batch slice. Collect key value pairs to a batch list that will all
// be written to the database at once when the batch is ended.
func (d *DB) BeginBatch() *Batch {
	return new(Batch)
}

// EndBatch
// Write all the transactions in a given batch of pending transactions.  The
// batch is emptied, so it can be reused.
func (d *DB) EndBatch(batch *Batch) {
	txn := db.badgerDB.NewTransaction(true)
	for _, e := range *batch {
		if err := txn.Set(e.Key, e.Value); err != nil {
			panic("Could not write entry")
		}
	}
	if err := txn.Commit(); err != nil {
		panic("Could not write transaction")
	}
	*batch = (*batch)[:0]
}

// PutBatch
// Put a key value pair into a batch.  These commits are effectively pending and will
// only be written to the database if the batch is passed to EndBatch
func (d *DB) PutBatch(batch *Batch, bucket int, key []byte, value []byte) error {
	CKey := GetKey(bucket, key)
	entry := new(BatchEntry)
	entry.Key = CKey
	entry.Value = value
	*batch = append(*batch, entry)
	return nil
}

// PutInt
// Put a key/value in the database, where the key is an index.  We return an error if there was a problem
// writing the key/value pair to the database.
func (d *DB) PutInt32(bucket int, ikey int, value []byte) error {
	key := Uint32Bytes(uint32(ikey))
	return d.Put(bucket, key, value)
}
