package factomSync

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/AccumulateNetwork/SMT/smt"
	"github.com/FactomProject/factom"
	"github.com/PaulSnow/AnchorMaker/database"
)

var MS smt.MerkleState

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

// AddHash()
// Add the Hash to the database and to the Merkle Tree for all elements in factomd
// The factom library handles hashes as strings, but we don't need the bloat, so we convert the strings to binary
// before we add them to the Anchor Platform database.
func AddHash(db *database.DB, hash string, dbheight int64) (err error) {

	if hashBytes, err := hex.DecodeString(hash); err != nil { // Convert the hash string to a binary hash
		return err // return any errors
	} else {
		_ = db.Put(database.ObjectBucket, database.Int64Bytes(dbheight), hashBytes) // Put the hash into the database
	}

	return nil
}

// AddDirectoryBlock()
// Add the Directory Block hash to the database and to the Merkle tree.  We also add all the other blocks
// to the database that the Directory Block references, i.e. the Admin Block, the Entry Credit Block, the
// Factoid Block, and the various Entry Blocks for chains modified in this directory block
func AddDirectoryBlock(db *database.DB, dbheight int64) (err error) {
	if dBlock, err := factom.GetDBlockByHeight(dbheight); err != nil {
		return fmt.Errorf("failed to read the dBlock at height %d\n", dbheight)
	} else {

		// Add the hash of the directory block to the database
		if err := AddHash(db, dBlock.KeyMR, dbheight); err != nil {
			return fmt.Errorf("could not decode directory block dbkeymr %s at height %d", dBlock.KeyMR, dbheight)
		}

		// We put in all the directory block entries.  This includes the hash for the admin block, entry credit block,
		// and factoid block.
		for i, v := range dBlock.DBEntries {
			if err := AddHash(db, v.KeyMR, dbheight); err != nil {
				return fmt.Errorf("database failure on DBlock entry %s at height %d", v.KeyMR, dbheight)
			}
			if i >= 3 {
				if err := AddEntryBlock(db, dbheight, v.KeyMR); err != nil {
					return fmt.Errorf("failure with entry block %s : %v", v.KeyMR, err)
				}
			}
		}

		// Now add all the entries of the entry credit block to the database
		if err = AddFactoidBlock(db, dbheight, dBlock); err != nil {
			return fmt.Errorf("could not add the Factoid Block : %v ", err)
		}

		return nil
	}
}

// AddEntryCreditBlock()
// Unlike the old anchoring support, the Anchor Platform will allow receipts for commits made to the Factom protocol
func AddFactoidBlock(db *database.DB, dbheight int64, dBlock *factom.DBlock) (err error) {

	// Pull the Factoid Block from the Directory Block
	if FBlock, err := factom.GetFBlock(dBlock.DBEntries[2].KeyMR); err != nil {
		return fmt.Errorf("retrieve of factoid block %s failed at height %d", dBlock.DBEntries[2].KeyMR, dbheight)
	} else {

		// Add all the factoid transactions to the database and MerkleTree
		for _, v := range FBlock.Transactions {
			if err := AddHash(db, v.TxID, dbheight); err != nil {
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
func AddEntryBlock(db *database.DB, dbheight int64, keyMR string) (err error) {

	// Pull the Entry Block from the Directory Block
	if entryBlock, err := factom.GetEBlock(keyMR); err != nil {
		return fmt.Errorf("could not pull the entry block at %s", keyMR)
	} else {
		// Add all the entries added to Factom for this chain in this block
		for _, v := range entryBlock.EntryList {
			// Add the hash to the database
			if err := AddHash(db, v.EntryHash, dbheight); err != nil {
				// Fail on an error
				return fmt.Errorf("database failure on writing entry %s to the database", v.EntryHash)
			}
		}
	}

	return nil
}

// Sync
// The Sync process continues as long as the Anchor Platform is running.  When the Anchor Platform is
// behind factomd, then blocks are requested and added to our database.  Furthermore, we grab all the
// anchors written to external chains as they are posted to the Anchor Chain in factom.
func Sync() {

	db := database.GetDB()

	for {
		// Get the directory block height in factomd and in our database
		factomdHeight := getFactomdHeight()
		databaseHeight := GetDatabaseHeight(db)

		now := time.Now()
		startHeight := databaseHeight

		for dbheight := databaseHeight; dbheight <= factomdHeight; dbheight++ {
			if err := AddDirectoryBlock(db, dbheight); err != nil {
				panic(err)
			}
			if dbheight%1000 == 0 && factomdHeight-dbheight > 100 {
				timeSpent := time.Now().Sub(now)
				percentDone := float64(dbheight-startHeight) / float64(factomdHeight)
				t := float64(timeSpent.Seconds()) * (1 - percentDone) / (percentDone + .00001)
				factomdHeight = getFactomdHeight()
				fmt.Printf("%20v | Processed %8d blocks, %4.3f percent complete time left %8.3f hrs\n",
					timeSpent.String(),
					dbheight-startHeight,
					percentDone,
					t/60/60)
			}
		}
	}
}
