package factomSync

import (
	"fmt"
	"os"
	"time"

	"github.com/FactomProject/factom"
)

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
func (s *Sync) GetDatabaseHeight() (databaseHeight int64) {

	databaseHeightBytes := s.Manager.DBManager.Get("Factom",[]byte("Head"))
	if len(databaseHeightBytes) != 8 { // If we don't find an int64 value, then return zero
		return -1
	}
	databaseHeight, _ = BytesInt64(databaseHeightBytes)
	return databaseHeight // Return the height
}



// SetDatabaseHeight
// Set the Database height in the database, so we can pick up syncing where we left off.
// We don't worry about being exact, because syncing will just overwrite what is already there
// with the same values.
func (s *Sync) SetDatabaseHeight(dbheight int64) {
	_ = s.Manager.DBManager.PutBatch("Factom",[]byte("Head"),Int64Bytes(dbheight))
}
