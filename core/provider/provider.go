package provider

import (
	"github.com/patrickmn/go-cache"
	"pixel/core/model"
	"strings"
	"time"
)

// Provider структура
type Provider interface {
	CheckReceipt(productName, fp string, datePay time.Time, totalPrice int) (document model.Document, err error)
	GetReceipts(date time.Time)
	GetName() string
}

// TODO: Добавить валидацию конфига провайдера
// GetProvider тип провайдера
func GetProvider(c *cache.Cache, provider, credentials string) Provider {
	switch provider {
	case "ofd-ya":
		return &Ofdya{
			Cache: c,
			Type:  provider,
			Token: credentials,
		}
	case "1ofd":
		cr := strings.Split(credentials, ":")
		return &OneOfd{
			Cache:    c,
			Type:     provider,
			Login:    cr[0],
			Password: cr[1],
		}
	case "taxcom":
		cr := strings.Split(credentials, ":")
		return &TaxCom{
			Cache:        c,
			Type:         provider,
			IDIntegrator: cr[0],
			Login:        cr[1],
			Password:     cr[2],
		}
	case "platformofd":
		cr := strings.Split(credentials, ":")
		return &PlatformOfd{
			Cache:    c,
			Type:     provider,
			Login:    cr[0],
			Password: cr[1],
		}
	case "ofdru":
		cr := strings.Split(credentials, ":")
		return &OfdRu{
			Cache:    c,
			Type:     provider,
			Inn:      cr[0],
			Login:    cr[1],
			Password: cr[2],
		}
	case "sbis":
		cr := strings.Split(credentials, ":")
		return &Sbis{
			Cache:    c,
			Type:     provider,
			Inn:      cr[0],
			Login:    cr[1],
			Password: cr[2],
		}
	default:
		panic("OFD NOT SUPPORT!!!!!")
	}
}
