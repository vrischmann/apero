package main

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/nacl/box"
)

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

type serverConfig struct {
	RecipientPrivateKey string
	SenderPublicKey     string
}

func (c serverConfig) Validate() error {
	if c.RecipientPrivateKey == "" {
		return errors.New("recipient private key is empty")
	}
	if c.SenderPublicKey == "" {
		return errors.New("sender public key is empty")
	}
	return nil
}

func runGenKeys() error {
	senderPubKey, senderPrivKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		return err
	}

	recipientPubKey, recipientPrivKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		return err
	}

	{
		conf := clientConfig{
			RecipientPublicKey: hex.EncodeToString((*recipientPubKey)[:]),
			SenderPrivateKey:   hex.EncodeToString((*senderPrivKey)[:]),
		}

		f, err := os.Create("client.toml")
		if err != nil {
			return err
		}

		encoder := toml.NewEncoder(f)
		if err := encoder.Encode(conf); err != nil {
			return err
		}
	}

	{
		conf := serverConfig{
			RecipientPrivateKey: hex.EncodeToString((*recipientPrivKey)[:]),
			SenderPublicKey:     hex.EncodeToString((*senderPubKey)[:]),
		}

		f, err := os.Create("server.toml")
		if err != nil {
			return err
		}

		encoder := toml.NewEncoder(f)
		if err := encoder.Encode(conf); err != nil {
			return err
		}
	}

	return nil
}

func getNonce() [24]byte {
	var nonce [24]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	return nonce
}

func parseKey(s string) *[32]byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	var key [32]byte
	copy(key[:], data)

	return &key
}

func readStdin() string {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}

func main() {
	var (
		flConfig = flag.String("config", "~/.apero.toml", "Configuration file to use")
	)

	flag.Usage = func() {
		fmt.Printf("Usage: apero [option] <command> [option]\n\n")
		fmt.Printf("Available commands are:\n")
		fmt.Printf("    genkeys %50s\n", "generate public and private keys")
		fmt.Printf("    encrypt %50s\n", "encrypt a message")
		fmt.Printf("    decrypt %50s\n", "decrypt a message")
		fmt.Printf("\nOptions are:\n")
		flag.PrintDefaults()
		fmt.Println()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	//

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	var err error
	switch cmd {
	case "serve":
		err = nil

	case "genkeys":
		err = runGenKeys()

	case "encrypt":
		fs := flag.NewFlagSet("encrypt", flag.ContinueOnError)
		fs.Usage = func() {
			fmt.Printf("Usage: echo foobar | apero encrypt\n\n")
			fs.PrintDefaults()
			fmt.Println()
		}
		if err := fs.Parse(args); err != nil {
			fs.Usage()
			os.Exit(1)
		}

		//

		var conf clientConfig
		if _, err := toml.DecodeFile(*flConfig, &conf); err != nil {
			log.Fatal(err)
		}
		if err := conf.Validate(); err != nil {
			log.Fatal(err)
		}

		//

		var (
			msg             = readStdin()
			recipientPubKey = parseKey(conf.RecipientPublicKey)
			senderPrivKey   = parseKey(conf.SenderPrivateKey)
			nonce           = getNonce()
		)

		//

		ciphertext := box.Seal(nonce[:], []byte(msg), &nonce, recipientPubKey, senderPrivKey)

		fmt.Printf("%x", ciphertext)

	case "decrypt":
		fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
		fs.Usage = func() {
			fmt.Printf("Usage: apero decrypt --recipient-private-key <key> --sender-public-key <key>\n\n")
			fs.PrintDefaults()
			fmt.Println()
		}

		if err := fs.Parse(args); err != nil {
			fs.Usage()
			os.Exit(1)
		}

		//

		var conf serverConfig
		if _, err := toml.DecodeFile(*flConfig, &conf); err != nil {
			log.Fatal(err)
		}
		if err := conf.Validate(); err != nil {
			log.Fatal(err)
		}

		//

		ciphertext, err := hex.DecodeString(readStdin())
		if err != nil {
			log.Fatal(err)
		}

		//

		var (
			decryptNonce [24]byte

			recipientPrivKey = parseKey(conf.RecipientPrivateKey)
			senderPubKey     = parseKey(conf.SenderPublicKey)
		)

		// Nonce is preprended to ciphertext
		copy(decryptNonce[:], ciphertext[:24])
		ciphertext = ciphertext[24:]

		//

		msg, ok := box.Open(nil, ciphertext, &decryptNonce, senderPubKey, recipientPrivKey)
		if !ok {
			log.Fatal("decryption error")
		}

		fmt.Printf("%s\n", string(msg))
	}

	if err != nil {
		log.Fatal(err)
	}
}
