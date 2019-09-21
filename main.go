package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/vrischmann/hutil/v2"
)

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
		fmt.Printf("    serve %50s\n", "serve the API endpoints")
		fmt.Printf("    genkeys %50s\n", "generate public and private keys")
		fmt.Printf("    secretbox %50s\n", "seal and open secret boxes")
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
	case "copy":
		fs := flag.NewFlagSet("copy", flag.ContinueOnError)
		if err := fs.Parse(args); err != nil {
			fs.Usage()
			os.Exit(1)
		}

		var conf clientConfig
		if _, err := toml.DecodeFile(*flConfig, &conf); err != nil {
			log.Fatal(err)
		}
		if err := conf.Validate(); err != nil {
			log.Fatal(err)
		}

	case "serve":
		fs := flag.NewFlagSet("serve", flag.ContinueOnError)
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

		// TODO(vincent): configure this based on conf
		server := newServer(conf, newMemStore())

		var chain hutil.Chain
		chain.Use(hutil.NewLoggingMiddleware(func(req *http.Request, statusCode int, responseSize int, elapsed time.Duration) {
			log.Printf("[%3d] %s %d %s", statusCode, req.URL.Path, responseSize, elapsed)
		}))

		err = http.ListenAndServe(conf.ListenAddr, chain.Handler(server))

	case "genkeys":
		fs := flag.NewFlagSet("genkeys", flag.ContinueOnError)
		flSecretBox := fs.Bool("secretbox", false, "Generate a key for a secretbox")
		flKeyPair := fs.Bool("keypair", false, "Generate a ed25519 key pair. Useful only for debugging")
		if err := fs.Parse(args); err != nil {
			fs.Usage()
			os.Exit(1)
		}

		switch {
		case *flSecretBox:
			key := newSecretBoxKey()

			fmt.Printf("Key = %q\n", key)

		case *flKeyPair:
			publicKey, privateKey, err := generateKeyPair()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("PrivateKey = %q\n", privateKey)
			fmt.Printf("PublicKey = %q\n", publicKey)

		default:
			fs.Usage()
			os.Exit(1)
		}

	case "secretbox":
		fs := flag.NewFlagSet("secetbox", flag.ContinueOnError)
		flKey := fs.String("key", "", "The secret key used to seal/open the box")
		flOpen := fs.Bool("open", false, "Open a sealed box")
		flSeal := fs.Bool("seal", false, "Seal a message into a box")
		if err := fs.Parse(args); err != nil {
			fs.Usage()
			os.Exit(1)
		}

		keyData, err := hex.DecodeString(*flKey)
		if err != nil {
			log.Fatal(err)
		}
		if len(keyData) != secretBoxKeySize {
			log.Fatalf("invalid key size %d", len(keyData))
		}

		var key secretBoxKey
		copy(key[:], keyData)

		switch {
		case *flOpen:
			data, err := hex.DecodeString(fs.Arg(0))
			if err != nil {
				log.Fatal(err)
			}

			message, opened := secretBoxOpen(data, key)
			if !opened {
				log.Fatal("unable to open box")
			}

			fmt.Printf("%s\n", string(message))

		case *flSeal:
			message := []byte(fs.Arg(0))

			ciphertext := secretBoxSeal(message, key)

			fmt.Printf("%02x\n", ciphertext)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}
