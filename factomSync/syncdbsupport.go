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

		return -1
	
}

// SetDatabaseHeight
// Set the Database height in the database, so we can pick up syncing where we left off.
// We don't worry about being exact, because syncing will just overwrite what is already there
// with the same values.
func (s *Sync) SetDatabaseHeight(dbheight int64) {

}
