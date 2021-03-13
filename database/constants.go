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

// buckets
const (
	DBlockBucket      = iota + 1 // Index information about directory blocks
	ObjectBucket                 // Index information about general objects in Factom
	MerkleStateBucket            // Index information about merklestates
	BitcoinBucket                //
	EthereumBucket
	TestBucket
)

// labels
const (
	MetaLabel = (iota + 1) << 16
	RawLabel
)
