package main

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestClientConfigUnmarshalText(t *testing.T) {
	const data = `
Endpoint = "http://localhost:7568"
PSKey = "vfHdOcFfBYP2xvuIJuk+JSBB1o9uCdbOMG7imn0riZk="
EncryptKey = "jHBWWFhacJjEDo60pqvhVHE4rGVK4pOvxlhC8EoRNPw="
SignPrivateKey = "KSfVVKjHAlnMWmh5Xr7fCszINOjRGkGfe3Ykx3atTG97+LLeSFWhcmTUrW+20GKCvGwDQVJWqkmX4+sqiXO5ww=="
SignPublicKey = "e/iy3khVoXJk1K1vttBigrxsA0FSVqpJl+PrKolzucM="
`

	var conf clientConfig
	md, err := toml.Decode(data, &conf)
	require.NoError(t, err)

	require.Empty(t, md.Undecoded())
	require.NoError(t, conf.Validate())
}
