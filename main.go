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

func parseFlags(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		if err != flag.ErrHelp {
			fs.Usage()
		}
		os.Exit(1)
	}
}

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage: apero [option] <command> [option]\n\n")
		fmt.Printf("Available commands are:\n")
		fmt.Printf("    serve %50s\n", "serve the API endpoints")
		fmt.Printf("    genconfig %50s\n", "generate the different configurations")
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
		flConfig := fs.String("config", "~/.apero.toml", "Configuration file to use")
		parseFlags(fs, args)

		var conf clientConfig
		if _, err := toml.DecodeFile(*flConfig, &conf); err != nil {
			log.Fatal(err)
		}
		if err := conf.Validate(); err != nil {
			log.Fatal(err)
		}

	case "serve":
		fs := flag.NewFlagSet("serve", flag.ContinueOnError)
		flConfig := fs.String("config", "~/.apero.toml", "Configuration file to use")
		parseFlags(fs, args)

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

	case "genconfig":
		fs := flag.NewFlagSet("genconfig", flag.ContinueOnError)
		flClientConfig := fs.String("client-config", "client.toml", "File path for the client config")
		flServerConfig := fs.String("server-config", "server.toml", "File path for the server config")
		parseFlags(fs, args)

		pub, priv, err := generateKeyPair()
		if err != nil {
			log.Fatal(err)
		}

		//

		clientConf := clientConfig{
			Endpoint:       "http://localhost:7568",
			PSKey:          newSecretBoxKey(),
			EncryptKey:     newSecretBoxKey(),
			SignPublicKey:  pub,
			SignPrivateKey: priv,
		}
		f, err := os.OpenFile(*flClientConfig, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
		if err != nil {
			log.Fatalf("unable to create file %s. err=%v", *flClientConfig, err)
		}
		if err := toml.NewEncoder(f).Encode(clientConf); err != nil {
			log.Fatalf("unable to encode to file. err=%v", err)
		}
		f.Close()

		//

		serverConf := serverConfig{
			ListenAddr:    "http://localhost:7568",
			PSKey:         clientConf.PSKey,
			SignPublicKey: clientConf.SignPublicKey,
		}
		f, err = os.OpenFile(*flServerConfig, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
		if err != nil {
			log.Fatal(err)
		}
		if err := toml.NewEncoder(f).Encode(serverConf); err != nil {
			log.Fatal(err)
		}
		f.Close()

	case "secretbox":
		fs := flag.NewFlagSet("secetbox", flag.ContinueOnError)
		flKey := fs.String("key", "", "The secret key used to seal/open the box")
		flOpen := fs.Bool("open", false, "Open a sealed box")
		flSeal := fs.Bool("seal", false, "Seal a message into a box")
		parseFlags(fs, args)

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

	default:
		fmt.Printf("unknown command %q\n", cmd)
	}

	if err != nil {
		log.Fatal(err)
	}
}
