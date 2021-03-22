package receipts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"

	"github.com/FactomProject/AnchorPlatform/factomSync"

	"github.com/AccumulateNetwork/SMT/smt"

	"github.com/FactomProject/AnchorPlatform/database"
)

type ApplyHash struct {
	Right bool
	Hash  [32]byte
}

// Receipt
// Struct builds the Merkle Tree path component of a Merkle Tree Proof.
// TODO: Code around Receipt handling and generation should be moved to the SMT library
type Receipt struct {
	Object         [32]byte     // Hash for which we want a proof.
	ObjectDbheight int64        // Directory Block Height of the Object
	AnchorDbheight int64        // Directory Block Height of the Anchor
	ApplyHashes    []*ApplyHash // Apply these hashes to create an anchor
	Anchor         [32]byte     // The anchor result of applying hashes documented in ApplyHashes to the Object
}

// String
// Convert the receipt to a string
func (r *Receipt) String() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("\nObject %x\n", r.Object))
	b.WriteString(fmt.Sprintf("Anchor %x\n", r.Anchor))
	b.WriteString(fmt.Sprintf("DBHeight %d\n", r.ObjectDbheight))
	b.WriteString(fmt.Sprintf("AnchorHeight %d\n", r.AnchorDbheight))
	working := r.Object
	for i, v := range r.ApplyHashes {
		r := "L"
		if v.Right {
			r = "R"
			working = sha256.Sum256(append(working[:], v.Hash[:]...))
		} else {
			working = sha256.Sum256(append(v.Hash[:], working[:]...))
		}
		b.WriteString(fmt.Sprintf(" %10d Apply %s %x working: %x \n", i, r, v.Hash, working))
	}
	return b.String()
}

// Init
// Initialize the state of a Receipt to allow easy addition of hashes as we build up
// the hash path through the Merkle Tree
func (r *Receipt) Init(object [32]byte, objectDbheight, anchorDBHeight int64) {
	r.Object = object
	r.ObjectDbheight = objectDbheight
	r.AnchorDbheight = anchorDBHeight
}

// GetMerkleState
// Get the MerkleState at the given dbheight
func GetMerkleState(db *database.DB, dbheight int64) (MS *smt.MerkleState) {

	if data := db.Get(database.DBlockBucket, database.Int64Bytes(dbheight)); data == nil {
		return nil
	} else {
		MS = new(smt.MerkleState)
		MS.InitSha256()
		MS.UnMarshal(data)
		return MS
	}
}

// GetMarkState
// Get the MerkleState at the given the given mark.  DBHeight must have at least the low bits
// clear
func GetMarkState(db *database.DB, count int64) (MS *smt.MerkleState, markNext [32]byte) {
	var data []byte
	countBytes := database.Int64Bytes(count)
	if data = db.Get(database.MarkBucket+database.MerkleState, countBytes); data == nil {
		return nil, markNext
	}
	MS = new(smt.MerkleState)
	MS.InitSha256()
	MS.UnMarshal(data)

	if data = db.Get(database.MarkBucket+database.MarkNext, countBytes); data == nil {
		return nil, markNext
	}

	copy(markNext[:], data[:32])

	return MS, markNext
}

// CenterOnObject
// Returns a state
func CenterOnObject(
	db *database.DB,
	object [32]byte,
	dbheight int64) (SearchMerkleState, CurrentMerkleState *smt.MerkleState, SearchIndex int64) {

	// SearchMerkleState is the state at the end of a directory block
	SearchMerkleState = GetMerkleState(db, dbheight)
	// CurrentMerkleState needs to be the MerkleState is the state just before a directory block
	CurrentMerkleState = GetMerkleState(db, dbheight-1)
	// If no MerkleState exists prior to dbheight, then just use the empty Merkle State
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
			SearchIndex = int64(i)
			return SearchMerkleState, CurrentMerkleState, SearchIndex
		}
		// Add the hash to march the CurrentMerkleState forward. v is known to be prior to object.
		CurrentMerkleState.AddToMerkleTree(v)
	}
	// The database already asserted that the object was in this directory block.  So we should not
	// fail to find it under any circumstance.  If we do, just panic.  The database must be screwy,
	// and there is no fix.
	panic(fmt.Sprintf("entry %x not found at dbheight %d. The database may be corrupted.", object, dbheight))
}

// GetReceipt
// Return the receipt for the given object using the anchor at the anchorHeight
func GetReceipt(object [32]byte, anchorHeight int64) (receipt *Receipt, err error) {
	db := database.GetDB("db")

	// Look up the dbheight of the object.  If the object isn't in the database, GetDbheight will return -1
	dbheight := factomSync.GetObjectDbheight(db, object[:])
	if dbheight < 0 {
		return nil, fmt.Errorf("the object %x is not in the database", object)
	}

	// Make sure the anchorHeight actually includes the dbheight
	if dbheight > anchorHeight {
		return nil, fmt.Errorf("the anchorHeight %d is less than the object's dbheight %d", anchorHeight, dbheight)
	}

	// Make sure the anchorHeight exists in the database
	databaseHeight := factomSync.GetDatabaseHeight(db)
	if anchorHeight > databaseHeight {
		return nil, fmt.Errorf("the anchor height %d is greater than the database height %d", anchorHeight, databaseHeight)
	}

	// Get the AnchorMerkleState
	AnchorMerkleState := GetMerkleState(db, anchorHeight)

	SearchMerkleState, CurrentMerkleState, SearchIndex := CenterOnObject(db, object, dbheight)

	// CurrentMerkleState -- state just prior to object being added to the Merkle Tree
	// SearchMerkleState  -- state holding the list of hashes to add as we build a receipt
	// SearchIndex        -- Index in the SearchMerkleState HashList of the object
	// AnchorMerkleState  -- The Merkle State after adding the Anchor Directory Block
	receipt = new(Receipt)
	// Initialize the receipt state with the object and the heights
	receipt.Init(object, dbheight, anchorHeight)

	// Now add elements (starting with object) to the CurrentMerkleState until we reach our AnchorHeight
	// First add all the elements of the partial dbheight
	Right := false // The object will start out combining from the left (remember, object is at HashList[SearchIndex]
	Height := 0    // This is the height in Pending in the currentMerkleState that is the derivative of Object

	var AddAHash = func(v1 smt.Hash) {
		original := v1
		CurrentMerkleState.PadPending() // Pending always has to end in a nil to ensure we handle the "carry" case
		for j, v2 := range CurrentMerkleState.Pending {
			if v2 == nil {
				// If we find a nil spot, then we found where the hash will go.  To keep
				// the accounting square, we won't add it ourselves, but will let the Merkle Tree
				// library do the deed.  That keeps the Merkle Tree State square.
				CurrentMerkleState.AddToMerkleTree(original)
				if j == Height { // If we combine with our proof height, the NEXT combination will be
					Right = true // from the right.
				}
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
				//fmt.Printf("Height %3d %x\n", Height, applyHash.Hash)
				// v1 becomes HashOf(pending[j]+v1)
				Height++      // The Anchor will now move up one level in the Merkle Tree
				Right = false // The Anchor hashes on the left are now added to the Anchor
			}
			// v1 becomes HashOf(pending[j]+v1)
			v1 = CurrentMerkleState.HashFunction(append(v2[:], v1[:]...))
		}
	}

	// Add the Merkle Dag of the AnchorMerkleState to the receipt.  Note that we do this
	// in a couple places, so we are using a lamda function to make the code cleaner.
	var AddMerkleDag = func() {
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

	}

	// Add hashes to build up  the receipt and move the state to the target Anchor Height
	var ProcessHash = SearchMerkleState.HashList[SearchIndex:] // This is the list of hashes we are adding to the Merkle Tree

	for Height < database.MarkPower { // process Merkle states until we reach a sub Merkle Tree of MarkPower height

		if len(ProcessHash) > 0 { // If we have more, then process more
			AddAHash(ProcessHash[0])      // Add it to the Merkle Tree
			ProcessHash = ProcessHash[1:] // Then remove from list
		}

		// Get the next list if the previous list is done
		if len(ProcessHash) == 0 {
			// If we have reached the anchorHeight, we are done!
			if dbheight == anchorHeight {
				AddMerkleDag()      // Create the Merkle Dag
				return receipt, nil // Return the completed receipt
			}
			dbheight++                                       // point to the next Directory Block
			SearchMerkleState = GetMerkleState(db, dbheight) // Get the Hashes from that directory
			ProcessHash = SearchMerkleState.HashList         //  and put them in the ProcessHash and keep looking.
		}
	}

	// Now we combine sub Merkle Tree roots until we make it to the Anchor
	for {
		// Calculate the count for the next sub Merkle Tree root.  It must
		// be 2^height elements further in the blockchain
		step := int64(math.Pow(2, float64(Height)))

		// Figure out which Mark to load next.
		var markNext [32]byte
		currentCount := CurrentMerkleState.GetCount()
		markCount := currentCount&((step-1)^-1) + step - 1

		// If there is room to go to the next sub merkle tree, then go there.
		if markCount <= AnchorMerkleState.GetCount() {
			// Get the next object to add at the mark (updates our receipt properly
			LastMerkleState := CurrentMerkleState.Copy()
			_ = LastMerkleState
			CurrentMerkleState, markNext = GetMarkState(db, markCount)
			Right = true       // Add the hash from the right
			AddAHash(markNext) // Now we are either ready to continue the search, or build our receipt
		}

		// Have we found the highest Sub Merkle Tree prior to the Anchor,
		// and there is no Sub Merkle Tree that's higher that holds the Anchor
		if markCount+step >= AnchorMerkleState.GetCount() {
			AddMerkleDag()
			return receipt, nil
		}

	}
}

// Receipt
// Take a receipt and validate that the
func (r Receipt) Validate() bool {
	anchor := r.Object // To begin with, we start with the object as the anchor
	// Now apply all the path hashes to the anchor
	for _, applyHash := range r.ApplyHashes {
		// Need a [32]byte to slice
		hash := [32]byte(applyHash.Hash)
		if applyHash.Right {
			// If this hash comes from the right, apply it that way
			anchor = sha256.Sum256(append(anchor[:], hash[:]...))
		} else {
			// If this hash comes from the left, apply it that way
			anchor = sha256.Sum256(append(hash[:], anchor[:]...))
		}
	}
	// In the end, anchor should be the same hash the receipt expects.
	return anchor == r.Anchor
}
