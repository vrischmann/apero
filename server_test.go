package main

import (
	"net/http/httptest"
	"testing"

	"github.com/BurntSushi/toml"
	"rischmann.fr/apero/internal"
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

func TestShiftPath(t *testing.T) {
	testCases := []struct {
		input string
		head  string
		tail  string
	}{
		{"/copy", "copy", "/"},
		{"/", "", "/"},
		{"/api/v1/hello", "api", "/v1/hello"},
		{".", "", "/"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			head, tail := shiftPath(tc.input)
			if got, exp := head, tc.head; got != exp {
				t.Fatalf("expected %q but got %q", exp, got)
			}
			if got, exp := tail, tc.tail; got != exp {
				t.Fatalf("expected %q but got %q", exp, got)
			}
		})
	}
}

func TestServerClient(t *testing.T) {
	publicKey, privateKey := mustKeyPair(t)

	var conf serverConfig
	conf.PSKey = internal.NewSecretBoxKey()
	conf.SignPublicKey = publicKey

	server := newServer(conf)

	httpServer := httptest.NewServer(server)
	defer httpServer.Close()

	//

	var clientConf clientConfig
	clientConf.Endpoint = httpServer.URL
	clientConf.PSKey = conf.PSKey
	clientConf.EncryptKey = internal.NewSecretBoxKey()
	clientConf.SignPublicKey = publicKey
	clientConf.SignPrivateKey = privateKey

	client := newClient(clientConf)
	_ = client

	t.Run("copy", func(t *testing.T) {
		content := []byte("hello")
		signature := internal.Sign(clientConf.SignPrivateKey, content)

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

func mustKeyPair(t *testing.T) (internal.PublicKey, internal.PrivateKey) {
	pub, priv, err := internal.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	return pub, priv
}
