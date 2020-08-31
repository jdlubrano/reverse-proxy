package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jdlubrano/reverse-proxy/internal/logger"
	"github.com/jdlubrano/reverse-proxy/internal/proxy"
	"github.com/jdlubrano/reverse-proxy/internal/routes"
	"github.com/peterbourgon/ff/v3"
)

func main() {
	fs := flag.NewFlagSet("reverse-proxy", flag.ExitOnError)
	var (
		port       = fs.String("port", "8080", "port we will start sf-rpc serving on (also via PORT)")
		routesFile = fs.String("routes-file", "", "start sf-rpc serving on (also via ROUTES_FILE)")
	)

	err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("REVERSE_PROXY"))
	if err != nil {
		fmt.Println("failed to parse flags:", err)
		os.Exit(1)
	}

	if *routesFile == "" {
		fmt.Println("You must supply a --routes-file command line argument or set REVERSE_PROXY_ROUTES_FILE in your env")
		os.Exit(1)
	}

	logger := logger.NewStdoutLogger()

	routesConfig, err := routes.NewRoutesConfigFromYaml(*routesFile)

	if err != nil {
		logger.Errorf("failed to load routes from %s: %s", *routesFile, err)
	}

	logger.Infof("Loaded routes from %s", *routesFile)

	ctx, cancelCtx := context.WithCancel(context.Background())
	go CancelOnSignal(cancelCtx, logger)

	serverPort, err := strconv.Atoi(*port)

	if err != nil {
		logger.Errorf("Invalid PORT: %s", err)
	}

	proxy := proxy.NewProxy(logger, routesConfig, serverPort)

	logger.Infof("Starting server on port %d", serverPort)

	go func() {
		proxy.Start()
	}()

	logger.Infof("listening at: %s", *port)

	<-ctx.Done()

	err = proxy.Stop()

	if err != nil {
		logger.Errorf("stopping server: %s", err.Error())
	}
}

func CancelOnSignal(ctxCancel context.CancelFunc, logger logger.Logger) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)
	sig := <-s //wait for Signal
	logger.Infof("received signal %s", sig)
	ctxCancel()
}
