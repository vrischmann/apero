package internal

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"golang.org/x/crypto/ed25519"
)

const (
	// SharedKeySize is the size of the key shared between server
	// and clients. It must be equal to what is used by secretbox.
	// See https://godoc.org/golang.org/x/crypto/nacl/secretbox#Seal.
	SharedKeySize = 32

	// PublicKeySize is the size of a public key part of a key pair.
	// It must always be equal to ed25519.PublicKeySize.
	PublicKeySize = ed25519.PublicKeySize
)

// PublicKey is the public key part of a key pair.
// We redefined the type so we can implement encoding.TextUnmarshaler.
type PublicKey ed25519.PublicKey

func (k *PublicKey) String() string {
	return base64.StdEncoding.EncodeToString(*k)
}

func (k *PublicKey) UnmarshalText(s string) error {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	if len(data) != PublicKeySize {
		return fmt.Errorf("invalid shared key size")
	}

	copy((*k)[:], data)

	return nil
}

// SharedKey is the key shared between server and clients
// to encrypt and authenticate messages.
//
// It is _not_ used to
type SharedKey [SharedKeySize]byte

// NewSharedKey creates a new, random device id.
func NewSharedKey() SharedKey {
	var id SharedKey
	if _, err := crypto_rand.Read(id[:]); err != nil {
		log.Fatalf("unable to read random data. err=%v", err)
	}
	return id
}

// SharedKeyFromString parses a string as a shared key.
func SharedKeyFromString(s string) (*SharedKey, error) {
	var key SharedKey

	if err := (&key).UnmarshalText(s); err != nil {
		return nil, err
	}
	return &key, nil
}

// UnmarshalText implements encoding.TextUnmarshaler
// It assumes the string is base64 encoded.
func (k *SharedKey) UnmarshalText(s string) error {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	if len(data) != SharedKeySize {
		return fmt.Errorf("invalid shared key size")
	}

	copy((*k)[:], data)

	return nil
}

// String returns the key as a base64 encoded string.
func (k SharedKey) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

func VerifySignature(pk PublicKey, content, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(pk), content, signature)
}
