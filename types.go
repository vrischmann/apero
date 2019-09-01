package main

import (
	"bytes"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"golang.org/x/crypto/ed25519"
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

func deviceIDFromBytes(p []byte) deviceID {
	if len(p) != deviceIDSize {
		panic("invalid device id size")
	}

	var id deviceID
	copy(id[:], p)

	return id
}

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
	DeviceID  []byte
	Signature []byte
	Content   []byte
}

func (r copyRequest) Validate() error {
	if len(r.DeviceID) != deviceIDSize {
		return fmt.Errorf("DeviceID is empty or invalid")
	}
	if len(r.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("Signature size is invalid")
	}
	if len(r.Content) == 0 {
		return fmt.Errorf("Content is empty")
	}
	return nil
}

type registerRequest struct {
	DeviceID  []byte
	PublicKey ed25519.PublicKey
}

func (r registerRequest) Validate() error {
	if len(r.DeviceID) != deviceIDSize {
		return fmt.Errorf("DeviceID is empty or invalid")
	}
	if len(r.PublicKey) < ed25519.PublicKeySize {
		return fmt.Errorf("PublicKey size is invalid")
	}
	return nil
}
