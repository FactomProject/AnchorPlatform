package factomSync

import (
	"fmt"
	"os"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/FactomProject/AnchorPlatform/config"
	"github.com/FactomProject/AnchorPlatform/database"
	"github.com/FactomProject/factom"
)

// Sync
// The Sync process continues as long as the Anchor Platform is running.  When the Anchor Platform is
// behind factomd, then blocks are requested and added to our database.  Furthermore, we grab all the
// anchors written to external chains as they are posted to the Anchor Chain in factom.
func Sync(conf *config.Config) {

	// Check for custom factomd configuration
	if conf.Factom.Server != "" {
		fmt.Printf("Factomd server: %s\n", conf.Factom.Server)
		factom.SetFactomdServer(conf.Factom.Server)
	} else {
		fmt.Printf("Factomd server: localhost\n")
	}
	if conf.Factom.User != "" && conf.Factom.Password != "" {
		factom.SetFactomdRpcConfig(conf.Factom.User, conf.Factom.Password)
	}

	// Set our hashing function in the Merkle State
	MS.InitSha256()
	// Get our database
	db := database.GetDB(conf.DBName)

	defer func() { _ = db.Close() }()

	// Get the database height from the database
	databaseHeight := GetDatabaseHeight(db)
	// Get the initial factomdHeight
	factomdHeight := getFactomdHeight()

	// On a control-c, we close the database and wait a bit before exiting the program.
	database.AddInterruptHandler(func() {
		Stop = true
		fmt.Println("Exiting Anchor Platform")
		_ = db.Close()
		time.Sleep(20 * time.Second)
		os.Exit(0)
	})

	running := time.Now()

	// Update the user where we are right from the start, since the next feedback won't
	// be until there is a new block, if we are already synced with factomd.
	fmt.Printf("Starting the Anchor Platform at Height: %s  with %s Objects\n",
		humanize.Comma(databaseHeight),
		humanize.Comma(MS.GetCount()))

	for {
		// Update the user where we are
		fmt.Printf("Height: %s  Objects: %s  Running time: %s\n",
			humanize.Comma(databaseHeight),
			humanize.Comma(MS.GetCount()),
			database.FormatTimeLapse(time.Now().Sub(running)))
		for {
			// Make sure we do not busy wait
			time.Sleep(10 * time.Second)
			// If we have been signaled to stop, then return out of the syncing process
			// This will close the database
			if Stop {
				return
			}
			// Get the directory block height in factomd and the database and see if we have something to do
			databaseHeight = GetDatabaseHeight(db)
			factomdHeight = getFactomdHeight()
			// If we are caught up, continue (wait 10 seconds before we check again)
			if factomdHeight == databaseHeight {
				continue
			}
			break
		}
		// To estimate how long this is taking and how long it will take, we need the time we started,
		// and the databaseHeight we started at.
		begin := time.Now()
		startHeight := databaseHeight

		// Writing to the database is really slow if each key/value pair is written by itself, so we
		// batch the writes.
		batch := db.BeginBatch()

		for dbheight := databaseHeight + 1; !Stop && dbheight <= factomdHeight; dbheight++ {

			// Process the directory block at the height dbheight
			if err := AddDirectoryBlock(db, batch, dbheight); err != nil {
				panic(err)
			}

			// Every 10k key pairs or so, write all the key values to the database.  We have to do this on
			// Directory Block boundaries or we risk writing a partial block to the database.
			//
			// Note that batch can be reused.
			//
			// Toward the end (less than 10 blocks to go) then end the batch after every directory block processed.
			if len(*batch) > 1000 || factomdHeight-dbheight < 10 {
				SetDatabaseHeight(batch, db, dbheight)
				db.EndBatch(batch)
			}

			// Every so often we are going to give feedback to the user about syncing.  But only if
			// we are syncing more than 100 blocks.  Otherwise, we just do our job and let the UI provide
			// feedback.  Always print the first time.
			if dbheight != 0 && dbheight%5000 == 0 && factomdHeight-dbheight >= 10 {
				// Figure out how long we have been syncing
				timeSpent := time.Now().Sub(begin)
				// Figure out how much of the missing blocks we have processed
				percentDone := float64(dbheight-startHeight) / float64(factomdHeight)
				// Figure out an estimate of how much longer this is going to take.
				timePerBlock := float64(timeSpent.Seconds()) / float64(dbheight-startHeight)
				timeLeft := int64(float64(factomdHeight-dbheight) * timePerBlock)
				// fmt.Printf("Time per block %6.4f and blocks left %d of %d\n", timePerBlock, factomdHeight-dbheight, factomdHeight)
				// Print our feedback
				fmt.Printf("%s Height: %10s  Objects: %13s  Done: %02.0f%%  Left: %s  ~total: %s\n",
					database.FormatTimeLapse(timeSpent),
					humanize.Comma(dbheight),
					humanize.Comma(MS.GetCount()),
					percentDone*100,
					database.FormatTimeLapseSeconds(int64(timeLeft)),
					database.FormatTimeLapseSeconds(int64(timeLeft)+int64(timeSpent.Seconds())),
				)
			}
		}
	}
}
