package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	zipapi "zipapi/api"
	"zipapi/api/config"
)

var argConfigFile = flag.String(
	"config",
	"./config.toml",
	"path to the configuration file",
)

func main() {
	flag.Parse()

	conf, err := config.FromFile(*argConfigFile)
	if err != nil {
		log.Fatalf("reading config: %s", err)
	}

	api, err := zipapi.NewServer(conf)
	if err != nil {
		log.Fatalf("API server init: %s", err)
	}

	// Setup termination signal listener
	onTerminate(func() {
		if err := api.Shutdown(context.Background()); err != nil {
			log.Fatalf("API server shutdown: %s", err)
		}
	})

	if err := api.Run(); err != nil {
		log.Fatalf("API server: %s", err)
	}
}

func onTerminate(callback func()) {
	// Setup termination signal listener
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		callback()
	}()
}
