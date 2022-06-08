package factomSync

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestCacheEntry(t *testing.T){
	const numEntries =100
	var rh RandHash
	ce := new(CacheEntry)
	for i:= 0; i < numEntries; i++{
		ce.ChainID = rh.NextA()
		e := rh.GetRandBuff(int(rh.GetRandUint64())%1000)
		ce.Entries = append(ce.Entries,e)
		ce.EntryHashes = append(ce.EntryHashes,sha256.Sum256(e))
	}

	data := ce.Marshal()
	ce2 := new(CacheEntry)
	ce2.UnMarshal(data)
	fmt.Print(len(data))
	switch {
	case ce.ChainID != ce2.ChainID:
		t.Fail()
	case ce.Index != ce2.Index:
		t.Fail()	
	case len(ce.EntryHashes)!= len(ce2.EntryHashes):
		t.Fail()
	case len(ce.Entries)!= len(ce2.Entries):
		t.Fail()
	case len(ce.Entries)!= len(ce.EntryHashes):
		t.Fail()
	}

	for i, e := range ce.EntryHashes{
		if e != ce.EntryHashes[i]{
			t.Fail()
		}
	}
}