package main

import (
	"net/url"

	"rischmann.fr/apero/internal"
)

type clientConfig struct {
	Endpoint   string
	Key        internal.SecretBoxKey
	EncryptKey string
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
	conf clientConfig
}

func newClient(conf clientConfig) *client {
	return &client{
		conf: conf,
	}
}

func (c *client) doCopy(req copyRequest) error {
	panic("not implemented")
}

func (c *client) doRegister(req registerRequest) error {
	panic("not implemented")
}
