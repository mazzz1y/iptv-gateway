package main

import (
	"context"
	"flag"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.TODO()
	configPath := flag.String("config-dir", "./config", "path to configuration")
	flag.Parse()

	logging.Info(ctx, "starting iptv-gateway", "config_path", *configPath)

	c, err := config.Load(*configPath)
	if err != nil {
		logging.Error(ctx, err, "failed to load config")
		os.Exit(1)
	}

	logging.SetLevel(c.LogLevel)

	s, err := server.NewServer(c)
	if err != nil {
		logging.Error(ctx, err, "failed to create server")
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.Start(); err != nil {
			logging.Error(ctx, err, "server error")
			os.Exit(1)
		}
	}()

	<-stop
	if err := s.Stop(); err != nil {
		logging.Error(ctx, err, "error during shutdown")
		os.Exit(1)
	}
}
