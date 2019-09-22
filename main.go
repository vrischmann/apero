package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/oklog/ulid/v2"
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
	globalFlags  = flag.NewFlagSet("apero", flag.ExitOnError)
	globalConfig = globalFlags.String("config", os.Getenv("HOME")+"/.apero.toml", "Configuration file to use")

	copyFlags  = flag.NewFlagSet("copy", flag.ExitOnError)
	moveFlags  = flag.NewFlagSet("move", flag.ExitOnError)
	serveFlags = flag.NewFlagSet("serve", flag.ExitOnError)

	genconfigFlags        = flag.NewFlagSet("genconfig", flag.ExitOnError)
	genconfigClientConfig = genconfigFlags.String("client-config", "client.toml", "File path for the client config")
	genconfigServerConfig = genconfigFlags.String("server-config", "client.toml", "File path for the server config")
)

func runCopy(args []string) error {
	var conf clientConfig
	if _, err := toml.DecodeFile(*globalConfig, &conf); err != nil {
		return err
	}
	if err := conf.Validate(); err != nil {
		return err
	}

	if len(args) < 1 {
		return errors.New("need at least one path to copy")
	}

	var (
		data []byte
		err  error
	)
	switch {
	case args[0] == "-":
		data, err = ioutil.ReadAll(os.Stdin)
	default:
		data, err = ioutil.ReadFile(args[0])
	}
	if err != nil {
		return err
	}

	//

	ciphertext := secretBoxSeal(data, conf.EncryptKey)
	signature := sign(conf.SignPrivateKey, ciphertext)

	//

	client := newClient(conf)

	req := copyRequest{
		Signature: signature,
		Content:   ciphertext,
	}

	body, err := client.doRequest(req, "/copy")
	if err != nil {
		return err
	}

	var id ulid.ULID
	copy(id[:], body)

	fmt.Printf("id: %s\n", id)

	return nil
}

func doRunMoveOrPaste(args []string, action string) error {
	var conf clientConfig
	if _, err := toml.DecodeFile(*globalConfig, &conf); err != nil {
		return err
	}
	if err := conf.Validate(); err != nil {
		return err
	}

	var id ulid.ULID
	if len(args) > 0 {
		var err error
		id, err = ulid.Parse(args[0])
		if err != nil {
			return err
		}
	}

	req := moveOrPasteRequest{
		ID:        id,
		Signature: sign(conf.SignPrivateKey, id[:]),
	}

	//

	client := newClient(conf)

	body, err := client.doRequest(req, action)
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return nil
	}

	plaintext, ok := secretBoxOpen(body, conf.EncryptKey)
	if !ok {
		return fmt.Errorf("unable to decipher content")
	}

	os.Stdout.Write(plaintext)

	return nil
}

func runMove(args []string) error {
	return doRunMoveOrPaste(args, "/move")
}

func runPaste(args []string) error {
	return doRunMoveOrPaste(args, "/paste")
}

func runServe(args []string) error {
	var conf serverConfig
	if _, err := toml.DecodeFile(*globalConfig, &conf); err != nil {
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
		ShortHelp: "copy a file to the staging server",
		LongHelp: `Copy a file to the staging server.

If the path given is - it will read from stdin.

This command will print an ID which can be further used with move and paste.
`,
		Exec: runCopy,
	}

	moveCommand := &ffcli.Command{
		Name:      "move",
		Usage:     "apero move [entry id]",
		FlagSet:   moveFlags,
		ShortHelp: "move an entry from the staging server to here",
		LongHelp: `Move an entry from the staging server to here.

Without an argument it moves the oldest entry.
With an argument it moves the specific entry if it exists.`,
		Exec: runMove,
	}

	pasteCommand := &ffcli.Command{
		Name:      "paste",
		Usage:     "apero paste [entry id]",
		FlagSet:   moveFlags,
		ShortHelp: "paste an entry from the staging server to here",
		LongHelp: `Paste an entry from the staging server to here.

Without an argument it pastes the oldest entry.
With an argument it pastes the specific entry if it exists.`,
		Exec: runMove,
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
		Subcommands: []*ffcli.Command{copyCommand, moveCommand, pasteCommand, serveCommand, genconfigCommand},
		Exec: func(args []string) error {
			return errors.New("specify a subcommand")
		},
	}

	if err := root.Run(os.Args[1:]); err != nil {
		log.Fatalf("error: %v", err)
	}
}
