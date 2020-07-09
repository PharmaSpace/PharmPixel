// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Simple service that only works by printing a log message every few seconds.
package main

import (
	"Pixel/config"
	"Pixel/core"
	"Pixel/store/engine"
	"Pixel/store/service"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/juju/fslock"
	windowsService "github.com/kardianos/service"
	"github.com/pkg/errors"
)

var revision = "2.3.5"
var logger windowsService.Logger

type program struct {
	dataService *service.DataStore
	exit        chan struct{}
}

func (p *program) Start(s windowsService.Service) error {
	if windowsService.Interactive() {
		logger.Info("Running in terminal.")
	} else {
		logger.Info("Running under service manager.")
	}
	p.exit = make(chan struct{})

	go p.run()
	return nil
}

func (p *program) run() error {
	logger.Infof("I'm running %v.", windowsService.Platform())
	logger.Infof("Version: %s.", revision)
	logger.Infof("Format: %s.", config.Cfg.Format)
	ticker := time.NewTicker(10 * time.Second)
	if config.Cfg.WatchingTime > 0 {
		ticker = time.NewTicker(config.Cfg.WatchingTime * time.Minute)
	}

	for {
		select {
		case <-ticker.C:
			lock := fslock.New(config.Cfg.Files.WorkingFolder + "/../pixel.lock")
			lock.Lock()
			lock.Unlock()
			marketplace := &service.Marketpalce{
				Revision: revision,
				Log:      logger,
				BaseUrl:  config.Cfg.MarketplaceOptions.BaseUrl,
				Username: config.Cfg.MarketplaceOptions.Username,
				Password: config.Cfg.MarketplaceOptions.Password,
			}

			c := core.Core{
				Log:         logger,
				Version:     revision,
				DataService: p.dataService,
				SourceDir:   config.Cfg.Files.SourceFolder,
				Marketplace: marketplace,
				Config:      config.Cfg,
			}

			c.Exec()
			os.Remove(config.Cfg.Files.WorkingFolder + "/../pixel.lock")
		case <-p.exit:
			ticker.Stop()
			return nil
		}
	}
}

func (p *program) Stop(s windowsService.Service) error {
	if e := p.dataService.Close(); e != nil {
		logger.Warning("[WARN] failed to close data store, %s", e)
	}
	logger.Info("I'm Stopping!")
	close(p.exit)
	os.Remove(config.Cfg.Files.WorkingFolder + "/../pixel.lock")
	return nil
}

func main() {
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	options := make(windowsService.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	svcConfig := &windowsService.Config{
		Name:        "Pixel",
		DisplayName: "Pixel",
		Description: "Данный сервис служит для обмена информацией между маркетплейсом и аптечной точко",
		Dependencies: []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"},
		Option: options,
	}
	storeEngine, err := makeDataStore()
	if err != nil {
		logger.Error("failed to make data store engine")
	}

	dataService := &service.DataStore{
		Engine: storeEngine,
	}
	prg := &program{
		dataService: dataService,
	}
	s, err := windowsService.New(prg, svcConfig)
	if err != nil {
		logger.Error(err)
	}
	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		logger.Error(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				logger.Error(err)
			}
		}
	}()

	if len(*svcFlag) != 0 {
		err := windowsService.Control(s, *svcFlag)
		if err != nil {
			logger.Errorf("Valid actions: %q\n", windowsService.ControlAction)
			logger.Error(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

func getDump() string {
	maxSize := 5 * 1024 * 1024
	stacktrace := make([]byte, maxSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxSize {
		length = maxSize
	}
	return string(stacktrace[:length])
}

func init() {
	_ = godotenv.Load(".env")
	config.Load()

	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			logger.Infof("[INFO] SIGQUIT detected, dump:\n%s", getDump())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}

func makeDataStore() (result engine.Interface, err error) {
	log.Printf("[INFO] make data store, type=sqllite")

	if err = makeDirs(config.Cfg.Store.Path); err != nil {
		return nil, errors.Wrap(err, "failed to create sqllite store")
	}
	result, err = engine.NewSQLiteDB(fmt.Sprintf("%s/pixel.sqlite", config.Cfg.Store.Path))
	return result, errors.Wrap(err, "can't initialize data store")
}

// mkdir -p for all dirs
func makeDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil { // If path is already a directory, MkdirAll does nothing
			return errors.Wrapf(err, "can't make directory %s", dir)
		}
	}
	return nil
}
