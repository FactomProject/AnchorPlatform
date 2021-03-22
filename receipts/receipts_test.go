package receipts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/FactomProject/AnchorPlatform/factomSync"

	"github.com/FactomProject/AnchorPlatform/database"
)

func TestGetReceipt(t *testing.T) {
	db := database.GetDB()

	DBHead := factomSync.GetDatabaseHeight(db)
	fail := false
	failCount := 0
	start := time.Now()
	for objectHeight := int64(3); objectHeight < DBHead-1; objectHeight += 100 {

		if true || objectHeight%1000 == 0 {
			sofar := time.Now().Sub(start)
			objectsPerSecond := float64(objectHeight/100) / sofar.Seconds()
			fmt.Printf("Processed: %6d Failures: %d Test Time: %s Receipts per second: %8.0f\n",
				objectHeight,
				failCount,
				database.FormatTimeLapse(sofar),
				objectsPerSecond)
			fmt.Println("Processed ", objectHeight, " blocks. ", failCount, " failures")
		}
		ms := GetMerkleState(db, objectHeight)
		msPrior := GetMerkleState(db, objectHeight-1)

		object := ms.HashList[2] // Get the hash of the Factoid block

		AnchorState := GetMerkleState(db, DBHead)
		AnchorPrior := GetMerkleState(db, DBHead-1)
		if receipt, err := GetReceipt(object, DBHead); err != nil {
			fmt.Print(fmt.Sprintf("%v Receipt for %d %x", err, ms.GetCount(), object))
		} else {

			// Keeping these prints around because they make debugging easier
			if false {
				fmt.Print(receipt.String())
			}

			if !receipt.Validate() {
				if !fail {
					t.Error(fmt.Sprint(fmt.Sprintf("failed obj(%d) at objects: %d-%d for anchor at dbheight: %6d %x Objects: %d-%d",
						objectHeight,
						msPrior.GetCount(),
						ms.GetCount(),
						DBHead,
						object,
						AnchorPrior.GetCount(),
						AnchorState.GetCount())))
				}
				fail = true
				failCount++
			} else {
				fail = false
			}

		}
	}
}

func TestGetReceiptBig(t *testing.T) {

	var object [32]byte
	objectSlice, _ := hex.DecodeString("abf01aabe99bdf9adf70545c466ae9d7f509dfdda90519f1e08e8a77d140385c")
	copy(object[:], objectSlice)
	if receipt, err := GetReceipt(object, 289139); err != nil {
		t.Error(err)
	} else {
		fmt.Printf("Object %x\n", receipt.Object)
		fmt.Printf("Anchor %x\n", receipt.Anchor)
		fmt.Printf("DBHeight %d\n", receipt.ObjectDbheight)
		fmt.Printf("AnchorHeight %d\n", receipt.AnchorDbheight)
		working := object
		for _, v := range receipt.ApplyHashes {
			r := "L"
			if v.Right {
				r = "R"
				working = sha256.Sum256(append(working[:], v.Hash[:]...))
			} else {
				working = sha256.Sum256(append(v.Hash[:], working[:]...))
			}
			fmt.Printf(" Apply %s %x working: %x \n", r, v.Hash, working)
		}
	}

}
func TestPrintHashes(t *testing.T) {
	db := database.GetDB()
	for i := 0; i <= 1; i++ {
		merkleState := GetMerkleState(db, 0+int64(i))
		fmt.Println("Count: ", merkleState.GetCount())
		fmt.Println("Pending")
		for _, v := range merkleState.Pending {
			if v == nil {
				println("nil")
			} else {
				fmt.Printf("%x\n", *v)
			}
		}
		println("HashList")
		for _, v := range merkleState.HashList {
			fmt.Printf("%x\n", v)
		}
	}
}

func TestPrintTree(t *testing.T) {
	h1, _ := hex.DecodeString("c1fdd8c0a09e55e8b5f5e92eacb695d915b83af79735236fbd9274e37b26adc7")
	h2, _ := hex.DecodeString("3a4014a3ea9ccbe3f64b6178972e4a8828317cc3ffb491accf1ee4f8fcce747f")
	h3, _ := hex.DecodeString("a566023a9d7b824e4a12121ee38bc4d3c4987988f04eb8cfecc63570936d7c56")
	h4, _ := hex.DecodeString("0f426b2433103e606fb16b9340c14d88da8a02268af3c58b9a7896fa86e2904f")
	h5, _ := hex.DecodeString("6d00e14907c1b38e4882f5d6d7cfdb3595e6f85974404d43db4fc5c1a79381f4")
	h6, _ := hex.DecodeString("914333898b4cd3a87091ced94d6276090a1a266e1f4b7578e2b036cfaf9aaf3e")
	h7, _ := hex.DecodeString("5b756bf0ffcbfccd538e7ca7fb1b1e19c7a4285ff7aa3ae6c3a8ecd23ccaa32c")
	h8, _ := hex.DecodeString("2c4856fd7fbb208df8962a37ed9dbfbb766abc5c1b8ba0cba1e49b4c2c4b121d")
	h9, _ := hex.DecodeString("bc9747db86cb80a5ed8f03a6b59a2d4663f34eb1b5ece22daa27b4b3e13e1672")
	h10, _ := hex.DecodeString("92eb7a93907de6e48931c8dd8bceb4016461ae0e32f663d97400938730d1a715")
	h11, _ := hex.DecodeString("2ac11086b2ea3615940030482196cc9c86ef3798de93a7b9ad921b5784c4326b")

	h12 := sha256.Sum256(append(h1[:], h2[:]...))
	h34 := sha256.Sum256(append(h3[:], h4[:]...))
	h56 := sha256.Sum256(append(h5[:], h6[:]...))
	h78 := sha256.Sum256(append(h7[:], h8[:]...))
	h910 := sha256.Sum256(append(h9[:], h10[:]...))

	h1234 := sha256.Sum256(append(h12[:], h34[:]...))
	h5678 := sha256.Sum256(append(h56[:], h78[:]...))
	h91011 := sha256.Sum256(append(h910[:], h11[:]...))

	h12345678 := sha256.Sum256(append(h1234[:], h5678[:]...))

	anchor := sha256.Sum256(append(h12345678[:], h91011[:]...))

	fmt.Printf("%8s %v\n", "h1", h1[:8])
	fmt.Printf("%8s %v\n", "h2", h2[:8])
	fmt.Printf("%8s %v\n", "h3", h3[:8])
	fmt.Printf("%8s %v\n", "h4", h4[:8])
	fmt.Printf("%8s %v\n", "h5", h5[:8])
	fmt.Printf("%8s %v\n", "h6", h6[:8])
	fmt.Printf("%8s %v\n", "h7", h7[:8])
	fmt.Printf("%8s %v\n", "h8", h8[:8])
	fmt.Printf("%8s %v\n", "h9", h9[:8])
	fmt.Printf("%8s %v\n", "h10", h10[:8])
	fmt.Printf("%8s %v\n", "h11", h11[:8])
	fmt.Printf("%8s %v\n", "h12", h12[:8])
	fmt.Printf("%8s %v\n", "h34", h34[:8])
	fmt.Printf("%8s %v\n", "h56", h56[:8])
	fmt.Printf("%8s %v\n", "h78", h78[:8])
	fmt.Printf("%8s %v\n", "h910", h910[:8])
	fmt.Printf("%8s %v\n", "h1234", h1234[:8])
	fmt.Printf("%8s %v\n", "h5678", h5678[:8])
	fmt.Printf("%8s %v\n", "h91011", h91011[:8])
	fmt.Printf("%8s %v\n", "h12345678", h12345678[:8])
	fmt.Printf("%8s %v\n", "anchor", anchor[:8])
}
