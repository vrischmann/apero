package internal

import (
	"bytes"
	crypto_rand "crypto/rand"
	"encoding"
	"encoding/base64"
	"fmt"
	"log"
)

const (
	// DeviceIDSize is the size of a device id.
	DeviceIDSize = 32
)

var deviceIDAllZeroes DeviceID

type DeviceID [DeviceIDSize]byte

// NewDeviceID creates a new, random device id.
func NewDeviceID() DeviceID {
	var id DeviceID
	if _, err := crypto_rand.Read(id[:]); err != nil {
		log.Fatalf("unable to read random data. err=%v", err)
	}
	return id
}

// deviceIDFromBytes creates a DeviceID
// from a slice of bytes.
func DeviceIDFromBytes(p []byte) DeviceID {
	if len(p) != DeviceIDSize {
		panic("invalid device id size")
	}

	var id DeviceID
	copy(id[:], p)

	return id
}

// UnmarshalText implements encoding.TextUnmarshaler
// It assumes the string is base64 encoded.
func (d *DeviceID) UnmarshalText(p []byte) error {
	data, err := base64.StdEncoding.DecodeString(string(p))
	if err != nil {
		return err
	}

	if len(data) != DeviceIDSize {
		return fmt.Errorf("invalid device id size")
	}

	copy((*d)[:], data)

	return nil
}

func (d DeviceID) String() string {
	return base64.StdEncoding.EncodeToString(d[:])
}

func (d DeviceID) IsValid() bool {
	return !bytes.Equal(d[:], deviceIDAllZeroes[:])
}

var _ encoding.TextUnmarshaler = (*DeviceID)(nil)
