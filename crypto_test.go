package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestKeyPairString(t *testing.T) {
	pub, priv, err := generateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("public", func(t *testing.T) {
		var k publicKey
		err := (&k).UnmarshalText([]byte(pub.String()))
		if err != nil {
			t.Fatal(err)
		}

		if exp, got := pub, k; !bytes.Equal(exp, got) {
			t.Fatalf("expected %q but got %q", exp, got)
		}
	})

	t.Run("private", func(t *testing.T) {
		var k privateKey
		err := (&k).UnmarshalText([]byte(priv.String()))
		if err != nil {
			t.Fatal(err)
		}

		if exp, got := priv, k; !bytes.Equal(exp, got) {
			t.Fatalf("expected %q but got %q", exp, got)
		}
	})
}

func TestPublicKeyUnmarshalJSON(t *testing.T) {
	pub, _, err := generateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	var obj struct {
		Key publicKey
	}
	obj.Key = pub

	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}

	var obj2 struct {
		Key publicKey
	}
	err = json.Unmarshal(data, &obj2)
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := pub, obj2.Key; !bytes.Equal(exp, got) {
		t.Fatalf("expected %q but got %q", exp, got)
	}
}

func TestKeyPairUnmarshalText(t *testing.T) {
	pub, priv, err := generateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("public", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			s := pub.String()

			var key publicKey
			err := (&key).UnmarshalText([]byte(s))
			if err != nil {
				t.Fatal(err)
			}

			if exp, got := pub, key; !bytes.Equal(exp, got) {
				t.Fatalf("expected %q, got %q", exp, got)
			}
		})

		t.Run("toml", func(t *testing.T) {
			var obj struct {
				Key publicKey
			}

			md, err := toml.Decode(`Key = "`+pub.String()+`"`, &obj)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(md.Undecoded()); got > 0 {
				t.Fatal("expected no undecoded keys")
			}
			if exp, got := pub, obj.Key; !bytes.Equal(exp, got) {
				t.Fatalf("expected %q, got %q", exp, got)
			}
		})
	})

	t.Run("private", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			s := priv.String()

			var key privateKey
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
				Key privateKey
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

	var key secretBoxKey
	err := (&key).UnmarshalText([]byte(s))
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := s, key.String(); exp != got {
		t.Fatalf("expected %q, got %q", exp, got)
	}
}

func TestSecretBox(t *testing.T) {
	k := newSecretBoxKey()

	data := []byte("foobar")

	box := secretBoxSeal(data, k)
	decrypted, ok := secretBoxOpen(box, k)
	if !ok {
		t.Fatal("expected to open the box")
	}
	if exp, got := data, decrypted; !bytes.Equal(exp, got) {
		t.Fatalf("expected %q but bot %q", string(exp), string(got))
	}
}
