package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("worker", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() {
	cfg, err := miner.NewConfig(configFlag)
	if err != nil {
		fmt.Printf("cannot load config: %s\r\n", err)
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.Logging().Level, true)
	ctx := log.WithLogger(context.Background(), logger)

	key, err := cfg.ETH().LoadKey()
	if err != nil {
		log.G(ctx).Error("failed load private key", zap.Error(err))
		os.Exit(1)
	}

	if _, err := os.Stat(cfg.UUIDPath()); os.IsNotExist(err) {
		ioutil.WriteFile(cfg.UUIDPath(), []byte(uuid.New()), 0660)
	}

	uuidData, err := ioutil.ReadFile(cfg.UUIDPath())
	if err != nil {
		log.G(ctx).Error("cannot load uuid", zap.Error(err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(ctx)
	m, err := miner.NewMiner(cfg, miner.WithContext(ctx), miner.WithKey(key), miner.WithUUID(string(uuidData)))
	if err != nil {
		log.G(ctx).Error("cannot create worker instance", zap.Error(err))
		os.Exit(1)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()
	}()

	go util.StartPrometheus(ctx, cfg.MetricsListenAddr())

	if err = m.Serve(); err != nil {
		log.G(ctx).Error("Server stop", zap.Error(err))
	}

	m.Close()
}
