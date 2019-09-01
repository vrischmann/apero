package main

import (
	"testing"

	"github.com/BurntSushi/toml"
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
