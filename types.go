package main

import (
	"fmt"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/ed25519"
)

// copyRequest is a request to copy content to the server.
//
// A copy request must be signed by the client public key.
// This means that before attempting a copy a client must register
// its device.
//
// The content must not be empty but there's no other constraint otherwise.
type copyRequest struct {
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

type moveRequest struct {
	Signature []byte
	ID        ulid.ULID
}

// Validate validates the request parameters.
func (r moveRequest) Validate() error {
	if len(r.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("Signature size is invalid")
	}
	return nil
}

type pasteRequest struct {
	Signature []byte
	ID        ulid.ULID
}

// Validate validates the request parameters.
func (r pasteRequest) Validate() error {
	if len(r.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("Signature size is invalid")
	}
	return nil
}

type listRequest struct {
	Signature []byte
}

// Validate validates the request parameters.
func (r listRequest) Validate() error {
	if len(r.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("Signature size is invalid")
	}
	return nil
}

type listResponse struct {
	Entries []ulid.ULID
}
