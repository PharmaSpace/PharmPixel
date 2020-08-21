package format

import (
	"Pixel/core/model"
	"Pixel/helper"
	"Pixel/store"
	"Pixel/store/service"
	"fmt"
	"github.com/patrickmn/go-cache"
	"strconv"
	"strings"
	"time"
)

type FormatInterface interface {
	Parse()
	GetCache(key string) (interface{}, bool)
	GetMP() service.MarketPlaceInterface
	GetOFDCache(key string) (interface{}, bool)
	GetERPCache(key string) (interface{}, bool)
	SetOFDCache(key string, val interface{}, duration time.Duration)
	SetERPCache(key string, val interface{}, duration time.Duration)
	GetDate() time.Time
}

func nameMatching(f FormatInterface, erpReceipts []*model.Receipt, ofdRecieptCacheItems map[string]cache.Item) ([]store.Receipt, []string) {
	checkOfdProductNames := make([]string, 0)
	receipts := make([]store.Receipt, 0)
	// проходимся по чекам ОФД из кеша
	for cacheId, _ := range ofdRecieptCacheItems {
		if cacheItem, ok := f.GetCache(cacheId); ok {
			if documentList, ok := cacheItem.([]model.Document); ok {
				for _, document := range documentList {
					// ищим матчинг по названию товара
					text := helper.Cut(strings.ToLower(document.ProductName), 32)
					if cacheItem, ok := f.GetOFDCache(text); ok {
						if matchItemList, ok := cacheItem.([]service.MatchProductItem); ok {
							// находим матчинг разрешающий выгрузку
							matchItem := service.MatchProductItem{}
							for _, val := range matchItemList {
								if val.Export && len(val.ID) > 0 {
									matchItem = val
									break
								}
							}
							if len(matchItem.ID) == 0 {
								continue
							}

							var erpReciept *model.Receipt
							// ищем совпадения чеков из OFD с ERP
							for _, erpReceiptItem := range erpReceipts {
								var isFound = false
								if len(erpReceiptItem.InvoiceNumber) > 0 {
									// если InvoiceNumber не пустой то ищем совпадение по FiscalDocumentNumber
									isFound = erpReceiptItem.InvoiceNumber == fmt.Sprint(document.FiscalDocumentNumber) ||
										erpReceiptItem.InvoiceNumber == document.FiscalDocumentNumber2 ||
										erpReceiptItem.InvoiceNumber == document.FiscalDocumentNumber3
								} else {
									// в противном случае по сумме чека
									tp, _ := strconv.ParseInt(erpReceiptItem.TotalPrice, 10, 32)
									isFound = int(tp) == document.TotalSum
								}

								// проходимся по матчингу - ищем соответсвие по названию
								if isFound {
									isFound = false
									for _, val := range matchItemList {
										if val.Export && len(val.ID) > 0 {
											if strings.ReplaceAll(strings.ToLower(erpReceiptItem.Name), "№", "n") == strings.ToLower(val.Name) ||
												helper.Cut(strings.ReplaceAll(strings.ToLower(erpReceiptItem.Name), "№", "n"), 32) == cut(strings.ToLower(val.Name), 32) {
												matchItem = val
												isFound = true
												break
											}
										}
									}
								}

								if isFound {
									erpReciept = erpReceiptItem
									break
								}
							}

							dateR := time.Unix(document.DateTime, 0)

							rc := store.Receipt{
								DateTime:             dateR.Format(time.RFC3339Nano),
								FiscalDocumentNumber: document.FiscalDocumentNumber,
								KktRegId:             document.KktRegId,
								Link:                 document.Link,
								Name:                 document.ProductName,
								Ofd:                  document.Ofd,
								Price:                document.ProductPrice,
								PriceSellIn:          document.Nds20,
								ProductId: matchItem.ID,
								Quantity:  fmt.Sprint(document.ProductQuantity),
								CreatedAt: time.Now().Format(time.RFC3339Nano),
								UpdatedAt: time.Now().Format(time.RFC3339Nano),
								TotalSum:  document.TotalSum,
							}

							if erpReciept != nil {
								rc.IsValidated = true
								pointName := strings.Split(erpReciept.PharmacyID, ",")
								rc.PointName = pointName[0]
								rc.SupplerName = erpReciept.Supplier
								rc.SupplerInn = erpReciept.SupplierINN
								rc.Series = erpReciept.Series

							}

							receipts = append(receipts, rc)
						}
					} else {
						if !helper.ContainsString(checkOfdProductNames, strings.TrimSpace(document.ProductName)) {
							checkOfdProductNames = append(checkOfdProductNames, strings.TrimSpace(document.ProductName))
						}

					}
				}

			}
		}
	}
	return receipts, checkOfdProductNames

}

func OfdProductsForMatching(ofdProductNames []string) []service.Product {
	var products []service.Product
	for _, name := range ofdProductNames {
		products = append(products, service.Product{
			Name: strings.ToLower(name),
		})
	}
	return products
}
func ErpProductsForMatching(f FormatInterface, erpProducts []service.Product) []service.Product {
	var checkErpProducts []service.Product
	// проходимся по товарам из ЕРП - определяем какие нужно по новой сматчить
	for _, erpProduct := range erpProducts {
		needCheck := checkProduct(f, erpProduct)

		if needCheck {
			checkErpProducts = append(checkErpProducts, erpProduct)
		}
	}
	var products []service.Product
	for _, erpProduct := range checkErpProducts {
		productName := strings.Replace(erpProduct.Name, "№", "N", -1)
		products = append(products, service.Product{
			Name:          strings.ToLower(productName),
			Manufacturer:  erpProduct.Manufacturer,
			PartNumber:    erpProduct.PartNumber,
			Serial:        erpProduct.Serial,
			Stock:         erpProduct.Stock,
			WarehouseName: erpProduct.WarehouseName,
			SupplerName:   erpProduct.SupplerName,
			SupplerInn:    erpProduct.SupplerInn,
		})
	}
	return products
}

func checkProduct(f FormatInterface, erpProduct service.Product) bool {
	needCheck := true
	productName := strings.Replace(erpProduct.Name, "№", "N", -1)
	if cacheItem, ok := f.GetERPCache(strings.ToLower(productName)); ok {
		if matchItemList, ok := cacheItem.([]service.MatchProductItem); ok {
			for _, matchItem := range matchItemList {
				// если название, ИНН поставщика, остатки - совпадают, то матчит по новой не нужно
				if strings.ToLower(matchItem.Name) == strings.ToLower(erpProduct.Name) && matchItem.SupplierInn == erpProduct.SupplerInn && matchItem.Stock == erpProduct.Stock {
					needCheck = false
					break
				}
			}
		}
	}
	return needCheck
}

func getMatchProducts(f FormatInterface) {
	getMatchFromSystem(f, true, false)
	getMatchFromSystem(f, false, true)
}

func getMatchFromSystem(f FormatInterface, isOfd, isErp bool) {
	filterDate := f.GetDate().Format("02.01.2006")
	matchProducts := f.GetMP().GetMatchProducts(filterDate, isOfd, isErp)
	if len(matchProducts.Data) > 0 {
		cacheItemsByName := make(map[string][]service.MatchProductItem)
		for _, matchItem := range matchProducts.Data {
			cacheItemsByName[matchItem.Name] = append(cacheItemsByName[matchItem.Name], matchItem)
		}
		for key, val := range cacheItemsByName {
			if isOfd {
				f.SetOFDCache(strings.ToLower(key), val, 12*time.Hour)
			}
			if isErp {
				f.SetERPCache(strings.ToLower(key), val, 12*time.Hour)
			}
		}
	}
}

func cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) >= limit {
		return string(runes[:limit])
	}
	return text
}
