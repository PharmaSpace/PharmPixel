package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// UniFarm структура
type UniFarm struct {
	Username string
	Password string
}

// UniFarmReceiptResponse структура
type UniFarmReceiptResponse struct {
	Response ResponseForQueryReceipt `json:"ОтветНаДанныеПоЗапросу"`
}

// UniFarmProductResponse структура
type UniFarmProductResponse struct {
	Response ResponseForQueryProduct `json:"ОтветНаДанныеПоЗапросу"`
}

// ResponseForQueryReceipt структура
type ResponseForQueryReceipt struct {
	Products []UniFarmReceipt `json:"МассивДанных"`
}

// ResponseForQueryProduct структура
type ResponseForQueryProduct struct {
	Products []UniFarmProduct `json:"МассивДанных"`
}

// UniFarmReceipt структура
type UniFarmReceipt struct {
	Date                 string `json:"Период"`                    // дата, время продажи
	ProductCode          string `json:"ТоварКод"`                  // Код товара в нашей номенклатуре
	ProductName          string `json:"ТоварНаименование"`         // Наименование товара в нашей номенклатуре
	PartNumber           string `json:"Партия"`                    // Номер партии (внутренний)
	PriceSellIn          string `json:"ЦенаЗакуп"`                 // Цена закупки
	PriceSellOut         string `json:"ЦенаРозн"`                  // Цена продажи
	Serial               string `json:"Серия"`                     // серия товаропроизводителя
	ManufacturerName     string `json:"ПроизводительНаименование"` // Наименование производителя
	ProviderName         string `json:"ПоставщикНаименование"`     // Наименование поставщика
	InnProvider          string `json:"ПоставщикИНН"`              // ИНН поставщика
	NumberKKT            string `json:"ЗаводскойНомерККТ"`         // Внутренний номер кассы
	WarehouseName        string `json:"СкладНаименование"`         // Наименование и адрес аптеки
	DiscountPrice        string `json:"СуммаСкидки"`               // сумма предоставленной скидки
	Quantity             string `json:"Количество"`                // количество проданного товара в чеке
	SumPriceSellIn       string `json:"СуммаЗакуп"`                // Сумма закупки товара
	SumPriceSellOut      string `json:"СуммаРозн"`                 // Розничная сумма продажи
	DocumentNumber       string `json:"НомерЧека"`                 // Номер чека в нашей программе
	NumberFiscalDocument string `json:"НомерФискальногоДокумента"` // фискальный номер документа
	ManufNumberFNKKT     string `json:"ЗаводскойНомерФН"`          // Заводской номер ФН ККТ
	FiscalData           string `json:"ФискальныеДанные"`          // Строка с фискальными данными получаемая с кассы.(содержит фискальный признак документа ФПД)
	SumReceipt           string `json:"СуммаЧека"`                 // итоговая сумма чека ()
	Barcode              string `json:"Штрихкод"`                  // ШК - товара в нашей системе EAN13
	ExpireDate           string `json:"СрокГодности"`              //  - срок годности проданного товара
}

// UniFarmProduct структура
type UniFarmProduct struct {
	ProductName      string `json:"ТоварНаименование"`
	ManufacturerName string `json:"ПроизводительНаименование"`
	SupplierName     string `json:"ПоставщикИНН"`
	PartNumber       string `json:"Партия"`
	Serial           string `json:"Серия"`
	Date             string `json:"ДатаЧека"`
	Stock            string
}

// NewUniFarm интеграция с юнифармой
func NewUniFarm(username, password string) *UniFarm {
	return &UniFarm{Username: username, Password: password}
}

// GetProduct получение продукта
func (uf *UniFarm) GetProduct(date time.Time) (products []UniFarmProduct) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)
	client := &http.Client{}
	req, err = http.NewRequest("GET", fmt.Sprintf("http://1c.unifarma.ru:180/unifarma_aptechka/hs/exchange?type=ПолучитьПродажиПартнеров&ДатаНачала=%s&ДатаОкончания=%s", date.Format("02.01.2006"), date.Format("02.01.2006")), nil)
	if err != nil {
		log.Printf("[ERROR] Не можем получить данные с http://1c.unifarma.ru:180")
	}
	if req == nil {
		log.Printf("[ERROR] пустой запрос")
		return
	}
	req.SetBasicAuth(uf.Username, uf.Password)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("[ERROR] Не можем получить данные с http://1c.unifarma.ru:180")
	}
	if resp == nil {
		log.Printf("[ERROR] пустой ответ")
		return
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERR] не получается прочитать данные: %v", err)
	}
	uniFarmData := &UniFarmProductResponse{}
	err = uniFarmData.UnmarshalJSON(bodyText)
	if err != nil {
		log.Printf("[ERR] не получается маппить данные: %v", err)
	}
	return uniFarmData.Response.Products
}

// GetReceipt получение продаж
func (uf *UniFarm) GetReceipt(date time.Time) (receipts []UniFarmReceipt) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)
	client := &http.Client{}
	req, err = http.NewRequest("GET", fmt.Sprintf("http://1c.unifarma.ru:180/unifarma_aptechka/hs/exchange?type=ПолучитьПродажиПартнеров&ДатаНачала=%s&ДатаОкончания=%s", date.Format("02.01.2006"), date.Format("02.01.2006")), nil)
	if err != nil {
		log.Printf("[ERROR] Не можем получить данные с http://1c.unifarma.ru:180")
	}
	if req == nil {
		log.Print("[ERROR] пустой запрос")
		return
	}
	req.SetBasicAuth(uf.Username, uf.Password)
	resp, err = client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Не можем получить данные с http://1c.unifarma.ru:180")
	}
	if resp == nil {
		log.Print("[ERROR] пустой ответ")
		return
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERR] не получается прочитать данные: %v", err)
	}
	uniFarmData := &UniFarmReceiptResponse{}
	err = uniFarmData.UnmarshalJSON(bodyText)
	if err != nil {
		log.Printf("[ERR] не получается маппить данные: %v", err)
	}
	return uniFarmData.Response.Products
}
