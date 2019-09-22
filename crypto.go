package main

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
	// secretBoxKeySize is the size of the key shared between server
	// and clients. It must be equal to what is used by secretbox.
	// See https://godoc.org/golang.org/x/crypto/nacl/secretbox#Seal.
	secretBoxKeySize = 32

	// publicKeySize is the size of a public key part of a key pair.
	// It must always be equal to ed25519.publicKeySize.
	publicKeySize = ed25519.PublicKeySize

	// privateKeySize is the size of a private key part of a key pair.
	// It must always be equal to ed25519.privateKeySize.
	privateKeySize = ed25519.PrivateKeySize
)

// generateKeyPair generates a public/private key pair.
// It uses ed25519 under the hood.
func generateKeyPair() (publicKey, privateKey, error) {
	pub, priv, err := ed25519.GenerateKey(crypto_rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return publicKey(pub), privateKey(priv), nil
}

// publicKey is the public key part of a key pair.
// We redefined the type so we can implement encoding.TextUnmarshaler.
type publicKey ed25519.PublicKey

func (k publicKey) String() string {
	return base64.StdEncoding.EncodeToString(k)
}

func (k publicKey) IsValid() bool {
	return len(k) == publicKeySize
}

// UnmarshalJSON implements json.Unmarshaler
func (k *publicKey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return k.UnmarshalText([]byte(s))
}

// MarshalText implements encoding.TextMarshaler
func (k publicKey) MarshalText() ([]byte, error) {
	s := k.String()
	return []byte(s), nil
}

// UnmarshalTExt implements encoding.TextUnmarshaler
func (k *publicKey) UnmarshalText(p []byte) error {
	data, err := base64.StdEncoding.DecodeString(string(p))
	if err != nil {
		return err
	}

	if len(data) != publicKeySize {
		return fmt.Errorf("invalid public key size")
	}

	*k = make(publicKey, publicKeySize)

	copy((*k)[:], data)

	return nil
}

// privateKey is the public key part of a key pair.
// We redefined the type so we can implement encoding.TextUnmarshaler.
type privateKey ed25519.PrivateKey

func (k privateKey) String() string {
	return base64.StdEncoding.EncodeToString(k)
}

func (k privateKey) IsValid() bool {
	return len(k) == privateKeySize
}

// MarshalText implements encoding.TextMarshaler
func (k privateKey) MarshalText() ([]byte, error) {
	s := k.String()
	return []byte(s), nil
}

// UnmarshalTExt implements encoding.TextUnmarshaler
func (k *privateKey) UnmarshalText(p []byte) error {
	data, err := base64.StdEncoding.DecodeString(string(p))
	if err != nil {
		return err
	}

	if len(data) != privateKeySize {
		return fmt.Errorf("invalid private key size")
	}

	*k = make(privateKey, privateKeySize)

	copy((*k)[:], data)

	return nil
}

// secretBoxKey is the key shared between server and clients
// to encrypt and authenticate messages.
//
// It is _not_ used to
type secretBoxKey [secretBoxKeySize]byte

// newSecretBoxKey creates a new, random device id.
func newSecretBoxKey() secretBoxKey {
	var id secretBoxKey
	if _, err := crypto_rand.Read(id[:]); err != nil {
		log.Fatalf("unable to read random data. err=%v", err)
	}
	return id
}

// secretBoxKeyFromString parses a string as a shared key.
func secretBoxKeyFromString(s string) (*secretBoxKey, error) {
	var key secretBoxKey

	if err := (&key).UnmarshalText([]byte(s)); err != nil {
		return nil, err
	}
	return &key, nil
}

func (k secretBoxKey) IsValid() bool {
	return len(k) == secretBoxKeySize
}

// MarshalText implements encoding.TextMarshaler
func (k secretBoxKey) MarshalText() ([]byte, error) {
	s := k.String()
	return []byte(s), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
// It assumes the string is base64 encoded.
func (k *secretBoxKey) UnmarshalText(p []byte) error {
	data, err := base64.StdEncoding.DecodeString(string(p))
	if err != nil {
		return err
	}

	if len(data) != secretBoxKeySize {
		return fmt.Errorf("invalid secret box key size")
	}

	copy((*k)[:], data)

	return nil
}

// String returns the key as a base64 encoded string.
func (k secretBoxKey) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

func secretBoxSeal(data []byte, key secretBoxKey) []byte {
	nonce := getNonce()
	encrypted := secretbox.Seal(nonce[:], data, &nonce, (*[32]byte)(&key))

	return encrypted
}

func secretBoxOpen(box []byte, key secretBoxKey) ([]byte, bool) {
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

func sign(priv privateKey, content []byte) []byte {
	return ed25519.Sign(ed25519.PrivateKey(priv), content)
}

func verify(pk publicKey, content, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(pk), content, signature)
}

var (
	_ encoding.TextUnmarshaler = (*publicKey)(nil)
	_ encoding.TextUnmarshaler = (*privateKey)(nil)
	_ encoding.TextUnmarshaler = (*secretBoxKey)(nil)
)
