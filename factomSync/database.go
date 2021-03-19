package factomSync

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/FactomProject/AnchorPlatform/database"
	"github.com/FactomProject/factom"
)

// getFactomdHeight()
// Return the directory block height in factomd.
// NOTE:  If factomd is offline or we have networking issues of some sort, getFactomdHeight will be blocking.
// Once a height can be retrieved from factomd, getFactomdHeight will return that height.
func getFactomdHeight() (factomdHeight int64) {

	heights, err := factom.GetHeights() // Get all the heights from factomd
	for err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("issue getting directory block height from factomd: %v", err))
		time.Sleep(30 * time.Second)
		if heights, err = factom.GetHeights(); err == nil {
			break
		}
	}
	return heights.DirectoryBlockHeight // Return the Directory Block Height

}

// getDatabaseHeight()
// Return highest directory block currently processed and included in the Anchor Platform Database
func GetDatabaseHeight(db *database.DB) (databaseHeight int64) {

	databaseHeightBytes := db.Get(database.DBlockBucket, []byte("head")) // Get the int64 highest directory block height
	if len(databaseHeightBytes) != 8 {                                   // If we don't find an int64 value, then return zero
		return -1
	}

	databaseHeight, _ = database.BytesInt64(db.Get(database.DBlockBucket, []byte("head"))) // Convert bytes to int64
	return databaseHeight                                                                  // Return the height
}

// GetObjectDbheight
// Get the dbheight of a given object.  If the object is not found in the database, return -1
func GetObjectDbheight(db *database.DB, object []byte) int64 {
	// Pull the value from the database
	value := db.Get(database.ObjectBucket, object[:])
	// If I don't find a dbheight for the object, return -1
	if value == nil {
		return -1
	}
	// Return the int64 value
	dbheight, _ := database.BytesInt64(value)
	return dbheight
}

// SetDatabaseHeight
// Set the Database height in the database, so we can pick up syncing where we left off.
// We don't worry about being exact, because syncing will just overwrite what is already there
// with the same values.
func SetDatabaseHeight(batch *database.Batch, db *database.DB, dbheight int64) {
	_ = db.PutBatch(batch, database.DBlockBucket, []byte("head"), database.Int64Bytes(dbheight))
}

// AddHash()
// Add the Hash to the database and to the Merkle Tree for all elements in factomd
// The factom library handles hashes as strings, but we don't need the bloat, so we convert the strings to binary
// before we add them to the Anchor Platform database.
func AddHash(db *database.DB, batch *database.Batch, hash string, dbheight int64) (err error) {

	// Get the binary for the hash, because that is what will be needed here.
	var hashBytes [32]byte
	if hashSlice, err := hex.DecodeString(hash); err != nil { // Convert the hash string to a binary hash
		return err // return any errors
	} else {
		copy(hashBytes[:], hashSlice)
	}

	// Look for the power of 2 boundary.  We need the hash that spills into the database.Mark to
	// create fast proofs.  That's going to be the HashFunction(MS.pending[database.MarkPower] + hash)
	count := MS.GetCount()
	// Add the hash to our Merkle Tree, no matter what.  However, if this is a duplicate we won't index it.
	MS.AddToMerkleTree(hashBytes)
	// Are we at a mark boundary? count has low bits clear (but ignore count == 0)
	if count > 0 && count&database.MarkMask == 0 {
		// Marshal the Mark State
		data := MS.Marshal()
		// Save the state at the Mark Point
		_ = db.PutBatch(batch,
			database.MarkBucket+database.MerkleState, // Our bucket
			database.Int64Bytes(count),               // The Mark count
			data[:])                                  // The Merkle State at the Mark Point
		// Save the dbheight
		_ = db.PutBatch(batch,
			database.MarkBucket+database.Dbheight, // Our bucket
			database.Int64Bytes(count),            // The mark count
			database.Int64Bytes(dbheight))         // The dbheight holding the Mark Point
		// Save the MerkleState that reflects the adding of the mark

	} else if count > 0 && (count-1)&database.MarkMask == 0 {
		// Add the hash to our Merkle Tree, no matter what.  However, if this is a duplicate we won't index it.
		_ = db.PutBatch(batch,
			database.MarkBucket+database.MarkNext, // Our bucket
			database.Int64Bytes(count),            // This is the Mark Point +1 because it helps us move forward
			hashBytes[:])                          // The dbheight holding the Mark Point
	}

	// Batch this key value pair
	// First check if we have already indexed this object.  If so, we keep the oldest (what is in the database)
	objectDbheight := db.Get(database.ObjectBucket, hashBytes[:])
	if objectDbheight == nil {
		// If we don't have a reference to this object, then add it to the database
		_ = db.PutBatch(batch, database.ObjectBucket, hashBytes[:], database.Int64Bytes(dbheight))
	}

	return nil
}
