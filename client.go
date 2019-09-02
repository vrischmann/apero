package main

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"rischmann.fr/apero/internal"
)

type clientConfig struct {
	Endpoint       string
	PSKey          internal.SecretBoxKey
	EncryptKey     internal.SecretBoxKey
	SignPublicKey  internal.PublicKey
	SignPrivateKey internal.PrivateKey
}

func (c clientConfig) Validate() error {
	if _, err := url.Parse(c.Endpoint); err != nil {
		return err
	}
	// TODO(vincent): encrypt key
	return nil
}

type client struct {
	conf       clientConfig
	httpClient http.Client
}

func newClient(conf clientConfig) *client {
	return &client{
		conf:       conf,
		httpClient: http.Client{},
	}
}

func (c *client) makeURL(path string) string {
	return c.conf.Endpoint + path
}

func (c *client) doCopy(req copyRequest) error {
	panic("not implemented")
}

func maybeReadHTTPResponseBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	data, _ := readHTTPResponseBody(resp)
	return string(data)
}

func readHTTPResponseBody(resp *http.Response) ([]byte, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return data, nil
}
