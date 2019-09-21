package main

import (
	"net/http/httptest"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestServerConfigUnmarshalText(t *testing.T) {
	const data = `
ListenAddr = "localhost:7568"
PSKey = "vfHdOcFfBYP2xvuIJuk+JSBB1o9uCdbOMG7imn0riZk="
SignPublicKey = "GKlTcESb8Qm8KH+3wWoPWMf7DvVUWYzsKymvUKhhTo8="
`

	var conf serverConfig
	md, err := toml.Decode(data, &conf)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(md.Undecoded()); got > 0 {
		t.Fatal("expected no undecoded keys")
	}
	if err := conf.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestServerClient(t *testing.T) {
	publicKey, privateKey := mustKeyPair(t)

	var conf serverConfig
	conf.PSKey = newSecretBoxKey()
	conf.SignPublicKey = publicKey

	server := newServer(conf)

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
		if err != nil {
			t.Fatal(err)
		}
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
