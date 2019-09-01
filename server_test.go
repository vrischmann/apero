package main

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/BurntSushi/toml"
	"rischmann.fr/apero/internal"
)

func TestServerConfig(t *testing.T) {
	const data = `
ListenAddr = "localhost:7568"
PSKey = "vfHdOcFfBYP2xvuIJuk+JSBB1o9uCdbOMG7imn0riZk="
`

	var conf serverConfig
	md, err := toml.Decode(data, &conf)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(md.Undecoded()); got > 0 {
		t.Fatal("expected no undecoded keys")
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
	var conf serverConfig
	conf.PSKey = internal.NewSecretBoxKey()

	server := newServer(conf)

	httpServer := httptest.NewServer(server)
	defer httpServer.Close()

	//

	var clientConf clientConfig
	clientConf.Endpoint = httpServer.URL
	clientConf.PSKey = conf.PSKey
	clientConf.EncryptKey = internal.NewSecretBoxKey()
	clientConf.PublicKey, clientConf.PrivateKey = mustKeyPair(t)

	client := newClient(clientConf)

	t.Run("register", func(t *testing.T) {
		req := registerRequest{
			DeviceID:  internal.NewDeviceID(),
			PublicKey: clientConf.PublicKey,
		}

		err := client.doRegister(req)
		if err != nil {
			t.Fatal(err)
		}

		// verify the store

		pk, err := server.store.LookupPublicKey(req.DeviceID)
		if err != nil {
			t.Fatal(err)
		}

		if exp, got := clientConf.PublicKey, pk; !bytes.Equal(exp, got) {
			t.Fatalf("expected %q but got %q", exp, got)
		}
	})

	t.Run("copy", func(t *testing.T) {
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
