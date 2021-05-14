package anchormaker

import (
	"encoding/hex"
	"testing"

	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/serveridentity/identity"
	"github.com/stretchr/testify/assert"
)

// PromoteAnchorMaker does NOT check if ANO Identity SK is valid for specific ANO identity and this ANO is promoted
// Anchor Platform will be validating it during parsing of anchormaker promotions entries

func TestPromoteAnchorMaker(t *testing.T) {

	factom.SetFactomdServer("https://api.factomd.net")

	var err error
	ANOIdentity := "888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f4"
	AnchorMakerIdentity := "888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f4"
	ANOIdentitySK := "sk11pz4AG9XgB1eNVkbppYAWsgyg7sftDXqBASsagKJqvVRKYodCU"

	id := identity.NewIdentity()

	seedPriv, err := extractSecretFromIdentityKey(ANOIdentitySK)
	assert.NoError(t, err)

	id.GenerateIdentityFromPrivateKey(&seedPriv, 1)
	pubk := id.GetPublicKey()

	anchorMakerEntry, err := PromoteAnchorMaker(AnchorMakerIdentity, ANOIdentity, ANOIdentitySK)
	assert.NoError(t, err)
	assert.NotNil(t, anchorMakerEntry)

	sigBytes, err := hex.DecodeString(string(anchorMakerEntry.ExtIDs[1]))
	assert.NoError(t, err)

	var sig [ed25519.SignatureSize]byte
	copy(sig[:ed25519.SignatureSize], sigBytes[:ed25519.SignatureSize])

	sigIsValid := ed25519.VerifyCanonical(pubk, anchorMakerEntry.Content, &sig)

	assert.True(t, sigIsValid)

}
