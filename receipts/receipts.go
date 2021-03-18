package receipts

import (
	"fmt"

	"github.com/AccumulateNetwork/SMT/smt"

	"github.com/FactomProject/AnchorPlatform/database"
)

type ApplyHash struct {
	Right bool
	Hash  [32]byte
}

type Receipt struct {
	Right          bool         // True if the hash is coming at our anchor from the right
	Height         int          // Height of the target hash in the merkle state (working variable, not part of proof)
	Object         [32]byte     // Hash for which we want a proof.
	ObjectDbheight int64        // Directory Block Height of the Object
	AnchorDbheight int64        // Directory Block Height of the Anchor
	ApplyHashes    []*ApplyHash // Apply these hashes to create an anchor
	Anchor         [32]byte     // The anchor result of applying hashes documented in ApplyHashes to the Object
}

// Init
// Initialize the state of a Receipt to allow easy addition of hashes as we build up
// the hash path through the Merkle Tree
func (r *Receipt) Init(object [32]byte, objectDbheight, anchorDBHeight int64) {
	r.Object = object
	r.ObjectDbheight = objectDbheight
	r.AnchorDbheight = anchorDBHeight
}

// AddHash
// Add the given hash and update the target hash.  We update the ApplyHashes list,
// the receipt height, and the Anchor values.
func (r *Receipt) AddHash(MS *smt.MerkleState, hash [32]byte, height int) {

}

// GetDbheight
// Get the dbheight of a given object.  If the object is not found in the database, return -1
func GetDbheight(db *database.DB, object []byte) int64 {
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

// GetMerkleState
// Get the MerkleState at the given dbheight
func GetMerkleState(db *database.DB, dbheight int64) (MS *smt.MerkleState) {
	if data := db.Get(database.MerkleStateBucket, database.Int64Bytes(dbheight)); data == nil {
		return nil
	} else {
		MS = new(smt.MerkleState)
		MS.InitSha256()
		MS.UnMarshal(data)
		return MS
	}
}

// GetReceipt
// Return the receipt for the given object using the anchor at the anchorHeight
func GetReceipt(object [32]byte, anchorHeight int64) (receipt *Receipt, err error) {
	db := database.GetDB()

	// Look up the dbheight of the object.  If the object isn't in the database, GetDbheight will return -1
	dbheight := GetDbheight(db, object[:])
	if dbheight < 0 {
		return nil, fmt.Errorf("the object %x is not in the database", object)
	}

	// We have to look at dbheight all the way to anchorHeight.
	searchHeight := dbheight // searchHeight tracks which dbheight we are looking at

	// Get the Merkle State of the directory block.  We are going to use this state to move forward through
	// the blockchain until we have reached the anchor point
	SearchMerkleState := GetMerkleState(db, dbheight)     // This state must exist because the object exists
	AnchorMerkleState := GetMerkleState(db, anchorHeight) // This height might not exist
	if AnchorMerkleState == nil {
		return nil, fmt.Errorf("the anchor height %d is not in the database", anchorHeight)
	}
	// Keep an index into the SearchMerkleState HashList of where we are as we search for object, and later
	// as we build the receipt
	SearchIndex := 0
	CurrentMerkleState := GetMerkleState(db, dbheight-1)
	if CurrentMerkleState == nil {
		CurrentMerkleState = new(smt.MerkleState)
		CurrentMerkleState.InitSha256()
	}

	// look through the hashes for the Merkle State JUST BEFORE the object is added to the Merkle Tree
	for i, v := range SearchMerkleState.HashList {
		// If I find the object, return before the object is added, and the index of the next
		// value in the HashList
		if object == v {
			// Index of Object is i, so let's start there, but don't add object yet.  The proof starts here.
			SearchIndex = i
			break
		}
		// Add the hash to march the CurrentMerkleState forward. v is prior to object.
		CurrentMerkleState.AddToMerkleTree(v)
	}

	// CurrentMerkleState -- state just prior to object being added to the Merkle Tree
	// AnchorMerkleState  -- The state that is the anchor state we need to reach with the receipt
	// SearchMerkleState  -- state holding the list of hashes to add as we build a receipt
	// SearchIndex        -- Index in the SearchMerkleState HashList of the object

	receipt = new(Receipt)
	// Initialize the receipt state with the object and the heights
	receipt.Init(object, dbheight, anchorHeight)
	// Are going to be adding elements we search to the CurrentMerkleState, so we need to
	// Pad the pending list (so we don't have to check for end of Pending when we add an element)
	CurrentMerkleState.PadPending()

	// Now add elements (starting with object) to the CurrentMerkleState until we reach our AnchorHeight
	// First add all the elements of the partial dbheight
	Right := false // The object will start out combining from the left (remember, object is at HashList[SearchIndex]
	Height := 0
	for { // process Merkle states until the MerkleState at AnchorDbheight is reached.
		for _, v1 := range SearchMerkleState.HashList[SearchIndex:] {
			CurrentMerkleState.PadPending() // Pending always has to end in a nil to ensure we handle the "carry" case
			for j, v2 := range CurrentMerkleState.Pending {
				if v2 == nil {
					// If we find a nil spot, then add v1 into Pending
					hash := smt.Hash(v1)
					CurrentMerkleState.Pending[j] = &hash
					break
				}
				// If this is the creation of a higher derivative of object, put it in our path
				if j == Height {
					applyHash := new(ApplyHash)
					receipt.ApplyHashes = append(receipt.ApplyHashes, applyHash)
					applyHash.Right = Right
					if Right {
						applyHash.Hash = v1
					} else {
						applyHash.Hash = *v2
					}
					// v1 becomes HashOf(pending[j]+v1)
					Height++      // The Anchor will now move up one level in the Merkle Tree
					Right = false // The Anchor hashes on the left are now added to the Anchor
				}
				// v1 becomes HashOf(pending[j]+v1)
				v1 = CurrentMerkleState.HashFunction(append(v2[:], v1[:]...))
				CurrentMerkleState.Pending[j] = nil
			}
			Right = true
		}
		// If we have reached the anchorHeight, we are done!
		if searchHeight == anchorHeight {
			break
		}
		// Get the next MerkleState
		searchHeight++
		SearchMerkleState = GetMerkleState(db, searchHeight)
		SearchIndex = 0

	}
	var anchor *smt.Hash
	for i, v := range AnchorMerkleState.Pending {
		if v == nil { // if v is nil, there is nothing to do regardless. Note i cannot == Height
			continue // because the previous code tracks Height such that there is always a value.
		}
		if anchor == nil { // Find the first non nil in pending
			anchor = new(smt.Hash) // Make an anchor
			*anchor = *v           // copy over the value of v (which is never nil
			if i == Height {       // Note that there is no way this code does not
				Right = false //      execute before the following code that adds
			} else { //               applyHash records to the receipt
				Right = true
			}
			continue
		}
		if i >= Height { // At this point, we have a hash in the anchor, and we are
			applyHash := new(ApplyHash) // combining hashes as we go.
			receipt.ApplyHashes = append(receipt.ApplyHashes, applyHash)
			applyHash.Right = Right // We record if the proof path is in the anchor
			if Right {              // or in the pending array
				applyHash.Hash = *anchor
			} else {
				applyHash.Hash = *v
			}
			// Note once the object derivative is in anchor, it never leaves. All combining
			Right = false // then comes from the left
			// Combine all the pending hash values into the anchor
			hash := AnchorMerkleState.HashFunction(append((*v)[:], (*anchor)[:]...))
			anchor = &hash
		}
	}
	// We are testing anchor, but it can't be nil
	if anchor == nil {
		panic("anchor was nil, and this should not be possible.")
	}
	// Put the resulting anchor into the receipt
	receipt.Anchor = *anchor

	return receipt, nil
}
