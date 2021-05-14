package anchormaker

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/serveridentity/identity"
)

const (
	AnchorMakerEntryTag = "Anchor Maker"
)

// AnchorMakerEntry represents the content of the Factom Entry that is used for promotions/demotions
type AnchorMakerEntryContent struct {
	IdentityChainID string `json:"identitychainid"`
	ANOIdentity     string `json:"ano"`
	Time            int64  `json:"time"`
}

func PromoteAnchorMaker(AnchorMakerIdentity string, ANOIdentity string, ANOIdentitySK string) (*factom.Entry, error) {

	var err error

	// validate existence of Anchormaker identity chainID
	chainExist := factom.ChainExists(AnchorMakerIdentity)
	if !chainExist {
		return nil, fmt.Errorf("anchor maker Identity ChainID %s does not exist", AnchorMakerIdentity)
	}

	// validate existence of ANO identity chainID
	chainExist = factom.ChainExists(ANOIdentity)
	if !chainExist {
		return nil, fmt.Errorf("ano Identity ChainID %s does not exist", ANOIdentity)
	}

	content := &AnchorMakerEntryContent{}
	content.IdentityChainID = AnchorMakerIdentity
	content.ANOIdentity = ANOIdentity
	content.Time = time.Now().Unix()

	entry := &factom.Entry{}
	entry.Content, err = json.Marshal(content)
	if err != nil {
		return nil, err
	}

	id := identity.NewIdentity()

	seedPriv, err := extractSecretFromIdentityKey(ANOIdentitySK)
	if err != nil {
		return nil, err
	}

	id.GenerateIdentityFromPrivateKey(&seedPriv, 1)

	sig := ed25519.Sign(&seedPriv, entry.Content)
	sigString := hex.EncodeToString((*sig)[:])

	entry.ExtIDs = append(entry.ExtIDs, []byte(AnchorMakerEntryTag), []byte(sigString))

	return entry, nil

}

func extractSecretFromIdentityKey(key string) ([ed25519.PrivateKeySize]byte, error) {

	var seedPriv [ed25519.PrivateKeySize]byte
	var seedKey []byte
	var err error

	// if key is in sk format
	if len(key) == 53 && strings.Compare(key[:2], "sk") == 0 {

		// check sk level range (1-4)
		if lev, err := strconv.Atoi(key[2:3]); err != nil {
			return seedPriv, fmt.Errorf("error in input: " + err.Error())
		} else if lev < 1 || lev > 4 {
			return seedPriv, fmt.Errorf("key level is outside range (1-4)")
		}

		p := base58.Decode(key[:53])
		if !identity.CheckHumanReadable(p[:]) {
			return seedPriv, fmt.Errorf("not a valid private key, end hash is incorrect")
		}

		seedKey = p[3:35]
		copy(seedPriv[:32], seedKey[:32])

	} else {

		seedKey, err = hex.DecodeString(key)
		if err != nil {
			return seedPriv, err
		}
		copy(seedPriv[:32], seedKey[:32])

	}

	return seedPriv, nil

}
