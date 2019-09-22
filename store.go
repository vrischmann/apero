package main

import (
	crypto_rand "crypto/rand"
	"sort"
	"sync"

	"github.com/oklog/ulid/v2"
)

var ulidEntropy struct {
	mu      sync.Mutex
	entropy *ulid.MonotonicEntropy
}

func init() {
	ulidEntropy.entropy = ulid.Monotonic(crypto_rand.Reader, 0)
}

func newULID() ulid.ULID {
	ulidEntropy.mu.Lock()
	id := ulid.MustNew(ulid.Now(), ulidEntropy.entropy)
	ulidEntropy.mu.Unlock()

	return id
}

type store interface {
	Add(data []byte) (ulid.ULID, error)
	Pop() ([]byte, error)
	Remove(id ulid.ULID) ([]byte, error)
	ListAll() ([]ulid.ULID, error)
}

type memStoreEntry struct {
	id      ulid.ULID
	content []byte
}

type memStore struct {
	mu      sync.Mutex
	entries []memStoreEntry
}

func newMemStore() *memStore {
	return &memStore{
		entries: make([]memStoreEntry, 0, 32),
	}
}

func (s *memStore) Add(data []byte) (ulid.ULID, error) {
	entry := memStoreEntry{
		id:      newULID(),
		content: data,
	}

	s.mu.Lock()
	s.entries = append(s.entries, entry)
	s.mu.Unlock()

	return entry.id, nil
}

func (s *memStore) Pop() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) < 1 {
		return nil, nil
	}

	entry := s.entries[0]
	s.entries = s.entries[1:]

	return entry.content, nil
}

func (s *memStore) Remove(id ulid.ULID) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos := sort.Search(len(s.entries), func(i int) bool {
		return s.entries[i].id == id
	})

	if pos < len(s.entries) && s.entries[pos].id == id {
		entry := s.entries[pos]

		s.entries = append(s.entries[:pos], s.entries[pos+1:]...)

		return entry.content, nil
	}

	return nil, nil
}

func (s *memStore) ListAll() ([]ulid.ULID, error) {
	ids := make([]ulid.ULID, 0, 32)

	s.mu.Lock()
	for _, entry := range s.entries {
		ids = append(ids, entry.id)
	}
	s.mu.Unlock()

	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Compare(ids[j]) < 0
	})

	return ids, nil
}
