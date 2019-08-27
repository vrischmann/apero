package main

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
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

type serverConfig struct {
	RecipientPrivateKey string
	SenderPublicKey     string
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
	flag.Usage = func() {
		fmt.Printf("Usage: apero [option] <command> [option]\n\n")
		fmt.Printf("Available commands are:\n")
		fmt.Printf("    genkeys %50s\n", "generate public and private keys")
		fmt.Printf("    encrypt %50s\n", "encrypt a message")
		fmt.Printf("    decrypt %50s\n", "decrypt a message")
		flag.PrintDefaults()
		fmt.Println()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	var err error
	switch cmd {
	case "genkeys":
		err = runGenKeys()

	case "encrypt":
		fs := flag.NewFlagSet("encrypt", flag.ContinueOnError)
		fs.Usage = func() {
			fmt.Printf("Usage: apero encrypt --recipient-public-key <key> --sender-private-key <key>\n\n")
			fs.PrintDefaults()
			fmt.Println()
		}
		var (
			flRecipientPubKey = fs.String("recipient-public-key", "", "The recipient's public key")
			flSenderPrivKey   = fs.String("sender-private-key", "", "The sender's private key")
		)

		if err := fs.Parse(args); err != nil {
			fs.Usage()
			os.Exit(1)
		}

		//

		switch {
		case *flRecipientPubKey == "":
			fmt.Printf("no recipient public key\n\n")
			fs.Usage()
			os.Exit(1)

		case *flSenderPrivKey == "":
			fmt.Printf("no sender private key\n\n")
			fs.Usage()
			os.Exit(1)
		}

		//

		var (
			msg             = readStdin()
			recipientPubKey = parseKey(*flRecipientPubKey)
			senderPrivKey   = parseKey(*flSenderPrivKey)
			nonce           = getNonce()
		)

		//

		ciphertext := box.Seal(nonce[:], []byte(msg), &nonce, recipientPubKey, senderPrivKey)

		fmt.Printf("%x\n", ciphertext)

	case "decrypt":
		fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
		fs.Usage = func() {
			fmt.Printf("Usage: apero decrypt --recipient-private-key <key> --sender-public-key <key>\n\n")
			fs.PrintDefaults()
			fmt.Println()
		}
		var (
			flRecipientPrivKey = fs.String("recipient-private-key", "", "The recipient's private key")
			flSenderPubKey     = fs.String("sender-public-key", "", "The sender's public key")
		)

		if err := fs.Parse(args); err != nil {
			fs.Usage()
			os.Exit(1)
		}

		//

		switch {
		case *flRecipientPrivKey == "":
			fmt.Printf("no recipient private key\n\n")
			fs.Usage()
			os.Exit(1)

		case *flSenderPubKey == "":
			fmt.Printf("no sender public key\n\n")
			fs.Usage()
			os.Exit(1)
		}

		//

		ciphertext, err := hex.DecodeString(readStdin())
		if err != nil {
			log.Fatal(err)
		}

		//

		var (
			decryptNonce [24]byte

			recipientPrivKey = parseKey(*flRecipientPrivKey)
			senderPubKey     = parseKey(*flSenderPubKey)
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
