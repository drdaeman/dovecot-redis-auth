package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var rdb *redis.Client

func main() {
	var debug bool
	listenAddress, err := NewListenAddress("unix:///run/dovecot2/auth-dict-service.socket")
	if err != nil {
		panic(fmt.Errorf("bug: malformed default listen address: %w", err))
	}
	redisUrl := "redis://localhost:6379/0"
	flag.BoolVar(&debug, "debug", false, "Switch to debug/development logging")
	flag.StringVar(&redisUrl, "redis", redisUrl, "Redis URL")
	flag.Var(listenAddress, "listen", "Listen address")
	flag.Parse()

	var logger *zap.Logger
	if debug {
		logConfig := zap.NewProductionConfig()
		logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		logConfig.Sampling = nil
		logger, err = logConfig.Build()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		_ = fmt.Sprintf("Failed to initialize logger: %s", err)
		os.Exit(1)
		return
	}
	defer Close("sync logger", nil, logger.Sync)
	logger.Debug("Starting")

	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		logger.Error("Failed to parse Redis URL", zap.Error(err))
		os.Exit(1)
	}
	rdb = redis.NewClient(opt)

	listener, err := listenAddress.Listen()
	if err != nil {
		logger.Error("Failed to set up listener", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("Listening", zap.String("Address", listener.Addr().String()))
	defer Close("close listener", nil, listener.Close)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection", zap.Error(err))
			continue
		}
		go NewConn(conn, logger, handleRequest)
	}
}
