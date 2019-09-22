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

	t.Run("copy", func(t *testing.T) {
		content := []byte("hello")
		signature := sign(clientConf.SignPrivateKey, content)
		req := copyRequest{Signature: signature, Content: content}

		body, err := client.doRequest(req, "/copy")
		require.NoError(t, err)

		//

		entries, err := server.st.ListAll()
		require.NoError(t, err)
		require.NotEmpty(t, entries)
		expID := entries[0]

		require.Equal(t, expID[:], body[:])

		entry, err := server.st.RemoveFirst()
		require.NoError(t, err)

		require.Equal(t, content, entry)
	})

	t.Run("move-oldest", func(t *testing.T) {
		_, err := server.st.Add([]byte("yoo"))
		require.NoError(t, err)

		//

		var req moveOrPasteRequest
		req.Signature = sign(clientConf.SignPrivateKey, req.ID[:])

		body, err := client.doRequest(req, "/move")
		require.NoError(t, err)
		require.Equal(t, []byte("yoo"), body)

		//

		entries, err := server.st.ListAll()
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run("move-specific", func(t *testing.T) {
		oldestID, err := server.st.Add([]byte("yoo"))
		require.NoError(t, err)
		id, err := server.st.Add([]byte("yezi"))
		require.NoError(t, err)
		require.False(t, isEmptyULID(id))

		//

		req := moveOrPasteRequest{
			ID:        id,
			Signature: sign(clientConf.SignPrivateKey, id[:]),
		}

		body, err := client.doRequest(req, "/move")
		require.NoError(t, err)
		require.Equal(t, []byte("yezi"), body)

		//

		entries, err := server.st.ListAll()
		require.NoError(t, err)
		require.Equal(t, entries[0], oldestID)

		server.st.RemoveFirst() // cleanup for the next test
	})

	t.Run("paste-oldest", func(t *testing.T) {
		id, err := server.st.Add([]byte("yoo"))
		require.NoError(t, err)

		//

		var req moveOrPasteRequest
		req.Signature = sign(clientConf.SignPrivateKey, req.ID[:])

		body, err := client.doRequest(req, "/paste")
		require.NoError(t, err)
		require.Equal(t, []byte("yoo"), body)

		//

		entries, err := server.st.ListAll()
		require.NoError(t, err)
		require.Equal(t, 1, len(entries))
		require.Equal(t, id, entries[0])

		server.st.RemoveFirst() // cleanup for the next test
	})

	t.Run("paste-specific", func(t *testing.T) {
		oldestID, err := server.st.Add([]byte("yoo"))
		require.NoError(t, err)
		id, err := server.st.Add([]byte("yeoa"))
		require.NoError(t, err)

		//

		req := moveOrPasteRequest{
			ID:        id,
			Signature: sign(clientConf.SignPrivateKey, id[:]),
		}

		body, err := client.doRequest(req, "/paste")
		require.NoError(t, err)
		require.Equal(t, []byte("yeoa"), body)

		//

		entries, err := server.st.ListAll()
		require.NoError(t, err)
		require.Equal(t, 2, len(entries))
		require.Equal(t, oldestID, entries[0])
		require.Equal(t, id, entries[1])

	})
}

func mustKeyPair(t *testing.T) (publicKey, privateKey) {
	pub, priv, err := generateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	return pub, priv
}
