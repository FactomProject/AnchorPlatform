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

	// We put in all the directory block entries.  This includes the hash for the admin block, entry credit block,
	// and factoid block.
	for i, v := range dBlock.DBEntries {
		// Go through all the Entry blocks
		if i >= 3 {
			if err = s.AddEntryBlock(v.KeyMR); err != nil {
				return err
			}
		}
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
		if eHash, err := hex.DecodeString(v.EntryHash); err != nil {
			panic("Bad Entry Hash")
		} else {
			var eHashA [32]byte
			copy(eHashA[:], eHash)
			if entry := s.db.DB.Get(eHashA); entry == nil {
				factom.GetEntry(v.EntryHash)
				s.db.DB.Put(eHashA, eHash)
				if eBlock.Header.BlockSequenceNumber == 0 {
				//	s.EntryCache[eBlock.Header.ChainID]
				}
			}
		}
	}
	

	return nil
}
