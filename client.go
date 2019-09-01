package main

import "errors"

type clientConfig struct {
	RecipientPublicKey string
	SenderPrivateKey   string
}

func (c clientConfig) Validate() error {
	if c.RecipientPublicKey == "" {
		return errors.New("recipient public key is empty")
	}
	if c.SenderPrivateKey == "" {
		return errors.New("sender private key is empty")
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
