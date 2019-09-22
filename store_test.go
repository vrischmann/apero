package main

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestMemStore(t *testing.T) {
	s := newMemStore()

	t.Run("add-remove", func(t *testing.T) {
		id, err := s.Add([]byte("foobar"))
		require.NoError(t, err)

		tmp, err := s.Remove(id)
		require.NoError(t, err)
		require.Equal(t, "foobar", string(tmp))

		tmp, err = s.Remove(id)
		require.NoError(t, err)
		require.Nil(t, tmp)
	})

	t.Run("add-multiple-pop", func(t *testing.T) {
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
		id, err := s.Add([]byte("nope"))
		require.NoError(t, err)

		tmp, err := s.Copy(id)
		require.NoError(t, err)
		require.Equal(t, []byte("nope"), tmp)

		var empty ulid.ULID
		tmp, err = s.Copy(empty)
		require.NoError(t, err)
		require.Nil(t, tmp)
	})
}
