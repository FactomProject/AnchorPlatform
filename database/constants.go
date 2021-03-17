// For key value stores where buckets are not supported, we add a byte to the
// key to represent a bucket. For now, all buckets are hard coded, but we could
// change that in the future.
//
// Buckets are not really enough to index everything we wish to index.  So
// we have labels as well.  Labels are shifted 8 bits left, so they can be
// combined with the buckets to create a unique key.
//
// This allows us to put the raw directory block at DBlockBucket+L_raw, and meta data
// about the directory block at DBlockBucket+MetaLabel
package database

// Map of buckets to the bucket byte
var bucket map[string][]byte

// General constants
const (
	// Marks are the states just before we hit a minimum power of 2.  This allows
	// us to skip all the intervening states when creating a Merkle proof for a receipt for a hash.
	MarkPower = 10
	Mark      = int64(1) << MarkPower // Merkle State Mark at every 1024 (hex 400 or 1<<10) elements
	MarkMask  = Mark - 1              // Mask to Mark, 1023
)

// buckets
const (
	DBlockBucket      = int64(iota + 1) // Index information about directory blocks
	ChainBucket                         // Index of ChainIDs to the first entry in that chain
	ObjectBucket                        // Index information about general objects in Factom
	MerkleStateBucket                   // Index information about Merkle States
	BitcoinBucket                       //
	EthereumBucket
	TestBucket
)

// labels
// Labels allow buckets to index different sorts of values.  For example, indexing dbheight vs element counts
// For example MerkleStateBucket can track indexes by DBHeight, and MerkleStateBucket+MerkleStateMarks can track
// indexes by element count.
const (
	MetaLabel        = int64(iota+1) << 16 // buckets in the low 16 bits, Labels in higher bits
	MerkleStateMarks                       // Merkle State at every 1024 element mark
	MerkleStateNext                        // Next hash added to a MerkleStateMark
)
