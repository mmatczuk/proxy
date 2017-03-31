package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/mmatczuk/proxy"
)

func main() {
	// http address
	var httpAddr string
	flag.StringVar(&httpAddr, "http", ":80", "HTTP bind address")

	flag.Parse()

	// remote addresses
	addrs := flag.Args()
	if len(addrs) == 0 {
		fmt.Fprintln(os.Stderr, "provide list of servers")
		os.Exit(1)
	}

	logger := logger()
	client := proxy.NewRemoteClient()

	var server http.Handler
	server = proxy.NewServer(proxy.NewService(client, addrs, logger))
	server = proxy.LoggingMiddleware{server, logger}

	logger.Log(
		"msg", "start",
		"addr", httpAddr,
	)

	err := http.ListenAndServe(httpAddr, server)
	if err != nil {
		logger.Log(
			"msg", "could not start",
			"addr", httpAddr,
			"err", err,
		)
	}
}

func logger() log.Logger {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	return logger
}
