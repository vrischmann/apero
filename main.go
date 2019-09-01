package main

import (
	crypto_rand "crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/vrischmann/hutil/v2"
)

func getNonce() [24]byte {
	var nonce [24]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	return nonce
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
		fmt.Printf("    serve %50s\n", "serve the API endpoints")
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
		fs := flag.NewFlagSet("serve", flag.ContinueOnError)
		flAddr := fs.String("addr", ":7568", "The listen address")
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

		server := newServer(conf)

		var chain hutil.Chain
		chain.Use(hutil.NewLoggingMiddleware(func(req *http.Request, statusCode int, responseSize int, elapsed time.Duration) {
			log.Printf("[%3d] %s %d %s", statusCode, req.URL.Path, responseSize, elapsed)
		}))

		err = http.ListenAndServe(*flAddr, chain.Handler(server))

	case "genkeys":
		fs := flag.NewFlagSet("genkeys", flag.ContinueOnError)
		flServer := fs.Bool("server", false, "Generate the keys for a server")
		flDeviceID := fs.Bool("device-id", false, "Generate a device ID. Useful only for debugging")
		if err := fs.Parse(args); err != nil {
			fs.Usage()
			os.Exit(1)
		}

		switch {
		case *flDeviceID:
			var id deviceID
			if _, err := crypto_rand.Read(id[:]); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", id.String())

		case *flServer:
			var key sharedKey
			if _, err := crypto_rand.Read(key[:]); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Key = %q\n", key)

		default:
			fs.Usage()
			os.Exit(1)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}
