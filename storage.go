package main

import (
	"errors"

	"golang.org/x/crypto/ed25519"
)

type store struct {
}

func newStore() *store {
	return &store{}
}

var errDeviceIDNotFound = errors.New("device id not found")

func (s *store) LookupPublicKey(id deviceID) (ed25519.PublicKey, error) {
	panic("not implemented")
}

func (s *store) AddPublicKey(id deviceID, key ed25519.PublicKey) error {
	panic("not implemented")
}
