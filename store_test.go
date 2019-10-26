package main

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestMemStore(t *testing.T) {
	t.Run("add-multiple-copy", func(t *testing.T) {
		s := newMemStore()

		// Add 3 entries and expect all 3 to still be in the list
		// after calling CopyFirst

		s.Add([]byte("foo"))
		s.Add([]byte("bar"))
		s.Add([]byte("baz"))

		data, err := s.CopyFirst()
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))
		data, err = s.CopyFirst()
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))

		ids, err := s.ListAll()
		require.NoError(t, err)
		require.Len(t, ids, 3)
	})

	t.Run("add-remove", func(t *testing.T) {
		s := newMemStore()

		// Add 3 entries and expect all of them to be removed correctly

		id, err := s.Add([]byte("foobar"))
		require.NoError(t, err)
		id2, err := s.Add([]byte("foobar2"))
		require.NoError(t, err)
		id3, err := s.Add([]byte("foobar3"))
		require.NoError(t, err)

		do := func(i ulid.ULID, exp string) {
			tmp, err := s.Remove(i)
			require.NoError(t, err)
			require.Equal(t, exp, string(tmp))

			tmp, err = s.Remove(i)
			require.EqualError(t, err, errEntryNotFound.Error())
			require.Nil(t, tmp)
		}

		do(id, "foobar")
		do(id2, "foobar2")
		do(id3, "foobar3")
	})

	t.Run("add-multiple-pop", func(t *testing.T) {
		s := newMemStore()

		// Add 3 entries and expect all of them to be in the list
		// and to be removed in FIFO order

		s.Add([]byte("foo"))
		s.Add([]byte("bar"))
		s.Add([]byte("baz"))

		entries, err := s.ListAll()
		require.NoError(t, err)
		require.Len(t, entries, 3)

		tmp, err := s.RemoveFirst()
		require.NoError(t, err)
		require.Equal(t, "foo", string(tmp))

		tmp, err = s.RemoveFirst()
		require.NoError(t, err)
		require.Equal(t, "bar", string(tmp))

		tmp, err = s.RemoveFirst()
		require.NoError(t, err)
		require.Equal(t, "baz", string(tmp))

		tmp, err = s.RemoveFirst()
		require.NoError(t, err)
		require.Nil(t, tmp)
	})

	t.Run("add-copy", func(t *testing.T) {
		s := newMemStore()

		// Add 1 entry, copy it and expect it to stay in the store

		id, err := s.Add([]byte("nope"))
		require.NoError(t, err)

		tmp, err := s.Copy(id)
		require.NoError(t, err)
		require.Equal(t, []byte("nope"), tmp)
		tmp, err = s.Copy(id)
		require.NoError(t, err)
		require.Equal(t, []byte("nope"), tmp)

		var empty ulid.ULID
		tmp, err = s.Copy(empty)
		require.EqualError(t, err, errEntryNotFound.Error())
		require.Nil(t, tmp)
	})
}
