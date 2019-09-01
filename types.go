package main

import (
	"bytes"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	sharedKeySize = 32
	deviceIDSize  = 32
)

type sharedKey [sharedKeySize]byte

func sharedKeyFromString(s string) (*sharedKey, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	if len(data) != sharedKeySize {
		return nil, fmt.Errorf("key has invalid size")
	}

	var key sharedKey
	copy(key[:], data)

	return &key, nil
}

func (k sharedKey) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

var deviceIDAllZeroes deviceID

type deviceID [deviceIDSize]byte

func (d deviceID) String() string {
	return base64.StdEncoding.EncodeToString(d[:])
}

func (d deviceID) IsValid() bool {
	return !bytes.Equal(d[:], deviceIDAllZeroes[:])
}

func newDeviceID() deviceID {
	var id deviceID
	if _, err := crypto_rand.Read(id[:]); err != nil {
		log.Fatalf("unable to read random data. err=%v", err)
	}
	return id
}

type copyRequest struct {
	DeviceID   []byte
	Signature  []byte
	Ciphertext []byte
}

func (r *copyRequest) GetDeviceID() deviceID {
	var id deviceID
	copy(id[:], r.DeviceID)
	return id
}

func (r copyRequest) Validate() error {
	if len(r.DeviceID) != deviceIDSize {
		return fmt.Errorf("DeviceID is empty or invalid")
	}
	if len(r.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("Signature is not the correct size")
	}
	if len(r.Ciphertext) < secretbox.Overhead {
		return fmt.Errorf("Ciphertext is too short")
	}
	return nil
}
