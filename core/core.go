package core

import (
	"Pixel/config"
	"Pixel/core/format"
	"Pixel/core/model"
	"Pixel/core/provider"
	"Pixel/store/service"
	"encoding/gob"
	"fmt"
	serviceLib "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"log"
	"strconv"
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
	if c.Config.UniFarmOptions.Date != "" {
		start := c.getDate()
		for m := start; m.Year() == start.Year(); m = m.AddDate(0, 1, 0) {
			if m.Month() != start.Month() {
				m = time.Date(m.Year(), m.Month(), 1, m.Hour(), m.Minute(), m.Second(), m.Nanosecond(), m.Location())
			}
			if m.Month() > time.Now().Month() {
				break
			}
			if c.iterationByDay(m, cReceipt) {
				break
			}
		}
	} else {
		c.parse(c.getDate(), cReceipt)
	}
}

func (c *Core) iterationByDay(t time.Time, cReceipt *cache.Cache) bool {
	for d := t; d.Month() == t.Month(); d = d.AddDate(0, 0, 1) {
		if d.Month() >= time.Now().Month() && d.Day() > time.Now().Day() {
			return true
		}
		log.Printf("Обрабока файлов за %s", d.Format(time.RFC3339))
		c.parse(d, cReceipt)
	}
	return false
}

func (c *Core) parse(date time.Time, cReceipt *cache.Cache) {
	for _, ofd := range c.Config.OfdOptions {
		isLocal, _ := strconv.ParseBool(ofd.IsLocal)
		if isLocal {
			log.Printf("Загрузка локальных данных ОФД %s за %s", ofd.Type, date.Format("01-02-2006"))
			gob.Register([]model.Document{})
			err := cReceipt.LoadFile(fmt.Sprintf("var/ofd/%s_%s", ofd.Type, date.Format("01-02-2006")))
			if err != nil {
				log.Printf("Не удалось найти локальные данные для ОФД %s за %s: %s", ofd.Type, date.Format("01-02-2006"), err)
				log.Printf("Получение данных из ОФД %s", ofd.Type)
				pr := provider.GetProvider(cReceipt, ofd.Type, ofd.AccessToken)
				pr.GetReceipts(date)
				err := cReceipt.SaveFile(fmt.Sprintf("var/ofd/%s_%s", ofd.Type, date.Format("01-02-2006")))
				if err != nil {
					log.Printf("Не удалось сохранить данные для ОФД %s за %s: %s", ofd.Type, date.Format("01-02-2006"), err)
				}
				log.Printf("Локальные данные для ОФД %s за %s сохранены", ofd.Type, date.Format("01-02-2006"))
				if err != nil {
					isLocal = false
				}
			}
		}
		if !isLocal {
			log.Printf("Получение данных из ОФД %s", ofd.Type)
			pr := provider.GetProvider(cReceipt, ofd.Type, ofd.AccessToken)
			pr.GetReceipts(date)
		}
	}
	switch c.Config.Format {
	case "unifarm":
		uf := format.UniFarm(c.Config, c.DataService, c.Marketplace, cReceipt, c.Log, date)
		uf.Parse()
	case "pixel":
		pixel := format.Pixel(c.Config, c.Marketplace, cReceipt, c.Log, date)
		pixel.Parse()
	/*case "partner":
	partner := format.Partner(c.Config, c.DataService, c.Marketplace, cReceipt, c.Log)
	partner.Parse()*/
	case "unico":
		db, err := format.ConnectToErpDB(c.Config)
		if err != nil {
			c.Log.Errorf("Ошибка подключения к базе данных ERP %s", err)
			break
		}
		defer db.Close()
		unico := format.Unico(c.Config, c.Marketplace, db, cReceipt, c.Log, date)
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
