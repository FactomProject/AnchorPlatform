package factomSync

import (
	"fmt"
	"os"
	"time"

	"github.com/AccumulateNetwork/SMT/storage/database"
	"github.com/FactomProject/AnchorPlatform/config"
	"github.com/FactomProject/factom"
	"github.com/dustin/go-humanize"
)

type Sync struct {
	// Don't Persist:
	db            database.Manager      // Our database
	Stop          bool                  // Flag to stop go routines
	Start         time.Time             // Time Syncing started in this pass
	EntryCache    map[string]CacheEntry // Chain cache with list of entries
	CurrentHeight int64                 // Current Height in the Cache

	// Persist:
	DatabaseHeight int64
	FactomdHeight  int64
	EntryCount     int64
	ChainCount     int64

	// Note:
	// We can cache all the data we want in memory, then write it to the database every so often.
	// As long as we only update the Sync to the database after our data is written, we don't have
	// to worry about updates to the database not reflected by the Sync.DatabaseHeight.  That's
	// because if the application is halted with some data written, restarting will just write that
	// same data again.  But no data will be lost.
}

func (s *Sync) DumpCache() {
	if s.EntryCache == nil {
		return
	}
	for _, CEntry := range s.EntryCache {
		s.db.DB.Put(GetKey("Chain", CEntry.Index), CEntry.Marshal())
	}

}

func (s *Sync) Marshal() (data []byte) {
	data = append(data, Int64Bytes(s.DatabaseHeight)...)
	data = append(data, Int64Bytes(s.FactomdHeight)...)
	data = append(data, Int64Bytes(s.EntryCount)...)
	data = append(data, Int64Bytes(s.ChainCount)...)
	return data
}

func (s *Sync) UnMarshal(data []byte) {
	s.DatabaseHeight = int64(data[0])<<24 + int64(data[1])<<16 + int64(data[2])<<8 + int64(data[3])
	s.FactomdHeight = int64(data[0])<<24 + int64(data[1])<<16 + int64(data[2])<<8 + int64(data[3])
	s.EntryCount = int64(data[0])<<24 + int64(data[1])<<16 + int64(data[2])<<8 + int64(data[3])
	s.ChainCount = int64(data[0])<<24 + int64(data[1])<<16 + int64(data[2])<<8 + int64(data[3])
	if s.EntryCache == nil {
		s.EntryCache = make(map[string]CacheEntry)
	}
}

// Initialize
// Initialize the database, the Merkle Tree, etc.

func (s *Sync) Init(conf *config.Config) {
	// Check for custom factomd configuration
	if conf != nil && conf.Factom.Server != "" {
		fmt.Printf("Factomd server: %s\n", conf.Factom.Server)
		factom.SetFactomdServer(conf.Factom.Server)
	} else {
		fmt.Printf("Factomd server: localhost\n")
	}

	if conf != nil && conf.Factom.User != "" && conf.Factom.Password != "" {
		factom.SetFactomdRpcConfig(conf.Factom.User, conf.Factom.Password)
	}

	// Initialize the database
	db := new(database.Manager)
	if conf != nil {
		if err := db.Init("badger", conf.DBName); err != nil {
			panic(fmt.Sprint("error configuring the database: ", err))
		}
	} else {
		if err := db.Init("memory", ""); err != nil {
			panic("memory databases should always work")
		}
	}

	db.AddBucket("Factom") // Give ourselves a bucket to keep our stuff in
	db.AddBucket("Chains") // Add a bucket to collect the first entries of chains

	// Get the database height from the database
	s.DatabaseHeight = s.GetDatabaseHeight()
	// Get the initial s.FactomdHeight
	s.FactomdHeight = getFactomdHeight()

	// On a control-c, we close the database and wait a bit before exiting the program.
	AddInterruptHandler(func() {
		s.Stop = true                          // Signal to the Run go routine that we are done
		fmt.Println("Exiting Anchor Platform") // A bit of feedback
		db.Close()                             // This is where we close the database.
		time.Sleep(20 * time.Second)           // Give some time for go routines to shut down
		os.Exit(0)                             // We are done
	})

	s.Start = time.Now()

	// Update the user where we are right from the start, since the next feedback won't
	// be until there is a new block, if we are already synced with factomd.
//	fmt.Printf("Starting the Anchor Platform at Height: %s  with %s Objects\n",
//		humanize.Comma(s.DatabaseHeight),
//		humanize.Comma(s.Manager.MS.Count))

}


// WaitForBlock
// Wait for factomd to produce a new block.  When a new block is found, then return.  WaitForBlock
// sleeps between each poll.
func (s *Sync) WaitForBlock() {
	for !s.Stop { // While stop isn't set... If Stop is set, we have to bug out.
		// Make sure we do not busy wait
		time.Sleep(10 * time.Second)
		// Get the directory block height in factomd and the database and see if we have something to do
		s.DatabaseHeight = s.GetDatabaseHeight()
		s.FactomdHeight = getFactomdHeight()
		// If we are caught up, continue (wait 10 seconds before we check again)
		if s.FactomdHeight > s.DatabaseHeight {
			return
		}
	}

}

// Run
// The Sync process continues as long as the Anchor Platform is running.  When the Anchor Platform is
// behind factomd, then blocks are requested and added to our database.  Furthermore, we grab all the
// anchors written to external chains as they are posted to the Anchor Chain in factom.
func (s *Sync) Run(conf *config.Config) {

	s.Init(conf)

	for !s.Stop { // Continue processing as long as Stop isn't set (set by an interrupt of the program)
		// Update the user where we are
		fmt.Printf("Height: %s  Entries: %s Chains: %s Running time: %s\n",
			humanize.Comma(s.DatabaseHeight),
			humanize.Comma(s.EntryCount),
			humanize.Comma(s.ChainCount),
			FormatTimeLapse(time.Now().Sub(s.Start)))
		// To estimate how long this is taking and how long it will take, we need the time we started,
		// and the s.DatabaseHeight we started at.

		s.WaitForBlock() // Wait for blocks in Factomd to show up to be processed

		begin := time.Now()
		startHeight := s.DatabaseHeight

		for dbheight := s.DatabaseHeight + 1; !s.Stop && dbheight <= s.FactomdHeight; dbheight++ {

			// Process the directory block at the height dbheight
			if err := s.AddDirectoryBlock(dbheight); err != nil {
				panic(err)
			}

			// Every 10k key pairs or so, write all the key values to the database.  We have to do this on
			// Directory Block boundaries or we risk writing a partial block to the database.
			//
			// Note that batch can be reused.
			//
			// Toward the end (less than 10 blocks to go) then end the batch after every directory block processed.
			if s.FactomdHeight-dbheight < 10 {
				s.SetDatabaseHeight(dbheight)
			}

			// Every so often we are going to give feedback to the user about syncing.  But only if
			// we are syncing more than 100 blocks.  Otherwise, we just do our job and let the UI provide
			// feedback.  Always print the first time.
			if dbheight != 0 && dbheight%50 == 0 && s.FactomdHeight-dbheight >= 10 {
				// Figure out how long we have been syncing
				timeSpent := time.Now().Sub(begin)
				// Figure out how much of the missing blocks we have processed
				percentDone := float64(dbheight-startHeight) / float64(s.FactomdHeight)
				// Figure out an estimate of how much longer this is going to take.
				timePerBlock := timeSpent.Seconds() / float64(dbheight-startHeight)
				timeLeft := int64(float64(s.FactomdHeight-dbheight) * timePerBlock)
				// fmt.Printf("Time per block %6.4f and blocks left %d of %d\n", timePerBlock, s.FactomdHeight-dbheight, s.FactomdHeight)
				// Print our feedback
				fmt.Printf("%s Height: %10s  Entries: %13s Chains %13s Done: %04.0f%%  Left: %s  ~total: %s\n",
					FormatTimeLapse(timeSpent),
					humanize.Comma(dbheight),
					humanize.Comma(s.EntryCount),
					humanize.Comma(s.ChainCount),
					percentDone*100,
					FormatTimeLapseSeconds(timeLeft),
					FormatTimeLapseSeconds(timeLeft+int64(timeSpent.Seconds())),
				)
			}
		}
	}
}
