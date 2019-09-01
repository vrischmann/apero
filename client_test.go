package main

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestClientConfigUnmarshalTOML(t *testing.T) {
	const data = `
Endpoint = "http://localhost:7568"
PSKey = "vfHdOcFfBYP2xvuIJuk+JSBB1o9uCdbOMG7imn0riZk="
EncryptKey = "jHBWWFhacJjEDo60pqvhVHE4rGVK4pOvxlhC8EoRNPw="
PrivateKey = "fQT3akTsjZkfdg53C9DkDboEJzPXGVFBAJ2TKgXoCpU="
PublicKey = "7wxfhseys0/UoZMWrOro/Ss019ym4Bw0Qe5RXL1l+/19BPdqROyNmR92DncL0OQNugQnM9cZUUEAnZMqBegKlQ=="
`

	var conf clientConfig
	md, err := toml.Decode(data, &conf)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(md.Undecoded()); got > 0 {
		t.Fatal("expected no undecoded keys")
	}
}
