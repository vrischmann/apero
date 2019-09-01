package main

import (
	"errors"
	"sync"

	"rischmann.fr/apero/internal"
)

type store struct {
	mu   sync.Mutex
	keys map[internal.DeviceID]internal.PublicKey
}

func newStore() *store {
	return &store{
		keys: make(map[internal.DeviceID]internal.PublicKey),
	}
}

var errDeviceIDNotFound = errors.New("device id not found")

func (s *store) LookupPublicKey(id internal.DeviceID) (internal.PublicKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k, ok := s.keys[id]
	if !ok {
		return nil, errDeviceIDNotFound
	}

	return k, nil
}

func (s *store) AddPublicKey(id internal.DeviceID, key internal.PublicKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[id] = key

	return nil
}
