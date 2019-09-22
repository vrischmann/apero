package main

import (
	"net/http/httptest"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestServerConfigUnmarshalText(t *testing.T) {
	const data = `
ListenAddr = "localhost:7568"
PSKey = "vfHdOcFfBYP2xvuIJuk+JSBB1o9uCdbOMG7imn0riZk="
SignPublicKey = "GKlTcESb8Qm8KH+3wWoPWMf7DvVUWYzsKymvUKhhTo8="
`

	var conf serverConfig
	md, err := toml.Decode(data, &conf)
	require.NoError(t, err)

	require.Empty(t, md.Undecoded())
	require.NoError(t, conf.Validate())
}

func TestServerClient(t *testing.T) {
	publicKey, privateKey := mustKeyPair(t)

	var conf serverConfig
	conf.PSKey = newSecretBoxKey()
	conf.SignPublicKey = publicKey

	server := newServer(conf, newMemStore())

	httpServer := httptest.NewServer(server)
	defer httpServer.Close()

	//

	var clientConf clientConfig
	clientConf.Endpoint = httpServer.URL
	clientConf.PSKey = conf.PSKey
	clientConf.EncryptKey = newSecretBoxKey()
	clientConf.SignPublicKey = publicKey
	clientConf.SignPrivateKey = privateKey

	client := newClient(clientConf)
	_ = client

	t.Run("copy", func(t *testing.T) {
		content := []byte("hello")
		signature := sign(clientConf.SignPrivateKey, content)

		req := copyRequest{
			Signature: signature,
			Content:   content,
		}

		err := client.doCopy(req)
		require.NoError(t, err)

		entries, err := server.st.ListAll()
		require.NoError(t, err)
		require.NotEmpty(t, entries)

		entry, err := server.st.Pop()
		require.NoError(t, err)

		require.Equal(t, content, entry)
	})

	t.Run("move", func(t *testing.T) {
	})

	t.Run("paste", func(t *testing.T) {
	})
}

func mustKeyPair(t *testing.T) (publicKey, privateKey) {
	pub, priv, err := generateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	return pub, priv
}
