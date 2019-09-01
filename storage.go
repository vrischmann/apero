package main

import (
	"errors"
	"sync"

	"golang.org/x/crypto/ed25519"
)

type store struct {
	mu   sync.Mutex
	keys map[deviceID]ed25519.PublicKey
}

func newStore() *store {
	return &store{
		keys: make(map[deviceID]ed25519.PublicKey),
	}
}

var errDeviceIDNotFound = errors.New("device id not found")

func (s *store) LookupPublicKey(id deviceID) (ed25519.PublicKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k, ok := s.keys[id]
	if !ok {
		return nil, errDeviceIDNotFound
	}

	return k, nil
}

func (s *store) AddPublicKey(id deviceID, key ed25519.PublicKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[id] = key

	return nil
}
