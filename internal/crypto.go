package internal

import (
	crypto_rand "crypto/rand"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// SecretBoxKeySize is the size of the key shared between server
	// and clients. It must be equal to what is used by secretbox.
	// See https://godoc.org/golang.org/x/crypto/nacl/secretbox#Seal.
	SecretBoxKeySize = 32

	// PublicKeySize is the size of a public key part of a key pair.
	// It must always be equal to ed25519.PublicKeySize.
	PublicKeySize = ed25519.PublicKeySize

	// PrivateKeySize is the size of a private key part of a key pair.
	// It must always be equal to ed25519.PrivateKeySize.
	PrivateKeySize = ed25519.PrivateKeySize
)

// GenerateKeyPair generates a public/private key pair.
// It uses ed25519 under the hood.
func GenerateKeyPair() (PublicKey, PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(crypto_rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return PublicKey(publicKey), PrivateKey(privateKey), nil
}

// PublicKey is the public key part of a key pair.
// We redefined the type so we can implement encoding.TextUnmarshaler.
type PublicKey ed25519.PublicKey

func (k PublicKey) String() string {
	return base64.StdEncoding.EncodeToString(k)
}

// UnmarshalJSON implements json.Unmarshaler
func (k *PublicKey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return k.UnmarshalText([]byte(s))
}

// UnmarshalTExt implements encoding.TextUnmarshaler
func (k *PublicKey) UnmarshalText(p []byte) error {
	data, err := base64.StdEncoding.DecodeString(string(p))
	if err != nil {
		return err
	}

	if len(data) != PublicKeySize {
		return fmt.Errorf("invalid public key size")
	}

	*k = make(PublicKey, PublicKeySize)

	copy((*k)[:], data)

	return nil
}

// PrivateKey is the public key part of a key pair.
// We redefined the type so we can implement encoding.TextUnmarshaler.
type PrivateKey ed25519.PrivateKey

func (k PrivateKey) String() string {
	return base64.StdEncoding.EncodeToString(k)
}

// UnmarshalTExt implements encoding.TextUnmarshaler
func (k *PrivateKey) UnmarshalText(p []byte) error {
	data, err := base64.StdEncoding.DecodeString(string(p))
	if err != nil {
		return err
	}

	if len(data) != PrivateKeySize {
		return fmt.Errorf("invalid private key size")
	}

	*k = make(PrivateKey, PrivateKeySize)

	copy((*k)[:], data)

	return nil
}

// SecretBoxKey is the key shared between server and clients
// to encrypt and authenticate messages.
//
// It is _not_ used to
type SecretBoxKey [SecretBoxKeySize]byte

// NewSecretBoxKey creates a new, random device id.
func NewSecretBoxKey() SecretBoxKey {
	var id SecretBoxKey
	if _, err := crypto_rand.Read(id[:]); err != nil {
		log.Fatalf("unable to read random data. err=%v", err)
	}
	return id
}

// SecretBoxKeyFromString parses a string as a shared key.
func SecretBoxKeyFromString(s string) (*SecretBoxKey, error) {
	var key SecretBoxKey

	if err := (&key).UnmarshalText([]byte(s)); err != nil {
		return nil, err
	}
	return &key, nil
}

// UnmarshalText implements encoding.TextUnmarshaler
// It assumes the string is base64 encoded.
func (k *SecretBoxKey) UnmarshalText(p []byte) error {
	data, err := base64.StdEncoding.DecodeString(string(p))
	if err != nil {
		return err
	}

	if len(data) != SecretBoxKeySize {
		return fmt.Errorf("invalid secret box key size")
	}

	copy((*k)[:], data)

	return nil
}

// String returns the key as a base64 encoded string.
func (k SecretBoxKey) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

func SecretBoxSeal(data []byte, key SecretBoxKey) []byte {
	nonce := getNonce()
	encrypted := secretbox.Seal(nonce[:], data, &nonce, (*[32]byte)(&key))

	return encrypted
}

func SecretBoxOpen(box []byte, key SecretBoxKey) ([]byte, bool) {
	var nonce [24]byte
	copy(nonce[:], box[:24])

	box = box[24:]

	return secretbox.Open(nil, box, &nonce, (*[32]byte)(&key))
}

func getNonce() [24]byte {
	var nonce [24]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	return nonce
}

func VerifySignature(pk PublicKey, content, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(pk), content, signature)
}

var (
	_ encoding.TextUnmarshaler = (*PublicKey)(nil)
	_ encoding.TextUnmarshaler = (*PrivateKey)(nil)
	_ encoding.TextUnmarshaler = (*SecretBoxKey)(nil)
)
