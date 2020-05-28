package core

import (
	"Pixel/config"
	"Pixel/core/format"
	"Pixel/core/provider"
	"Pixel/store/service"
	serviceLib "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"log"
	"time"
)

type Core struct {
	Log         serviceLib.Logger
	Version     string
	DataService *service.DataStore
	SourceDir   string
	Marketplace *service.Marketpalce
	Config      *config.Config
}

func (c *Core) Exec() {
	cReceipt := cache.New(5*time.Minute, 10*time.Minute)

	for _, ofd := range c.Config.OfdOptions {
		log.Printf("Получение данных из ОФД %s", ofd.Type)
		pr := provider.GetProvider(cReceipt, ofd.Type, ofd.AccessToken)
		pr.GetReceipts(c.getDate())
	}
	switch c.Config.Format {
	case "unifarm":
		uf := format.UniFarm(c.Config, c.DataService, c.Marketplace, cReceipt, c.Log)
		uf.Parse()
	case "pixel":
		pixel := format.Pixel(c.Config, c.DataService, c.Marketplace, cReceipt, c.Log)
		pixel.Parse()
	case "partner":
		partner := format.Partner(c.Config, c.DataService, c.Marketplace, cReceipt, c.Log)
		partner.Parse()
	case "unico":
		unico := format.Unico(c.Config, c.DataService, c.Marketplace, cReceipt, c.Log, c.getDate())
		unico.Parse()
	default:
		c.Log.Errorf("Формат интеграции не поддерживается %s", c.Config.Format)
	}
}

func (c *Core) getDate() time.Time {
	date := time.Now()
	if c.Config.UniFarmOptions.Date != "" {
		dt, _ := time.Parse("02.01.2006", c.Config.UniFarmOptions.Date)
		date = time.Date(dt.Year(), dt.Month(), dt.Day(), 23, 59, 59, 0, dt.Location())
	}
	if c.Config.Format == "partner" || c.Config.Format == "pixel" {
		date = date.AddDate(0, 0, -1)
	}

	return date
}
