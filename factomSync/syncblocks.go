package factomSync

import (
	"encoding/hex"
	"fmt"

	"github.com/AccumulateNetwork/SMT/smt"
	"github.com/FactomProject/AnchorPlatform/database"
	"github.com/FactomProject/factom"
)

var MS smt.MerkleState
var Stop bool

// AddDirectoryBlock()
// Add the Directory Block hash to the database and to the Merkle tree.  We also add all the other blocks
// to the database that the Directory Block references, i.e. the Admin Block, the Entry Credit Block, the
// Factoid Block, and the various Entry Blocks for chains modified in this directory block
func AddDirectoryBlock(db *database.DB, batch *database.Batch, dbheight int64) (err error) {

	var dBlock *factom.DBlock
	if dBlock, err = factom.GetDBlockByHeight(dbheight); err != nil {
		return fmt.Errorf("failed to read the dBlock at height %d\n", dbheight)
	}

	// Add the hash of the directory block to the database
	if err := AddHash(db, batch, dBlock.KeyMR, dbheight); err != nil {
		return fmt.Errorf("could not decode directory block dbkeymr %s at height %d", dBlock.KeyMR, dbheight)
	}

	// We put in all the directory block entries.  This includes the hash for the admin block, entry credit block,
	// and factoid block.
	for i, v := range dBlock.DBEntries {
		// Add all the hashes for the directory block entries, which covers the admin, entry credit, factoid,
		// and entry blocks.
		if err := AddHash(db, batch, v.KeyMR, dbheight); err != nil {
			return fmt.Errorf("database failure on DBlock entry %s at height %d", v.KeyMR, dbheight)
		}
		// Entry blocks start at index 3, so process them separately. Add all the entries in the entry blocks.
		if i >= 3 {
			if err := AddEntryBlock(db, batch, dbheight, v.KeyMR); err != nil {
				return fmt.Errorf("failure with entry block %s : %v", v.KeyMR, err)
			}
		}
	}

	// Now add all the transactions in the Factoid block to the database
	if err = AddFactoidBlock(db, batch, dbheight, dBlock); err != nil {
		return fmt.Errorf("could not add the Factoid Block : %v ", err)
	}

	// Save the MerkleState that results from a directory block, and other directory block indexing
	MakeMerkleBlock(db, batch, dbheight)

	// Note that we don't add the entries in the admin block or the entry credit block.  That could be done
	// in the future, if we so desire, but isn't done here.  One can get receipts to the blocks, but not their
	// individual entries.

	return nil

}

// MakeMerkleBlock
// After every Directory Block is added to the database, we make a Merkle Block and add it to the database
func MakeMerkleBlock(db *database.DB, batch *database.Batch, dbheight int64) {
	// Ending the Block returns the serialization of the Merkle State, and the hash of the Merkle State
	merkleState, _ := MS.EndBlock()
	// Index the Merkle State dbheight/merkleState
	_ = db.PutBatch(batch,
		database.DBlockBucket,
		database.Int64Bytes(dbheight),
		merkleState)
	// Index the Object Count  dbheight/count
	_ = db.PutBatch(batch,
		database.DBlockBucket+database.ObjectCount,
		database.Int64Bytes(dbheight),
		database.Int64Bytes(MS.GetCount()))
}

// AddEntryCreditBlock()
// Unlike the old anchoring support, the Anchor Platform will allow receipts for commits made to the Factom protocol
func AddFactoidBlock(db *database.DB, batch *database.Batch, dbheight int64, dBlock *factom.DBlock) (err error) {

	// Pull the Factoid Block from the Directory Block
	var FBlock *factom.FBlock
	if FBlock, err = factom.GetFBlock(dBlock.DBEntries[2].KeyMR); err != nil {
		return fmt.Errorf("retrieve of factoid block %s failed at height %d", dBlock.DBEntries[2].KeyMR, dbheight)
	}

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

// AddEntryBlock
// Add the entries in the given entry block to the database
func AddEntryBlock(db *database.DB, batch *database.Batch, dbheight int64, keyMR string) (err error) {

	// Pull the Entry Block from the Directory Block
	var eBlock *factom.EBlock
	if eBlock, err = factom.GetEBlock(keyMR); err != nil {
		return fmt.Errorf("could not pull the entry block at %s", keyMR)
	}

	// Add all the entries added to Factom for this chain in this block
	for _, v := range eBlock.EntryList {
		// Add the hash to the database
		if err := AddHash(db, batch, v.EntryHash, dbheight); err != nil {
			// Fail on an error
			return fmt.Errorf("database failure on writing entry %s to the database", v.EntryHash)
		}
	}
	// We add a key/value chainID/entryhash for the first entry in every chain.
	// The first entry block of every chain has a BlockSequenceNumber of 0, so we check for that
	if eBlock.Header.BlockSequenceNumber == 0 {
		// We need the binary of the chainID and the entryHash, which we have in the entry block.
		// Note that these MUST be valid hex or all sorts of things fail, so no checking here.
		entryhash, _ := hex.DecodeString(eBlock.EntryList[0].EntryHash)
		chainID, _ := hex.DecodeString(eBlock.Header.ChainID)
		// No realistic failure is possible with putting the key value pair into the batch list
		_ = db.PutBatch(batch, database.ChainBucket, chainID, entryhash)
	}

	return nil
}
