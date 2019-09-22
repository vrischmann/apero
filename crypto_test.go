package main

import (
	"encoding/json"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestKeyPairString(t *testing.T) {
	pub, priv, err := generateKeyPair()
	require.NoError(t, err)

	t.Run("public", func(t *testing.T) {
		var k publicKey
		err := (&k).UnmarshalText([]byte(pub.String()))
		require.NoError(t, err)
		require.Equal(t, pub, k)
	})

	t.Run("private", func(t *testing.T) {
		var k privateKey
		err := (&k).UnmarshalText([]byte(priv.String()))
		require.NoError(t, err)
		require.Equal(t, priv, k)
	})
}

func TestPublicKeyUnmarshalJSON(t *testing.T) {
	pub, _, err := generateKeyPair()
	require.NoError(t, err)

	var obj struct {
		Key publicKey
	}
	obj.Key = pub

	data, err := json.Marshal(obj)
	require.NoError(t, err)

	var obj2 struct {
		Key publicKey
	}
	err = json.Unmarshal(data, &obj2)
	require.NoError(t, err)
	require.Equal(t, pub, obj2.Key)
}

func TestKeyPairUnmarshalText(t *testing.T) {
	pub, priv, err := generateKeyPair()
	require.NoError(t, err)

	t.Run("public", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			s := pub.String()

			var key publicKey
			err := (&key).UnmarshalText([]byte(s))
			require.NoError(t, err)
			require.Equal(t, pub, key)
		})

		t.Run("toml", func(t *testing.T) {
			var obj struct {
				Key publicKey
			}

			md, err := toml.Decode(`Key = "`+pub.String()+`"`, &obj)
			require.NoError(t, err)
			require.Empty(t, md.Undecoded())
			require.Equal(t, pub, obj.Key)
		})
	})

	t.Run("private", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			s := priv.String()

			var key privateKey
			err := (&key).UnmarshalText([]byte(s))
			require.NoError(t, err)
			require.Equal(t, s, key.String())
		})

		t.Run("toml", func(t *testing.T) {
			var obj struct {
				Key privateKey
			}

			md, err := toml.Decode(`Key = "`+priv.String()+`"`, &obj)
			require.NoError(t, err)
			require.Empty(t, md.Undecoded())
		})
	})
}

func TestSecretBoxKeyUnmarshalText(t *testing.T) {
	const s = `WYBwj9jL9VxlaLlbpMPEMU3SJCgwh7fNVqJgSt74K38=`

	var key secretBoxKey
	err := (&key).UnmarshalText([]byte(s))
	require.NoError(t, err)
	require.Equal(t, s, key.String())
}

func TestSecretBox(t *testing.T) {
	k := newSecretBoxKey()

	data := []byte("foobar")

	box := secretBoxSeal(data, k)
	decrypted, ok := secretBoxOpen(box, k)
	require.True(t, ok, "expected to open the box")
	require.Equal(t, data, decrypted)
}
