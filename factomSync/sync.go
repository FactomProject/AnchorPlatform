package factomSync

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/AccumulateNetwork/SMT/smt"
	"github.com/FactomProject/factom"
	"github.com/PaulSnow/AnchorMaker/database"
)

var MS smt.MerkleState
var Stop bool

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
		return 0
	}

	databaseHeight, _ = database.BytesInt64(db.Get(database.DBlockBucket, []byte("head"))) // Convert bytes to int64
	return databaseHeight                                                                  // Return the height
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

	if hashBytes, err := hex.DecodeString(hash); err != nil { // Convert the hash string to a binary hash
		return err // return any errors
	} else {
		var hBytesArray [32]byte
		copy(hBytesArray[:], hashBytes)
		MS.AddToChain(hBytesArray)
		// Batch this key value pair
		_ = db.PutBatch(batch, database.ObjectBucket, database.Int64Bytes(dbheight), hashBytes)
	}

	return nil
}

// AddDirectoryBlock()
// Add the Directory Block hash to the database and to the Merkle tree.  We also add all the other blocks
// to the database that the Directory Block references, i.e. the Admin Block, the Entry Credit Block, the
// Factoid Block, and the various Entry Blocks for chains modified in this directory block
func AddDirectoryBlock(db *database.DB, batch *database.Batch, dbheight int64) (err error) {
	if dBlock, err := factom.GetDBlockByHeight(dbheight); err != nil {
		return fmt.Errorf("failed to read the dBlock at height %d\n", dbheight)
	} else {

		// Add the hash of the directory block to the database
		if err := AddHash(db, batch, dBlock.KeyMR, dbheight); err != nil {
			return fmt.Errorf("could not decode directory block dbkeymr %s at height %d", dBlock.KeyMR, dbheight)
		}

		// We put in all the directory block entries.  This includes the hash for the admin block, entry credit block,
		// and factoid block.
		for i, v := range dBlock.DBEntries {
			if err := AddHash(db, batch, v.KeyMR, dbheight); err != nil {
				return fmt.Errorf("database failure on DBlock entry %s at height %d", v.KeyMR, dbheight)
			}
			if i >= 3 {
				if err := AddEntryBlock(db, batch, dbheight, v.KeyMR); err != nil {
					return fmt.Errorf("failure with entry block %s : %v", v.KeyMR, err)
				}
			}
		}

		// Now add all the entries of the entry credit block to the database
		if err = AddFactoidBlock(db, batch, dbheight, dBlock); err != nil {
			return fmt.Errorf("could not add the Factoid Block : %v ", err)
		}

		return nil
	}
}

// AddEntryCreditBlock()
// Unlike the old anchoring support, the Anchor Platform will allow receipts for commits made to the Factom protocol
func AddFactoidBlock(db *database.DB, batch *database.Batch, dbheight int64, dBlock *factom.DBlock) (err error) {

	// Pull the Factoid Block from the Directory Block
	if FBlock, err := factom.GetFBlock(dBlock.DBEntries[2].KeyMR); err != nil {
		return fmt.Errorf("retrieve of factoid block %s failed at height %d", dBlock.DBEntries[2].KeyMR, dbheight)
	} else {

		// Add all the factoid transactions to the database and MerkleTree
		for _, v := range FBlock.Transactions {
			if err := AddHash(db, batch, v.TxID, dbheight); err != nil {
				return fmt.Errorf("database failure on factoid block %s at height %d",
					dBlock.DBEntries[1].KeyMR,
					dbheight)
			}
		}

		return nil
	}
}

// AddEntryBlock
// Add the entries in the given entry block to the database
func AddEntryBlock(db *database.DB, batch *database.Batch, dbheight int64, keyMR string) (err error) {

	// Pull the Entry Block from the Directory Block
	if entryBlock, err := factom.GetEBlock(keyMR); err != nil {
		return fmt.Errorf("could not pull the entry block at %s", keyMR)
	} else {
		// Add all the entries added to Factom for this chain in this block
		for _, v := range entryBlock.EntryList {
			// Add the hash to the database
			if err := AddHash(db, batch, v.EntryHash, dbheight); err != nil {
				// Fail on an error
				return fmt.Errorf("database failure on writing entry %s to the database", v.EntryHash)
			}
		}
	}

	return nil
}

// PrintState
// Helps to see our state from time to time during a long sync.
func PrintState(begin time.Time, dbheight, startHeight, factomdHeight int64) {
	// Figure out how long we have been syncing
	timeSpent := time.Now().Sub(begin)
	// Figure out how much of the missing blocks we have processed
	percentDone := float64(dbheight-startHeight) / float64(factomdHeight)
	// Figure out an estimate of how much longer this is going to take.
	timePerBlock := float64(timeSpent) / float64(dbheight-startHeight)
	timeLeft := int64(float64(factomdHeight-dbheight) * timePerBlock)
	// Print our feedback
	fmt.Printf("%s Height: %10s  Objects: %13s  Done: %02.0f%%  Left: %s \n",
		database.FormatTimeLapse(timeSpent),
		humanize.Comma(dbheight),
		humanize.Comma(MS.GetCount()),
		percentDone*100,
		database.FormatTimeLapseSeconds(int64(timeLeft)))
}

// MakeMerkleBlock
// After every Directory Block is added to the database, we make a Merkle Block and add it to the database
func MakeMerkleBlock(db *database.DB, batch *database.Batch, dbheight int64) {
	// Ending the Block returns the serialization of the Merkle State, and the hash of the Merkle State
	merkleState, MSHash := MS.EndBlock()
	// Put the key/value MSHash/merkleState into the database, so one can look up a merkle state by hash
	_ = db.PutBatch(batch, database.MerkleStateBucket, MSHash[:], merkleState)
	// Put the key/value dbheight/merkleState into the database, so one can look up the merkle state by dbheight
	_ = db.PutBatch(batch, database.MerkleStateBucket, database.Int64Bytes(dbheight), merkleState)
}

// SetMerkleState
// Set the Merkle State to the same state it was at the given dbheight
func SetMerkleState(db *database.DB, dbheight int64) {
	// We have no Merkle State at dbheight 0, making the initial Merkle State the correct one
	if dbheight > 0 {
		// Get the Merkle state for the dbheight
		merkleState := db.Get(database.MerkleStateBucket, database.Int64Bytes(dbheight))
		// Make that merkle state our state
		MS.UnMarshal(merkleState)
	}
}

// Sync
// The Sync process continues as long as the Anchor Platform is running.  When the Anchor Platform is
// behind factomd, then blocks are requested and added to our database.  Furthermore, we grab all the
// anchors written to external chains as they are posted to the Anchor Chain in factom.
func Sync() {
	// Set our hashing function in the Merkle State
	MS.InitSha256()
	// Get our database
	db := database.GetDB()

	defer func() { _ = db.Close() }()

	// Get the database height from the database
	databaseHeight := GetDatabaseHeight(db)
	// Set our merkle state to what we have in the database
	SetMerkleState(db, databaseHeight)

	// On a control-c, we close the database and wait a bit before exiting the program.
	database.AddInterruptHandler(func() {
		Stop = true
		fmt.Println("Exiting Anchor Platform")
		_ = db.Close()
		time.Sleep(20 * time.Second)
		os.Exit(0)
	})

	running := time.Now()

	for {
		// Make sure we do not busy wait
		time.Sleep(10 * time.Second)
		// If we have been signaled to stop, then return out of the syncing process
		// This will close the database
		if Stop {
			return
		}
		// Get the directory block height in factomd and the database and see if we have something to do
		databaseHeight = GetDatabaseHeight(db)
		factomdHeight := getFactomdHeight()
		// If we are caught up, continue (wait 10 seconds before we check again)
		if factomdHeight == databaseHeight {
			continue
		}

		// To estimate how long this is taking and how long it will take, we need the time we started,
		// and the databaseHeight we started at.
		begin := time.Now()
		startHeight := databaseHeight

		// Writing to the database is really slow if each key/value pair is written by itself, so we
		// batch the writes.
		batch := db.BeginBatch()

		// Update the user where we are
		fmt.Printf("Height: %s  Objects: %s  Running time: %s\n",
			humanize.Comma(databaseHeight),
			humanize.Comma(MS.GetCount()),
			database.FormatTimeLapse(time.Now().Sub(running)))

		for dbheight := databaseHeight + 1; !Stop && dbheight <= factomdHeight; dbheight++ {

			// Process the directory block at the height dbheight
			if err := AddDirectoryBlock(db, batch, dbheight); err != nil {
				panic(err)
			}
			// We make a Merkle Block for every Directory Block
			MakeMerkleBlock(db, batch, dbheight)

			// Every so often we are going to give feedback to the user about syncing.  But only if
			// we are syncing more than 100 blocks.  Otherwise, we just do our job and let the UI provide
			// feedback.  Always print the first time.
			if dbheight != 0 && dbheight%1000 == 0 && factomdHeight-dbheight >= 10 {
				PrintState(begin, dbheight, startHeight, factomdHeight)
			}

			// Every 10k key pairs, write all the key values to the database.  Note that batch can be reused.
			// Toward the end (less than 10 blocks to go) then end the batch after every directory block processed.
			if len(*batch) > 1000 || factomdHeight-dbheight < 10 {
				SetDatabaseHeight(batch, db, dbheight)
				db.EndBatch(batch)
			}
		}
	}
}
