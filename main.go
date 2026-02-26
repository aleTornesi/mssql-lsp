package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/urfave/cli/v2"

	"github.com/atornesi/tsql-ls/internal/config"
	"github.com/atornesi/tsql-ls/internal/handler"
)

const name = "tsql-ls"

const version = "0.1.0"

var revision = "HEAD"

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func realMain() error {
	app := &cli.App{
		Name:    name,
		Version: fmt.Sprintf("Version:%s, Revision:%s\n", version, revision),
		Usage:   "A T-SQL Language Server Protocol implementation.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log",
				Aliases: []string{"l"},
				Usage:   "Also log to this file. (in addition to stderr)",
			},
			&cli.BoolFlag{
				Name:    "trace",
				Aliases: []string{"t"},
				Usage:   "Print all requests and responses.",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to config file.",
			},
		},
		Action: func(c *cli.Context) error {
			return serve(c)
		},
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Print version.",
	}
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "Print help.",
	}

	return app.Run(os.Args)
}

func serve(c *cli.Context) error {
	logfile := c.String("log")
	trace := c.Bool("trace")
	configFile := c.String("config")

	var logWriter io.Writer
	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		logWriter = io.MultiWriter(os.Stderr, f)
	} else {
		logWriter = io.MultiWriter(os.Stderr)
	}
	log.SetOutput(logWriter)

	server := handler.NewServer()

	if configFile != "" {
		cfg, err := config.GetConfig(configFile)
		if err != nil {
			log.Printf("config load %q: %v", configFile, err)
		} else {
			server.SpecificFileCfg = cfg
		}
	} else {
		cfg, err := config.GetDefaultConfig()
		if err != nil && !errors.Is(err, config.ErrNotFoundConfig) {
			log.Printf("default config load: %v", err)
		} else if err == nil {
			server.DefaultFileCfg = cfg
		}
	}
	defer func() {
		if err := server.Stop(); err != nil {
			log.Println(err)
		}
	}()
	h := jsonrpc2.HandlerWithError(server.Handle)

	var connOpt []jsonrpc2.ConnOpt
	if trace {
		connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(logWriter, "", 0)))
	}

	log.Println("tsql-ls: reading on stdin, writing on stdout")
	<-jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		h,
		connOpt...,
	).DisconnectNotify()
	log.Println("tsql-ls: connections closed")

	return nil
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
