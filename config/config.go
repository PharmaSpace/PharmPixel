package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Cfg глобальная перменная
var Cfg *Config

// Config структура конфигуроуионного файла
type Config struct {
	Debug              bool
	WatchingTime       time.Duration
	BackupLocation     string
	Format             string
	FormatDate         int
	Files              Files
	MarketplaceOptions MarketplaceOptions
	OfdOptions         []OfdOptions
	UniFarmOptions     UniFarmOptions
	UnicoOptions       UnicoOptions
	PixelOptions       PixelOptions
}

// Files Рабочие папки
type Files struct {
	DaysToCheck   string
	SourceFolder  string
	WorkingFolder string
	BackupFolder  string
}

// MarketplaceOptions gподключение к маркетплейсу
type MarketplaceOptions struct {
	BaseURL  string
	Username string
	Password string
}

// OfdOptions опции для подключение к ОФД
type OfdOptions struct {
	Type        string
	AccessToken string
	IsLocal     string
}

// UniFarmOptions подключение к Юнифарме
type UniFarmOptions struct {
	Username string
	Password string
	Date     string
}

// UnicoOptions подключение к Unico
type UnicoOptions struct {
	SQLDriver  string
	ConnString string
}

type PixelOptions struct {
	CSVType string
}

// Load загружаем данные
func Load() {
	fileStore := Files{
		DaysToCheck:   os.Getenv("DAYS_TO_CHECK"),
		SourceFolder:  os.Getenv("SOURCE_FOLDER"),
		WorkingFolder: os.Getenv("WORKING_FOLDER"),
		BackupFolder:  os.Getenv("BACKUP_FOLDER"),
	}

	marketplaceOptions := MarketplaceOptions{
		BaseURL:  os.Getenv("MP_BASE_URL"),
		Username: os.Getenv("MP_USERNAME"),
		Password: os.Getenv("MP_PASSWORD"),
	}

	uniFarm := UniFarmOptions{
		Username: os.Getenv("UF_USERNAME"),
		Password: os.Getenv("UF_PASSWORD"),
		Date:     os.Getenv("UF_DATE"),
	}

	unico := UnicoOptions{
		SQLDriver:  os.Getenv("SQL_DRIVER"),
		ConnString: os.Getenv("UC_CONNSTRING"),
	}

	pixel := PixelOptions{
		CSVType: os.Getenv("PIXEL_CSV_TYPE"),
	}

	var ofdOptions []OfdOptions
	ofdOptions = append(ofdOptions, OfdOptions{
		Type:        os.Getenv("OFD_TYPE"),
		AccessToken: os.Getenv("OFD_TOKEN"),
		IsLocal:     os.Getenv("OFD_LOCAL"),
	})

	for i := 1; i <= 50; i++ {
		stringType := fmt.Sprintf("OFD_TYPE_%d", i)
		stringTypeAccessToken := fmt.Sprintf("OFD_TOKEN_%d", i)
		stringTypeIsLocal := fmt.Sprintf("OFD_LOCAL_%d", i)
		if os.Getenv(stringType) != "" && os.Getenv(stringTypeAccessToken) != "" {
			ofdOpts := OfdOptions{
				Type:        os.Getenv(stringType),
				AccessToken: os.Getenv(stringTypeAccessToken),
			}
			if os.Getenv(stringTypeIsLocal) != "" {
				ofdOpts.IsLocal = os.Getenv(stringTypeIsLocal)
			}
			ofdOptions = append(ofdOptions, ofdOpts)
		}
	}
	watchingTime, _ := strconv.Atoi(os.Getenv("WATCHING_TIME"))
	formatDate := 0
	if os.Getenv("FORMAT_DATE") != "" {
		formatDate, _ = strconv.Atoi(os.Getenv("FORMAT_DATE"))
	}
	Cfg = &Config{
		Debug:              false,
		WatchingTime:       time.Duration(watchingTime),
		BackupLocation:     os.Getenv("BACKUP_LOCATION"),
		Format:             os.Getenv("FORMAT"),
		FormatDate:         formatDate,
		Files:              fileStore,
		MarketplaceOptions: marketplaceOptions,
		OfdOptions:         ofdOptions,
		UniFarmOptions:     uniFarm,
		UnicoOptions:       unico,
		PixelOptions:       pixel,
	}

	debug := os.Getenv("DEBUG")
	if debug == "true" {
		Cfg.Debug = true
	}
}
