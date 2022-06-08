package factomSync

// CacheEntry
// An entry in the cache for the Sync structure.
// We cache all the entry hashes for every user chain in Factom.  The entries
// are put into the database by entry hash. Each entry hash belongs to chains.
// Rather than making us do a database read for every entry in a chain, the entry
// hashes are cached, and written out in batches (a CacheEntry).
type CacheEntry struct {
	ChainID     [32]byte
	Index       int64
	EntryHashes [][32]byte
	Entries     [][]byte
}

// Marshal the ChainEntry to bytes. The ChainID and Index isn't needed later, but
// for now marshal them anyway
func (c *CacheEntry) Marshal() (data []byte) {
	data = append(data, c.ChainID[:]...)                          // Append the ChainID
	data = append(data, Int64Bytes(c.Index)...)                   // Append the Index
	data = append(data, Int64Bytes(int64(len(c.EntryHashes)))...) // Count of entries
	for _, eh := range c.EntryHashes {                            // Write out all the Entry Hashes
		data = append(data, eh[:]...)
	}
	for _, e := range c.Entries { //                                 Write out all the entries
		data = append(data, Uint16Bytes(uint16(len(e)))...) //       Each entry has a length
		data = append(data, e...)                           //       followed by its data
	}
	return data
}

// UnMarshal the ChainEntry from bytes.  When loading Accumulate, each chain's list
// of entry hashes is needed.  The entries are pulled by hash.  We can consider
//
func (c *CacheEntry) UnMarshal(data []byte) {
	copy(c.ChainID[:], data[:32])
	data = data[32:]
	c.Index, data = BytesInt64(data)
	count, data := BytesInt64(data) // Count of EntryHashes (and of Entries; must be the same)

	for i:= int64(0);i<count;i++ { // Load EntryHashes
		eh := [32]byte{}
		copy(eh[:], data[:32])
		data = data[32:]
		c.EntryHashes = append(c.EntryHashes, eh)
	}

	for i:= int64(0);i<count;i++{ // Load Entries; Each is an entry length followed by entry data
		len,data := BytesUint16(data)
		c.Entries = append(c.Entries, append([]byte{},data[:len]...))
		data = data[len:]
	}
}
