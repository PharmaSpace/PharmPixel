// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Simple service that only works by printing a log message every few seconds.
package main

import (
	"Pixel/config"
	"Pixel/core"
	"Pixel/store/service"
	"flag"
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

var revision = "3.0.6"
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
	ticker := time.NewTicker(1 * time.Second)
	if config.Cfg.WatchingTime > 0 {
		ticker = time.NewTicker(config.Cfg.WatchingTime * time.Minute)
	}

	for {
		select {
		case <-ticker.C:
			lock := fslock.New(config.Cfg.Files.WorkingFolder + "/../pixel.lock")
			if config.Cfg.MarketplaceOptions.Username == "s.antsupov+pil@pharmecosystem.ru" ||
				config.Cfg.MarketplaceOptions.Username == "apt4@pharmaspace.ru" ||
				config.Cfg.MarketplaceOptions.Username == "apt3@pharmaspace.ru" ||
				config.Cfg.MarketplaceOptions.Username == "apt2@pharmaspace.ru" ||
				config.Cfg.MarketplaceOptions.Username == "apt@pharmaspace.ru" ||
				config.Cfg.MarketplaceOptions.Username == "887530+at5@mail.ru" ||
				config.Cfg.MarketplaceOptions.Username == "887530+at4@mail.ru" ||
				config.Cfg.MarketplaceOptions.Username == "887530+at3@mail.ru" ||
				config.Cfg.MarketplaceOptions.Username == "887530+at2@mail.ru" ||
				config.Cfg.MarketplaceOptions.Username == "887530+at1@mail.ru" ||
				config.Cfg.MarketplaceOptions.Username == "887530@mail.ru" {
				_, err := os.Stat(config.Cfg.Files.WorkingFolder + "/../updated.lock")
				if os.IsNotExist(err) {
					config.Cfg.UniFarmOptions.Date = "01.08.2020"
					os.Create(config.Cfg.Files.WorkingFolder + "/../updated.lock")
				} else {
					config.Cfg.UniFarmOptions.Date = time.Now().Format("01.08.2020")
				}
			}

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

	prg := &program{}
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
	envFlag := flag.String("env", ".env", "env file path")
	flag.Parse()
	_ = godotenv.Load(*envFlag)
	config.Load()

	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			logger.Infof("[INFO] SIGQUIT detected, dump:\n%s", getDump())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
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
