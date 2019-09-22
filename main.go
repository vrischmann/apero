package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/peterbourgon/ff"
	"github.com/peterbourgon/ff/ffcli"
	"github.com/vrischmann/hutil/v2"
)

func readStdin() string {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}

var (
	globalFlags = flag.NewFlagSet("apero", flag.ExitOnError)

	copyFlags  = flag.NewFlagSet("copy", flag.ExitOnError)
	copyConfig = copyFlags.String("config", os.Getenv("HOME")+"/.apero.toml", "Configuration file to use")

	serveFlags  = flag.NewFlagSet("serve", flag.ExitOnError)
	serveConfig = serveFlags.String("config", os.Getenv("HOME")+"/.apero.toml", "Configuration file to use")

	genconfigFlags        = flag.NewFlagSet("genconfig", flag.ExitOnError)
	genconfigClientConfig = genconfigFlags.String("client-config", "client.toml", "File path for the client config")
	genconfigServerConfig = genconfigFlags.String("server-config", "client.toml", "File path for the server config")
)

func runCopy(args []string) error {
	var conf clientConfig
	if _, err := toml.DecodeFile(*copyConfig, &conf); err != nil {
		return err
	}
	if err := conf.Validate(); err != nil {
		return err
	}

	// TODO(vincent): write me

	return nil
}

func runServe(args []string) error {
	var conf serverConfig
	if _, err := toml.DecodeFile(*serveConfig, &conf); err != nil {
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

	return http.ListenAndServe(conf.ListenAddr, chain.Handler(server))
}

func runGenconfig(args []string) error {
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
	f, err := os.OpenFile(*genconfigClientConfig, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(clientConf); err != nil {
		return err
	}
	f.Close()

	//

	serverConf := serverConfig{
		ListenAddr:    "localhost:7568",
		PSKey:         clientConf.PSKey,
		SignPublicKey: clientConf.SignPublicKey,
	}
	f, err = os.OpenFile(*genconfigServerConfig, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(serverConf); err != nil {
		return err
	}
	f.Close()

	return nil
}

func main() {
	copyCommand := &ffcli.Command{
		Name:      "copy",
		Usage:     "apero copy <file path>",
		FlagSet:   copyFlags,
		ShortHelp: "copy a file to the server",
		LongHelp: `Copy a file to the server.
If the path given is - it will read from stdin.`,
		Exec: runCopy,
	}

	serveCommand := &ffcli.Command{
		Name:      "serve",
		Usage:     "apero serve [flags]",
		FlagSet:   serveFlags,
		ShortHelp: "serve requests to clients",
		Exec:      runServe,
	}

	genconfigCommand := &ffcli.Command{
		Name:      "genconfig",
		Usage:     "apero genconfig [flags]",
		FlagSet:   genconfigFlags,
		ShortHelp: "generate configuration files for client and server",
		LongHelp: `generate configuration files for client and server.
The path can be changed with a flag:

    apero genconfig -client-config=/tmp/client.toml -server-config=/tmp/server.toml`,
		Exec: runGenconfig,
	}

	root := &ffcli.Command{
		Usage:       "apero [global flags] <subcommand> [flags] [args...]",
		FlagSet:     globalFlags,
		Options:     []ff.Option{ff.WithEnvVarPrefix("APERO")},
		LongHelp:    `Run a staging server or communicate with one`,
		Subcommands: []*ffcli.Command{copyCommand, serveCommand, genconfigCommand},
		Exec: func([]string) error {
			return errors.New("specify a subcommand")
		},
	}

	if err := root.Run(os.Args[1:]); err != nil {
		log.Fatalf("error: %v", err)
	}
}
