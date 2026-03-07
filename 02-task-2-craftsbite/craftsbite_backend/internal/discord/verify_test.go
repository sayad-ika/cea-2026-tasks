package discord_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/sayad-ika/craftsbite/internal/discord"
)

func generateKeyPair(t *testing.T) (pubHex, privKey string, priv ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key pair: %v", err)
	}
	return hex.EncodeToString(pub), hex.EncodeToString(priv), priv
}

func sign(t *testing.T, priv ed25519.PrivateKey, timestamp, body string) string {
	t.Helper()
	msg := []byte(timestamp + body)
	sig := ed25519.Sign(priv, msg)
	return hex.EncodeToString(sig)
}

func TestVerifySignature(t *testing.T) {
	pubHex, _, priv := generateKeyPair(t)

	const timestamp = "1741366800"
	const body = `{"type":1,"id":"1234567890"}`

	validSig := sign(t, priv, timestamp, body)

	wrongPubHex, _, _ := generateKeyPair(t)

	tests := []struct {
		name      string
		publicKey string
		timestamp string
		body      string
		signature string
		want      bool
	}{
		{
			name:      "valid signature passes",
			publicKey: pubHex,
			timestamp: timestamp,
			body:      body,
			signature: validSig,
			want:      true,
		},
		{
			name:      "tampered body is rejected",
			publicKey: pubHex,
			timestamp: timestamp,
			body:      `{"type":1,"id":"TAMPERED"}`,
			signature: validSig,
			want:      false,
		},
		{
			name:      "wrong public key is rejected",
			publicKey: wrongPubHex,
			timestamp: timestamp,
			body:      body,
			signature: validSig,
			want:      false,
		},
		{
			name:      "missing signature is rejected",
			publicKey: pubHex,
			timestamp: timestamp,
			body:      body,
			signature: "",
			want:      false,
		},
		{
			name:      "missing public key is rejected",
			publicKey: "",
			timestamp: timestamp,
			body:      body,
			signature: validSig,
			want:      false,
		},
		{
			name:      "missing timestamp is rejected",
			publicKey: pubHex,
			timestamp: "",
			body:      body,
			signature: validSig,
			want:      false,
		},
		{
			name:      "invalid hex signature is rejected",
			publicKey: pubHex,
			timestamp: timestamp,
			body:      body,
			signature: "not-hex!",
			want:      false,
		},
		{
			name:      "invalid hex public key is rejected",
			publicKey: "not-hex!",
			timestamp: timestamp,
			body:      body,
			signature: validSig,
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := discord.VerifySignature(tc.publicKey, tc.timestamp, tc.body, tc.signature)
			if got != tc.want {
				t.Errorf("VerifySignature() = %v, want %v", got, tc.want)
			}
		})
	}
}
