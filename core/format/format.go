package format

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"log"
	"pixel/core/model"
	"pixel/helper"
	"pixel/store"
	"pixel/store/service"
	"strconv"
	"strings"
	"time"
)

// Interface описание для создания новых форматов
type Interface interface {
	Parse()
	GetCache(key string) (interface{}, bool)
	GetMP() service.MarketPlaceInterface
	GetOFDCache(key string) (interface{}, bool)
	GetERPCache(key string) (interface{}, bool)
	SetOFDCache(key string, val interface{}, duration time.Duration)
	SetERPCache(key string, val interface{}, duration time.Duration)
	GetDate() time.Time
}

func nameMatching(f Interface, erpReceipts []*model.Receipt, ofdReceiptCacheItems map[string]cache.Item) (receipts []store.Receipt, checkOfdProductNames []string) {
	// проходимся по чекам ОФД из кеша
	for _, cacheItem := range ofdReceiptCacheItems {
		if documentList, ok := cacheItem.Object.([]model.Document); ok {
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
						if matchItem.ID == "" {
							continue
						}

						var erpReceipt *model.Receipt
						// ищем совпадения чеков из OFD с ERP
						for _, erpReceiptItem := range erpReceipts {
							var isFound bool
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
										erpName := strings.ReplaceAll(strings.ToLower(erpReceiptItem.Name), "№", "n")
										if strings.EqualFold(erpName, val.Name) ||
											strings.EqualFold(helper.Cut(erpName, 32), helper.Cut(val.Name, 32)) {
											matchItem = val
											isFound = true
											break
										}
									}
								}
							}

							if isFound {
								erpReceipt = erpReceiptItem
								break
							}
						}

						dateR := time.Unix(document.DateTime, 0)
						//TODO: Подумать что можно сделать с этим безумным нэймингом полей
						rc := store.Receipt{
							DateTime:             dateR.Format(time.RFC3339Nano),
							FiscalDocumentNumber: document.FiscalDocumentNumber,
							KktRegID:             document.KktRegID,
							Link:                 document.Link,
							Name:                 document.ProductName,
							Ofd:                  document.Ofd,
							Price:                document.ProductPrice,
							PriceSellIn:          document.Nds20,
							ProductID:            matchItem.ID,
							Quantity:             fmt.Sprint(document.ProductQuantity),
							CreatedAt:            time.Now().Format(time.RFC3339Nano),
							UpdatedAt:            time.Now().Format(time.RFC3339Nano),
							TotalSum:             document.TotalSum,
							Total:                document.ProductTotalPrice,
						}

						if erpReceipt != nil {
							rc.IsValidated = true
							pointName := strings.Split(erpReceipt.PharmacyID, ",")
							rc.PointName = pointName[0]
							rc.SupplerName = erpReceipt.Supplier
							rc.SupplerInn = erpReceipt.SupplierINN
							rc.Series = erpReceipt.Series
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
	return receipts, checkOfdProductNames

}

// OfdProductsForMatching подготовка продуктов к отправке в МП
func OfdProductsForMatching(ofdProductNames []string) []service.Product {
	products := make([]service.Product, 0, len(ofdProductNames))
	for _, name := range ofdProductNames {
		products = append(products, service.Product{
			Name:      strings.ToLower(name),
			CreatedAt: time.Now().Format(time.RFC3339Nano),
		})
	}
	return products
}

// ErpProductsForMatching подготовка продуктов к отправке в МП
func ErpProductsForMatching(f Interface, erpProducts []service.Product) []service.Product {
	var (
		checkErpProducts []service.Product
	)
	// проходимся по товарам из ЕРП - определяем какие нужно по новой сматчить
	for _, erpProduct := range erpProducts {
		needCheck := checkProduct(f, erpProduct)

		if needCheck {
			checkErpProducts = append(checkErpProducts, erpProduct)
		}
	}
	products := make([]service.Product, 0, len(checkErpProducts))
	for _, erpProduct := range checkErpProducts {
		productName := strings.Replace(erpProduct.Name, "№", "N", -1)
		var product service.Product
		product.Name = strings.ToLower(productName)
		product.Manufacturer = erpProduct.Manufacturer
		product.PartNumber = erpProduct.PartNumber
		product.Serial = erpProduct.Serial
		product.Stock = erpProduct.Stock
		product.WarehouseName = erpProduct.WarehouseName
		product.SupplerName = erpProduct.SupplerName
		product.SupplerInn = erpProduct.SupplerInn
		product.CreatedAt = erpProduct.CreatedAt
		products = append(products, product)
	}
	return products
}

func checkProduct(f Interface, erpProduct service.Product) bool {
	needCheck := true
	productName := strings.Replace(erpProduct.Name, "№", "N", -1)
	if cacheItem, ok := f.GetERPCache(strings.ToLower(productName)); ok {
		if matchItemList, ok := cacheItem.([]service.MatchProductItem); ok {
			for _, matchItem := range matchItemList {
				// если название, ИНН поставщика, остатки - совпадают, то матчит по новой не нужно
				if strings.EqualFold(matchItem.Name, erpProduct.Name) && matchItem.SupplierInn == erpProduct.SupplerInn && matchItem.Stock == erpProduct.Stock {
					needCheck = false
					break
				}
			}
		}
	}
	return needCheck
}

func getMatchProducts(f Interface) {
	getMatchFromSystem(f, true, false)
	getMatchFromSystem(f, false, true)
}

func getMatchFromSystem(f Interface, isOfd, isErp bool) {
	filterDate := f.GetDate().Format("02.01.2006")
	matchProducts, err := f.GetMP().GetMatchProducts(filterDate, isOfd, isErp)
	if err != nil {
		log.Printf("[ERROR] Ошибка получения продуктов %v", err)
	}
	if len(matchProducts.Data) > 0 {
		cacheItemsByName := make(map[string][]service.MatchProductItem)
		for _, matchItem := range matchProducts.Data {
			cacheItemsByName[matchItem.Name] = append(cacheItemsByName[matchItem.Name], matchItem)
		}
		for key, val := range cacheItemsByName {
			if isOfd {
				f.SetOFDCache(helper.Cut(key, 32), val, 12*time.Hour)
			}
			if isErp {
				f.SetERPCache(helper.Cut(key, 32), val, 12*time.Hour)
			}
		}
	}
}
