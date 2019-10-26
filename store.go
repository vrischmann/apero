package main

import (
	crypto_rand "crypto/rand"
	"errors"
	"fmt"
	"sort"
	"strings"
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
	CopyFirst() ([]byte, error)
	Copy(id ulid.ULID) ([]byte, error)
	RemoveFirst() ([]byte, error)
	Remove(id ulid.ULID) ([]byte, error)
	ListAll() ([]ulid.ULID, error)
}

var errEntryNotFound = errors.New("entry not found")

type memStoreEntry struct {
	id      ulid.ULID
	content []byte
}

func (e memStoreEntry) String() string {
	return fmt.Sprintf("{id: %s, content: %s}", e.id.String(), string(e.content))
}

func (e *memStoreEntry) IsValid() bool {
	return len(e.content) > 0 && !isEmptyULID(e.id)
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

func (s *memStore) Dump() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var builder strings.Builder
	for _, entry := range s.entries {
		fmt.Fprintf(&builder, "entry: %s\n", entry)
	}

	return builder.String()
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

func (s *memStore) CopyFirst() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) < 1 {
		return nil, nil
	}

	entry := s.entries[0]

	tmp := make([]byte, len(entry.content))
	copy(tmp, entry.content)

	return tmp, nil
}

func (s *memStore) Copy(id ulid.ULID) ([]byte, error) {
	s.mu.Lock()

	var entry memStoreEntry
	for _, v := range s.entries {
		if v.id == id {
			entry = v
			break
		}
	}

	s.mu.Unlock()

	//

	if !entry.IsValid() {
		return nil, errEntryNotFound
	}

	tmp := make([]byte, len(entry.content))
	copy(tmp, entry.content)

	return tmp, nil
}

func (s *memStore) RemoveFirst() ([]byte, error) {
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

	var (
		pos   int
		entry memStoreEntry
	)
	for i, v := range s.entries {
		if v.id == id {
			pos = i
			entry = v
			break
		}
	}

	//

	if !entry.IsValid() {
		return nil, errEntryNotFound
	}

	s.entries = append(s.entries[:pos], s.entries[pos+1:]...)

	return entry.content, nil
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

var _ store = (*memStore)(nil)
