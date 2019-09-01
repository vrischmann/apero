package main

import (
	"fmt"

	"golang.org/x/crypto/ed25519"
	"rischmann.fr/apero/internal"
)

// copyRequest is a request to copy content into the staging
// area of the server.
//
// A copy request must be identified by a device id and signed
// by the public key for this device id.
// This means that before attempting a copy a client must register
// its device.
//
// The content must not be empty but there's no other constraint otherwise.
type copyRequest struct {
	DeviceID  internal.DeviceID
	Signature []byte
	Content   []byte
}

// Validate validates the request parameters.
func (r copyRequest) Validate() error {
	if len(r.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("Signature size is invalid")
	}
	if len(r.Content) == 0 {
		return fmt.Errorf("Content is empty")
	}
	return nil
}

type registerRequest struct {
	DeviceID  internal.DeviceID
	PublicKey internal.PublicKey
}

func (r registerRequest) Validate() error {
	return nil
}
