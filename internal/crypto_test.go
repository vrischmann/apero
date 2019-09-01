package internal

import "testing"

func TestPublicKeyUnmarshalText(t *testing.T) {
	const s = `83DR94DNz/oSND7xrZyRF2C9a18jpyq5tFZ22WUfFuk=`

	var key PublicKey
	err := (&key).UnmarshalText(s)
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := s, key.String(); exp != got {
		t.Fatalf("expected %q, got %q", exp, got)
	}
}

func TestSharedKeyUnmarshalText(t *testing.T) {
	const s = `WYBwj9jL9VxlaLlbpMPEMU3SJCgwh7fNVqJgSt74K38=`

	var key SharedKey
	err := (&key).UnmarshalText(s)
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := s, key.String(); exp != got {
		t.Fatalf("expected %q, got %q", exp, got)
	}
}
