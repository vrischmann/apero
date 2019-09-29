package main

import (
	"encoding/json"
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

func fatal(args ...interface{}) {
	fmt.Println(args...)
	os.Exit(1)
}
func fatalf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}

func readStdin() string {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fatal(err)
	}

	return string(data)
}

var (
	globalFlags  = flag.NewFlagSet("apero", flag.ExitOnError)
	globalConfig = globalFlags.String("config", os.Getenv("HOME")+"/.apero.toml", "Configuration file to use")

	copyFlags  = flag.NewFlagSet("copy", flag.ExitOnError)
	moveFlags  = flag.NewFlagSet("move", flag.ExitOnError)
	pasteFlags = flag.NewFlagSet("paste", flag.ExitOnError)
	listFlags  = flag.NewFlagSet("list", flag.ExitOnError)
	serveFlags = flag.NewFlagSet("serve", flag.ExitOnError)

	genconfigFlags        = flag.NewFlagSet("genconfig", flag.ExitOnError)
	genconfigClientConfig = genconfigFlags.String("client-config", "./client.toml", "File path for the client config")
	genconfigServerConfig = genconfigFlags.String("server-config", "./server.toml", "File path for the server config")
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

	body, err := client.doCopy(req)
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

	//

	client := newClient(conf)

	var (
		body []byte
		err  error
	)

	switch action {
	case "/move":
		body, err = client.doMove(moveRequest{
			ID:        id,
			Signature: sign(conf.SignPrivateKey, id[:]),
		})
	case "/paste":
		body, err = client.doPaste(pasteRequest{
			ID:        id,
			Signature: sign(conf.SignPrivateKey, id[:]),
		})
	}
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return errors.New("nothing in the staging server")
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

func runList(args []string) error {
	var conf clientConfig
	if _, err := toml.DecodeFile(*globalConfig, &conf); err != nil {
		return err
	}
	if err := conf.Validate(); err != nil {
		return err
	}

	//

	client := newClient(conf)

	body, err := client.doList(listRequest{
		Signature: sign(conf.SignPrivateKey, []byte("L")),
	})
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("nothing in the staging server")
	}

	//

	var resp listResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("unable to unmarshal response")
	}

	if len(resp.Entries) == 0 {
		fmt.Println("no entries")
		return nil
	}

	fmt.Printf("entries:\n")
	for _, entry := range resp.Entries {
		fmt.Printf("%s (time: %s)\n", entry, ulid.Time(entry.Time()).UTC().Format(time.RFC3339))
	}

	return nil
}

func runServe(args []string) error {
	var conf serverConfig
	if _, err := toml.DecodeFile(*globalConfig, &conf); err != nil {
		fatal(err)
	}
	if err := conf.Validate(); err != nil {
		fatal(err)
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
		fatal(err)
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
		FlagSet:   pasteFlags,
		ShortHelp: "paste an entry from the staging server to here",
		LongHelp: `Paste an entry from the staging server to here.

Without an argument it pastes the oldest entry.
With an argument it pastes the specific entry if it exists.`,
		Exec: runPaste,
	}

	listCommand := &ffcli.Command{
		Name:      "list",
		Usage:     "apero list",
		FlagSet:   listFlags,
		ShortHelp: "list all entries in the staging server",
		Exec:      runList,
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
		Subcommands: []*ffcli.Command{copyCommand, moveCommand, pasteCommand, listCommand, serveCommand, genconfigCommand},
		Exec: func(args []string) error {
			return errors.New("specify a subcommand")
		},
	}

	if err := root.Run(os.Args[1:]); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
