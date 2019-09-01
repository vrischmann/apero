package main

import (
	"net/url"

	"golang.org/x/crypto/ed25519"
	"rischmann.fr/apero/internal"
)

type clientConfig struct {
	Endpoint   string
	Key        internal.SharedKey
	EncryptKey string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
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
