package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"pixel/config"
	"pixel/core"
	"pixel/sentry"
	"pixel/store/service"
	"runtime"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/juju/fslock"
	windowsService "github.com/kardianos/service"
)

var revision = "3.0.40"

var logger windowsService.Logger

type program struct {
	exit chan struct{}
}

func (p *program) Start(s windowsService.Service) error {
	var err error
	if windowsService.Interactive() {
		err = logger.Info("Running in terminal.")
		if err != nil {
			return err
		}
	} else {
		err = logger.Info("Running under service manager.")
		if err != nil {
			return err
		}
	}

	p.exit = make(chan struct{})

	go p.run()
	return nil
}

func (p *program) run() {
	err := logger.Infof("I'm running %v.", windowsService.Platform())
	if err != nil {
		log.Print(err)
	}
	err = logger.Infof("Version: %s.", revision)
	if err != nil {
		log.Print(err)
	}
	ticker := time.NewTicker(1 * time.Second)
	if config.Cfg.WatchingTime > 0 {
		ticker = time.NewTicker(config.Cfg.WatchingTime * time.Minute)
	}

	for {
		select {
		case <-ticker.C:
			lock := fslock.New(config.Cfg.Files.WorkingFolder + "/../pixel.lock")
			err = lock.Lock()
			if err != nil {
				log.Print(err)
			}
			err = lock.Unlock()
			if err != nil {
				log.Print(err)
			}
			marketplace := &service.Marketplace{
				Revision: revision,
				Log:      logger,
				BaseURL:  config.Cfg.MarketplaceOptions.BaseURL,
				Username: config.Cfg.MarketplaceOptions.Username,
				Password: config.Cfg.MarketplaceOptions.Password,
			}
			merchant, err := marketplace.WhoAmi()
			if err != nil {
				log.Print(err)
			}
			s := sentry.NewSentry(merchant.MerchantID, merchant.PointID, revision)
			c := core.Core{
				Log:         logger,
				Version:     revision,
				SourceDir:   config.Cfg.Files.SourceFolder,
				Marketplace: marketplace,
				Config:      config.Cfg,
				Sentry:      s,
			}

			c.Exec()
			err = os.Remove(config.Cfg.Files.WorkingFolder + "/../pixel.lock")
			if err != nil {
				s.Error(err)
			}
		case <-p.exit:
			ticker.Stop()
			return
		}
	}
}

func (p *program) Stop(s windowsService.Service) error {
	err := logger.Info("I'm Stopping!")
	if err != nil {
		log.Print("I'm Stopping!")
	}
	close(p.exit)
	err = os.Remove(config.Cfg.Files.WorkingFolder + "/../pixel.lock")
	if err != nil {
		log.Printf("[ERROR] remove lock %v", err)
	}
	return nil
}

func main() {
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	options := make(windowsService.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	svcConfig := &windowsService.Config{
		Name:         "Pixel",
		DisplayName:  "Pixel",
		Description:  "Данный сервис служит для обмена информацией между маркетплейсом и аптечной точко",
		Dependencies: []string{},
		Option:       options,
	}

	prg := &program{}
	s, err := windowsService.New(prg, svcConfig)
	if err != nil {
		err = logger.Error(err)
		if err != nil {
			log.Print(err)
		}
	}
	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Print(err)
	}

	go func() {
		for {
			err = <-errs
			if err != nil {
				err = logger.Error(err)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}()

	if *svcFlag != "" {
		err = windowsService.Control(s, *svcFlag)
		if err != nil {
			err = logger.Errorf("Valid actions: %q\n", windowsService.ControlAction)
			if err != nil {
				log.Print(err)
			}
			err = logger.Error(err)
			if err != nil {
				log.Print(err)
			}
		}
		return
	}
	err = s.Run()
	if err != nil {
		err = logger.Error(err)
		if err != nil {
			log.Print(err)
		}
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

//nolint:gochecknoinits // can't avoid it in this place
func init() {
	_ = godotenv.Load(".env")
	config.Load()

	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] SIGQUIT detected, dump:\n%s", getDump())
		}
	}()

	signal.Notify(sigChan, syscall.SIGQUIT)
}
