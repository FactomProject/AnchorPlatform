package factomSync

import (
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/factom"
)

// AddDirectoryBlock()
// Add the Directory Block hash to the database and to the Merkle tree.  We also add all the other blocks
// to the database that the Directory Block references, i.e. the Admin Block, the Entry Credit Block, the
// Factoid Block, and the various Entry Blocks for chains modified in this directory block
func (s *Sync) AddDirectoryBlock(dbheight int64) (err error) {

	var dBlock *factom.DBlock
	if dBlock, err = factom.GetDBlockByHeight(dbheight); err != nil {
		return fmt.Errorf("failed to read the dBlock at height %d\n", dbheight)
	}

	// Add the hash of the directory block to the MerkleTree
	s.Manager.AddHashString(dBlock.KeyMR)

	// We put in all the directory block entries.  This includes the hash for the admin block, entry credit block,
	// and factoid block.
	for i, v := range dBlock.DBEntries {
		// Add all the hashes for the directory block entries, which covers the admin, entry credit, factoid,
		// and entry blocks.
		s.Manager.AddHashString(v.KeyMR)
		// Entry blocks start at index 3; Add all the entries in the entry blocks as well.
		if i >= 3 {
			if err = s.AddEntryBlock(v.KeyMR); err != nil {
				return err
			}
		}
	}

	// Now add all the transactions in the Factoid block to the database
	if err = s.AddFactoidBlock(dbheight, dBlock); err != nil {
		return fmt.Errorf("could not add the Factoid Block : %v ", err)
	}

	return nil

}


// AddEntryCreditBlock()
// Unlike the old anchoring support, the Anchor Platform will allow receipts for commits made to the Factom protocol
func (s *Sync) AddFactoidBlock(dbheight int64, dBlock *factom.DBlock) (err error) {

	// Pull the Factoid Block from the Directory Block
	var FBlock *factom.FBlock
	if FBlock, err = factom.GetFBlock(dBlock.DBEntries[2].KeyMR); err != nil {
		return fmt.Errorf("retrieve of factoid block %s failed at height %d", dBlock.DBEntries[2].KeyMR, dbheight)
	}

	// Add all the factoid transactions to the database and MerkleTree
	for _, v := range FBlock.Transactions {
		s.Manager.AddHashString(v.TxID)
	}

	return nil
}

// AddEntryBlock
// Add the entries in the given entry block to the database
func (s *Sync) AddEntryBlock(keyMR string) (err error) {

	// Pull the Entry Block from the Directory Block
	var eBlock *factom.EBlock
	if eBlock, err = factom.GetEBlock(keyMR); err != nil {
		return fmt.Errorf("could not pull the entry block at %s", keyMR)
	}

	// Add all the entries added to Factom for this chain in this block
	for _, v := range eBlock.EntryList {
		// Add the hash to the database
		s.Manager.AddHashString(v.EntryHash)
	}
	// We add a key/value chainID/entryhash for the first entry in every chain.
	// The first entry block of every chain has a BlockSequenceNumber of 0, so we check for that
	if eBlock.Header.BlockSequenceNumber == 0 {
		// We need the binary of the chainID and the entryHash, which we have in the entry block.
		// Note that these MUST be valid hex or all sorts of things fail, so no checking here.
		entryhash, _ := hex.DecodeString(eBlock.EntryList[0].EntryHash)
		chainID, _ := hex.DecodeString(eBlock.Header.ChainID)
		// No realistic failure is possible with putting the key value pair into the batch list
		_ = s.Manager.DBManager.PutBatch("Chains",chainID,entryhash)
	}

	return nil
}
