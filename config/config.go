package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var Cfg *Config

type Config struct {
	Debug              bool               `long:"debug" env:"DEBUG" default:"false" description:"Режим отладки"`
	WatchingTime       time.Duration      `long:"watchingTime" env:"WATCHING_TIME" default:"10" description:"Как часто проверять папку в минутах"`
	Store              StoreGroup         `group:"store" namespace:"store" env-namespace:"STORE"`
	BackupLocation     string             `long:"backup" env:"BACKUP_PATH" default:"./var/backup" description:"backups location"`
	Format             string             `long:"format" env:"FORMAT" description:"Формат обмена данными с ERP" choice:"XML1" choice:"CSV12" choice:"CSV14"  choice:"unifarm" default:"CSV14"`
	Files              Files              `group:"file" namespace:"file" env-namespace:"FILE"`
	MarketplaceOptions MarketplaceOptions `group:"marketplaceOptions" namespace:"MarketplaceOptions" env-namespace:"MARKETPLACE_OPTIONS"`
	OfdOptions         []OfdOptions       `group:"ofdOptions" namespace:"ofdOptions" env-namespace:"OFD_OPTIONS"`
	UniFarmOptions     UniFarmOptions
	UnicoOptions       UnicoOptions `group:"unicoOptions" namespace:"unicoOptions" env-namespace:"UNICO_OPTOONS"`
}

type StoreGroup struct {
	Path    string `long:"path" env:"BOLT_PATH" default:"./var" description:"parent dir for bolt files"`
	Timeout string `long:"timeout" env:"BOLT_TIMEOUT" default:"30s" description:"bolt timeout"`
}

type Files struct {
	DaysToCheck   string `long:"daysToCheck" env:"DAYS_TO_CHECK" default:"7" description:"Количество дней для проверки-загрузки (0 - без ограничений)"`
	SourceFolder  string `long:"sourceFolder" env:"SOURCE_FOLDER" default:"./var/source" description:"Имя исходной папки с файлами"`
	WorkingFolder string `long:"workingFolder" env:"WORKING_FOLDER" default:"./var" description:"Имя рабочей папки"`
}

type MarketplaceOptions struct {
	BaseUrl  string `long:"baseUrl" env:"BASE_URL" default:"https://stage.api.pharmecosystem.ru" description:"Адрес сервера"`
	Username string `long:"username" env:"USERNAME" default:"your@email.com" description:"Ваш email, он же логин"`
	Password string `long:"password" env:"PASSWORD" default:"your-password" description:"Ваш пароль"`
}
type OfdOptions struct {
	Type        string `long:"type" env:"TYPE" description:"выбор ОФД" choice:"OFD_YA" default:"OFD_YA"`
	AccessToken string `long:"accessToken" env:"ACCESS_TOKEN" default:"...." description:"Ключи доступа"`
}

type UniFarmOptions struct {
	Username string
	Password string
	Date     string
}

type UnicoOptions struct {
	SqlDriver  string `json:"sql_driver" env:"SQL_DRIVER"`
	ConnString string `long:"connstring" env:"UC_CONNSTRING" description:"Хост для подключения к базе unico"`
}

func Load() {
	storeGroup := StoreGroup{
		Path:    os.Getenv("SQLITE_PATH"),
		Timeout: os.Getenv("BOLT_TIMEOUT"),
	}
	fileStore := Files{
		DaysToCheck:   os.Getenv("DAYS_TO_CHECK"),
		SourceFolder:  os.Getenv("SOURCE_FOLDER"),
		WorkingFolder: os.Getenv("WORKING_FOLDER"),
	}
	marketplaceOptions := MarketplaceOptions{
		BaseUrl:  os.Getenv("MP_BASE_URL"),
		Username: os.Getenv("MP_USERNAME"),
		Password: os.Getenv("MP_PASSWORD"),
	}

	uniFarm := UniFarmOptions{
		Username: os.Getenv("UF_USERNAME"),
		Password: os.Getenv("UF_PASSWORD"),
		Date:     os.Getenv("UF_DATE"),
	}

	unico := UnicoOptions{
		SqlDriver:  os.Getenv("SQL_DRIVER"),
		ConnString: os.Getenv("UC_CONNSTRING"),
	}

	var ofdOptions []OfdOptions
	ofdOptions = append(ofdOptions, OfdOptions{
		Type:        os.Getenv("OFD_TYPE"),
		AccessToken: os.Getenv("OFD_TOKEN"),
	})

	for i := 1; i <= 50; i++ {
		stringType := fmt.Sprintf("OFD_TYPE_%d", i)
		stringTypeAccessToken := fmt.Sprintf("OFD_TOKEN_%d", i)
		if os.Getenv(stringType) != "" && os.Getenv(stringTypeAccessToken) != "" {
			ofdOptions = append(ofdOptions, OfdOptions{
				Type:        os.Getenv(stringType),
				AccessToken: os.Getenv(stringTypeAccessToken),
			})
		}
	}
	watchingTime, _ := strconv.Atoi(os.Getenv("WATCHING_TIME"))

	Cfg = &Config{
		Debug:              false,
		WatchingTime:       time.Duration(watchingTime),
		Store:              storeGroup,
		BackupLocation:     os.Getenv("BACKUP_LOCATION"),
		Format:             os.Getenv("FORMAT"),
		Files:              fileStore,
		MarketplaceOptions: marketplaceOptions,
		OfdOptions:         ofdOptions,
		UniFarmOptions:     uniFarm,
		UnicoOptions:       unico,
	}

	debug := os.Getenv("DEBUG")
	if debug == "true" {
		Cfg.Debug = true
	}
}
