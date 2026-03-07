package discord

import (
	"crypto/ed25519"
	"encoding/hex"
)

func VerifySignature(publicKey, timestamp, body, signature string) bool {
	if publicKey == "" || timestamp == "" || signature == "" {
		return false
	}

	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return false
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false
	}

	message := []byte(timestamp + body)
	return ed25519.Verify(ed25519.PublicKey(pubKeyBytes), message, sigBytes)
}
