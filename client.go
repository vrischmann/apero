package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"rischmann.fr/apero/internal"
)

type clientConfig struct {
	Endpoint   string
	PSKey      internal.SecretBoxKey
	EncryptKey internal.SecretBoxKey
	PublicKey  internal.PublicKey
	PrivateKey internal.PrivateKey
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

func (c *client) doRegister(req registerRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return err
	}

	data := internal.SecretBoxSeal(buf.Bytes(), c.conf.PSKey)

	u := c.makeURL("/register")
	hreq, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return err
	}

	//

	resp, err := c.httpClient.Do(hreq)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		body := maybeReadHTTPResponseBody(resp)
		return fmt.Errorf("invalid response. status: %v body: %q", resp.Status, body)
	}

	body, err := readHTTPResponseBody(resp)
	if err != nil {
		return err
	}

	log.Printf("body: %q", string(body))

	return nil
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
