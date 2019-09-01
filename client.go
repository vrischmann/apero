package main

import (
	"errors"
	"net/url"

	"golang.org/x/crypto/ed25519"
)

type clientConfig struct {
	Endpoint   string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

func (c clientConfig) Validate() error {
	if _, err := url.Parse(c.Endpoint); err != nil {
		return err
	}
	if len(c.PublicKey) != ed25519.PublicKeySize {
		return errors.New("public key size is invalid")
	}
	if len(c.PrivateKey) != ed25519.PrivateKeySize {
		return errors.New("private key size is invalid")
	}
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
