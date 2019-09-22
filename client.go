package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/oklog/ulid/v2"
)

type clientConfig struct {
	Endpoint       string
	PSKey          secretBoxKey
	EncryptKey     secretBoxKey
	SignPublicKey  publicKey
	SignPrivateKey privateKey
}

func (c clientConfig) Validate() error {
	if _, err := url.Parse(c.Endpoint); err != nil {
		return err
	}
	if !c.PSKey.IsValid() {
		return fmt.Errorf("ps key is invalid")
	}
	if !c.SignPrivateKey.IsValid() {
		return fmt.Errorf("sign public key is invalid")
	}
	if !c.SignPublicKey.IsValid() {
		return fmt.Errorf("sign public key is invalid")
	}
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
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	ciphertext := secretBoxSeal(data, c.conf.PSKey)

	//

	hreq, err := http.NewRequest(http.MethodPost, c.makeURL("/copy"), bytes.NewReader(ciphertext))
	if err != nil {
		return err
	}
	hreq.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.httpClient.Do(hreq)
	if err != nil {
		return fmt.Errorf("unable to copy to staging server. body=%q err: %v", maybeReadHTTPResponseBody(resp), err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("invalid status code %s. body=%q", resp.Status, maybeReadHTTPResponseBody(resp))
	}

	//

	body, err := readHTTPResponseBody(resp)
	if err != nil {
		return fmt.Errorf("unable to read response body. err: %v", err)
	}

	body, opened := secretBoxOpen(body, c.conf.PSKey)
	if !opened {
		return fmt.Errorf("unable to open response box")
	}

	var id ulid.ULID
	copy(id[:], body)

	fmt.Printf("id: %s\n", id)

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
