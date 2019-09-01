package internal

import (
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestKeyPairString(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("public", func(t *testing.T) {
		var k PublicKey
		err := (&k).UnmarshalText([]byte(pub.String()))
		if err != nil {
			t.Fatal(err)
		}

		if exp, got := pub, k; !bytes.Equal(exp, got) {
			t.Fatalf("expected %q but got %q", exp, got)
		}
	})

	t.Run("private", func(t *testing.T) {
		var k PrivateKey
		err := (&k).UnmarshalText([]byte(priv.String()))
		if err != nil {
			t.Fatal(err)
		}

		if exp, got := priv, k; !bytes.Equal(exp, got) {
			t.Fatalf("expected %q but got %q", exp, got)
		}
	})
}

func TestKeyPairUnmarshalText(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("public", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			s := pub.String()

			var key PublicKey
			err := (&key).UnmarshalText([]byte(s))
			if err != nil {
				t.Fatal(err)
			}

			if exp, got := s, key.String(); exp != got {
				t.Fatalf("expected %q, got %q", exp, got)
			}
		})

		t.Run("toml", func(t *testing.T) {
			var obj struct {
				Key PublicKey
			}

			md, err := toml.Decode(`Key = "`+pub.String()+`"`, &obj)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(md.Undecoded()); got > 0 {
				t.Fatal("expected no undecoded keys")
			}
		})
	})

	t.Run("private", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			s := priv.String()

			var key PrivateKey
			err := (&key).UnmarshalText([]byte(s))
			if err != nil {
				t.Fatal(err)
			}

			if exp, got := s, key.String(); exp != got {
				t.Fatalf("expected %q, got %q", exp, got)
			}
		})

		t.Run("toml", func(t *testing.T) {
			var obj struct {
				Key PrivateKey
			}

			md, err := toml.Decode(`Key = "`+priv.String()+`"`, &obj)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(md.Undecoded()); got > 0 {
				t.Fatal("expected no undecoded keys")
			}
		})
	})
}

func TestSecretBoxKeyUnmarshalText(t *testing.T) {
	const s = `WYBwj9jL9VxlaLlbpMPEMU3SJCgwh7fNVqJgSt74K38=`

	var key SecretBoxKey
	err := (&key).UnmarshalText([]byte(s))
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := s, key.String(); exp != got {
		t.Fatalf("expected %q, got %q", exp, got)
	}
}
