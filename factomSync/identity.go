package factomSync

import (
	"fmt"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
)

// ProcessIdentity
// Digs through the Admin block for the given directory block.  Returns true if any
// material updates are found for the identities in Factom.
func (s *Sync) ProcessIdentity(DBlock *factom.DBlock) (foundUpdates bool) {

	if AdminBlock, err := factom.GetRaw(DBlock.DBEntries[0].KeyMR); err != nil {
		panic(err)
	} else {
		aBlock := new(adminBlock.AdminBlock)
		err = aBlock.UnmarshalBinary(AdminBlock)
		if err != nil {
			panic(fmt.Sprint("failed to unmarshal the admin block at height ", DBlock.SequenceNumber))
		}

		if aBlock.ABEntries != nil {
			prev := s.GetANOList()
			var removed []string
			for _, v := range aBlock.ABEntries {
				switch v.Type() {
				case constants.TYPE_REMOVE_FED_SERVER:
					var e = v.(*adminBlock.RemoveFederatedServer)
					chainID := e.IdentityChainID.String()
					if s.CurrentANOs[chainID] == nil {
						fmt.Println("***************** should have an ANO if we are removing it: ", chainID)
					}
					removed = append(removed, chainID)
				case constants.TYPE_ADD_AUDIT_SERVER:
					var e = v.(*adminBlock.AddAuditServer)
					chainID := e.IdentityChainID.String()
					if s.CurrentANOs[chainID] == nil {
						ano := new(ANO)
						ano.ChainID = chainID
						s.CurrentANOs[chainID] = ano
					}
				case constants.TYPE_ADD_FED_SERVER:
					var e = v.(*adminBlock.AddFederatedServer)
					chainID := e.IdentityChainID.String()
					if s.CurrentANOs[chainID] == nil {
						ano := new(ANO)
						ano.ChainID = chainID
						s.CurrentANOs[chainID] = ano
					}
				}
			}

			for _, v := range removed {
				delete(s.CurrentANOs, v)
			}

			result := s.GetANOList()
			if len(prev) == len(result) {
				for i, v := range result {
					if v.ChainID != prev[i].ChainID {
						foundUpdates = true
					}
				}
			} else {
				foundUpdates = true
			}
		}
	}
	return foundUpdates
}
