package core

import (
	"encoding/gob"
	"fmt"
	serviceLib "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"log"
	"pixel/config"
	"pixel/core/format"
	"pixel/core/model"
	"pixel/core/provider"
	"pixel/db"
	"pixel/sentry"
	"pixel/store/service"
	"strconv"
	"time"
)

// Core структура
type Core struct {
	Log         serviceLib.Logger
	Sentry      *sentry.Sentry
	Version     string
	SourceDir   string
	Marketplace *service.Marketplace
	Config      *config.Config
}

// Exec включение ядра
func (c *Core) Exec() {
	cReceipt := cache.New(5*time.Minute, 10*time.Minute)
	if c.Config.UniFarmOptions.Date != "" {
		start := c.getDate()
		for m := start; m.Unix() <= time.Now().Unix(); m = m.AddDate(0, 0, 1) {
			cReceipt.Flush()
			log.Print(m.Format(time.RFC3339))
			c.parse(m, cReceipt)
		}
		c.Config.UniFarmOptions.Date = ""
	}
	cReceipt.Flush()
	c.parse(c.getDate(), cReceipt)
}

func (c *Core) parse(date time.Time, cReceipt *cache.Cache) {
	var (
		err      error
		database db.DB
	)
	for i, ofd := range c.Config.OfdOptions {
		isLocal, _ := strconv.ParseBool(ofd.IsLocal)
		if isLocal {
			log.Printf("[ОФД #%v]Загрузка локальных данных ОФД %s за %s", i, ofd.Type, date.Format("02-01-2006"))
			gob.Register([]model.Document{})
			err = cReceipt.LoadFile(fmt.Sprintf("var/ofd/%s_%s[%s]", ofd.Type, date.Format("02-01-2006"), config.Cfg.MarketplaceOptions.Username))
			if err != nil {
				log.Printf("Не удалось найти локальные данные для ОФД %s за %s: %s", ofd.Type, date.Format("02-01-2006"), err)
				log.Printf("Получение данных из ОФД %s", ofd.Type)
				pr := provider.GetProvider(cReceipt, ofd.Type, ofd.AccessToken, c.Sentry)
				pr.GetReceipts(date)
				err = cReceipt.SaveFile(fmt.Sprintf("var/ofd/%s_%s[%s]", ofd.Type, date.Format("02-01-2006"), config.Cfg.MarketplaceOptions.Username))
				if err != nil {
					c.Sentry.Error(err)
					log.Printf("Не удалось сохранить данные для ОФД %s за %s: %s", ofd.Type, date.Format("02-01-2006"), err)
					isLocal = false
				} else {
					log.Printf("Локальные данные для ОФД %s за %s сохранены", ofd.Type, date.Format("02-01-2006"))
				}
			}
		}

		if !isLocal {
			log.Printf("[ОФД #%v]Получение данных из ОФД %s за %s", i, ofd.Type, date.Format("02-01-2006"))
			pr := provider.GetProvider(cReceipt, ofd.Type, ofd.AccessToken, c.Sentry)
			pr.GetReceipts(date)
		}
	}
	switch c.Config.Format {
	case "unifarm":
		uf := format.NewUniFarm(c.Config, c.Marketplace, cReceipt, c.Log, date, c.Sentry)
		uf.Parse()
	case "pixel":
		pixel := format.NewPixel(c.Config, c.Marketplace, cReceipt, c.Log, date, c.Sentry)
		pixel.Parse()
	/*case "partner":
	partner := format.Partner(c.Config, c.DataService, c.Marketplace, cReceipt, c.Log)
	partner.Parse()*/
	case "unico":
		database, err = format.ConnectToErpDB(c.Config)
		if err != nil {
			err = c.Log.Errorf("Ошибка подключения к базе данных ERP %s", err)
			if err != nil {
				c.Sentry.Error(err)
				log.Printf("Ошибка подключения к базе данных ERP %s", err)
			}
			break
		}
		defer database.Close()
		unico := format.NewUnico(c.Config, c.Marketplace, database, cReceipt, c.Log, date, c.Sentry)
		unico.Parse()
	default:
		err = c.Log.Errorf("Формат интеграции не поддерживается %s", c.Config.Format)
		if err != nil {
			c.Sentry.Error(err)
			log.Printf("Формат интеграции не поддерживается %s", c.Config.Format)
		}
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
