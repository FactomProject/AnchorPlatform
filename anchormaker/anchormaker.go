package anchormaker

import (
	factomdid "github.com/DeFacto-Team/go-factom-did"
	"github.com/FactomProject/factom"
)

// NewDID generated W3C DID for Anchor Maker
func NewDID(extID []byte) (*factom.Entry, error) {

	did := factomdid.NewDID()

	didKey, _ := factomdid.NewDIDKey("public-0", factomdid.KeyTypeECDSA)
	didKey.AddPurpose(factomdid.KeyPurposePublic)
	mgmtKey, _ := factomdid.NewManagementKey("admin-0", factomdid.KeyTypeECDSA, 0)

	did.AddDIDKey(didKey)
	did.AddManagementKey(mgmtKey)

	entry, err := did.Create()
	if err != nil {
		return nil, err
	}

	if extID != nil {
		entry.ExtIDs[3] = extID
	}

	return entry, nil

}
