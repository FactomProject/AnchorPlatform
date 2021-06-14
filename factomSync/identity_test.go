package factomSync

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factom"
)

func TestIdentity(t *testing.T) {
	fsync := new(Sync)
	fsync.Init(nil)

	dbheight := getFactomdHeight()
	var err error
	var dBlock *factom.DBlock
	var prev []*ANO
	for i := int64(0); i < dbheight; i++ {
		dBlock, err = factom.GetDBlockByHeight(i)
		if err != nil {
			t.Fatal("should be able to get a dBlock")
		}
		if dBlock == nil {
			t.Fatal("should not have a nil dbloc")
		}
		if updated := fsync.ProcessIdentity(dBlock); updated {
			fmt.Println("----------", i)
			al := fsync.GetANOList()
			if prev != nil {
				if len(prev) == len(al) {

				}
			}
			for i, v := range al {
				fmt.Printf("%3d %s\n", i, v.ChainID)
			}
		}
		dbheight = getFactomdHeight()
	}
	if updated := fsync.ProcessIdentity(dBlock); updated {
		fmt.Println("----------", dbheight)
		al := fsync.GetANOList()
		for i, v := range al {
			fmt.Printf("%3d %s\n", i, v.ChainID)
		}
	}
}
