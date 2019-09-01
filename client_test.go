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
PrivateKey = "KSfVVKjHAlnMWmh5Xr7fCszINOjRGkGfe3Ykx3atTG97+LLeSFWhcmTUrW+20GKCvGwDQVJWqkmX4+sqiXO5ww=="
PublicKey = "e/iy3khVoXJk1K1vttBigrxsA0FSVqpJl+PrKolzucM="
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
