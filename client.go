package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

func (c *client) doCopy(req copyRequest) ([]byte, error) {
	return c.doRequest(req, http.MethodPost, http.StatusAccepted, "/api/v1/copy")
}
func (c *client) doMove(req moveRequest) ([]byte, error) {
	return c.doRequest(req, http.MethodDelete, http.StatusOK, "/api/v1/move")
}
func (c *client) doPaste(req pasteRequest) ([]byte, error) {
	return c.doRequest(req, http.MethodPost, http.StatusOK, "/api/v1/paste")
}
func (c *client) doList(req listRequest) ([]byte, error) {
	return c.doRequest(req, http.MethodPost, http.StatusOK, "/api/v1/list")
}

func (c *client) doRequest(req interface{}, method string, expCode int, path string) ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	ciphertext := secretBoxSeal(data, c.conf.PSKey)

	//

	hreq, err := http.NewRequest(method, c.makeURL(path), bytes.NewReader(ciphertext))
	if err != nil {
		return nil, err
	}
	hreq.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.httpClient.Do(hreq)
	if err != nil {
		return nil, fmt.Errorf("unable to copy to staging server. body=%q err: %v", maybeReadHTTPResponseBody(resp), err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != expCode {
		return nil, fmt.Errorf("invalid status code %s. body=%q", resp.Status, maybeReadHTTPResponseBody(resp))
	}

	//

	body, err := readHTTPResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body. err: %v", err)
	}

	body, opened := secretBoxOpen(body, c.conf.PSKey)
	if !opened {
		return nil, fmt.Errorf("unable to open response box")
	}

	return body, nil
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
