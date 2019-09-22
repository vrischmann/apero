package main

import (
	"testing"

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

		tmp, err := s.Pop()
		require.NoError(t, err)
		require.Equal(t, "foo", string(tmp))

		tmp, err = s.Pop()
		require.NoError(t, err)
		require.Equal(t, "bar", string(tmp))

		tmp, err = s.Pop()
		require.NoError(t, err)
		require.Equal(t, "baz", string(tmp))

		tmp, err = s.Pop()
		require.NoError(t, err)
		require.Nil(t, tmp)
	})
}
