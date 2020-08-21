package provider

import (
	"Pixel/core/model"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

type Provider interface {
	CheckReceipt(productName string, fp string, datePay time.Time, totalPrice int) (document model.Document, err error)
	GetReceipts(date time.Time)
	GetName() string
}

func GetProvider(c *cache.Cache, provider string, credentials string) Provider {
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
			IdIntegrator: cr[0],
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

