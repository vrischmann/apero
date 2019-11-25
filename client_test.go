package main

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestClientConfigUnmarshalText(t *testing.T) {
	const data = `
Endpoint = "http://localhost:7568"
PSKey = "PWiO5P760SBg4BH1P+GN9lOPiEnyLAtQAgNYKWIiKsU="
EncryptKey = "IOC0/OivsG4XkFtUBnmVP9Gb9EXHOOj2SMyB2g76bnM="
SignPublicKey = "bW7CIJBdDTHh5bL5UK7L+AXirOLHQMHSQUi206EoMCI="
SignPrivateKey = "fR2YhWSJFev9eLUXpQRYyr9CKL7F6Nsq6+JuF84Vz0w="
`

	var conf clientConfig
	md, err := toml.Decode(data, &conf)
	require.NoError(t, err)

	require.Empty(t, md.Undecoded())
	require.NoError(t, conf.Validate())
}
