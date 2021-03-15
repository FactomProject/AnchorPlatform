package factomSync

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/AccumulateNetwork/SMT/smt"
)

func TestMarshal(t *testing.T) {
	MS1 := new(smt.MerkleState)
	MS1.InitSha256()
	MS2 := new(smt.MerkleState)
	MS2.InitSha256()

	data1 := MS1.Marshal()
	MS2.UnMarshal(data1)
	data2 := MS2.Marshal()
	if !bytes.Equal(data1, data2) {
		t.Error("Should be the same")
	}
	fmt.Printf("%x\n%x\n", data1, data2)

	MS1.AddToChain(MS1.HashFunction([]byte{1, 2, 3, 4, 5}))

	data1 = MS1.Marshal()
	MS2.UnMarshal(data1)
	data2 = MS2.Marshal()
	if !MS1.Equal(*MS2) {
		t.Error("Should be the same")
	}
	if !bytes.Equal(data1, data2) {
		t.Error("Should be the same")
	}
	fmt.Printf("%x\n%x\n", data1, data2)

	for i := 0; i < 1; i++ {
		MS1.AddToChain(MS1.HashFunction([]byte(fmt.Sprintf("%8d", i))))
	}

	data1 = MS1.Marshal()
	MS2.UnMarshal(data1)
	data2 = MS2.Marshal()
	if !bytes.Equal(data1, data2) {
		t.Error("Should be the same")
	}
	fmt.Printf("%x\n%x\n", data1, data2)
}
