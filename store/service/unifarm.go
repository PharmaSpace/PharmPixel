package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type uniFarm struct {
	Username string
	Password string
}
type UniFarmReceiptResponse struct {
	Response ResponseForQueryReceipt `json:"ОтветНаДанныеПоЗапросу"`
}

type UniFarmProductResponse struct {
	Response ResponseForQueryProduct `json:"ОтветНаДанныеПоЗапросу"`
}

type ResponseForQueryReceipt struct {
	Products []UniFarmReceipt `json:"МассивДанных"`
}

type ResponseForQueryProduct struct {
	Products []UniFarmProduct `json:"МассивДанных"`
}

type UniFarmReceipt struct {
	Date                 string `json:"Период"`                    // дата, время продажи
	ProductCode          string `json:"ТоварКод"`                  //Код товара в нашей номенклатуре
	ProductName          string `json:"ТоварНаименование"`         // Наименование товара в нашей номенклатуре
	PartNumber           string `json:"Партия"`                    //Номер партии (внутренний)
	PriceSellIn          string `json:"ЦенаЗакуп"`                 //Цена закупки
	PriceSellOut         string `json:"ЦенаРозн"`                  //Цена продажи
	Serial               string `json:"Серия"`                     //серия товаропроизводителя
	ManufacturerName     string `json:"ПроизводительНаименование"` //Наименование производителя
	ProviderName         string `json:"ПоставщикНаименование"`     //Наименование поставщика
	InnProvider          string `json:"ПоставщикИНН"`              //ИНН поставщика
	NumberKKT            string `json:"ЗаводскойНомерККТ"`         // Внутренний номер кассы
	WarehouseName        string `json:"СкладНаименование"`         //Наименование и адрес аптеки
	DiscountPrice        string `json:"СуммаСкидки"`               //сумма предоставленной скидки
	Quantity             string `json:"Количество"`                //количество проданного товара в чеке
	SumPriceSellIn       string `json:"СуммаЗакуп"`                //Сумма закупки товара
	SumPriceSellOut      string `json:"СуммаРозн"`                 //Розничная сумма продажи
	DocumentNumber       string `json:"НомерЧека"`                 //Номер чека в нашей программе
	NumberFiscalDocument string `json:"НомерФискальногоДокумента"` //фискальный номер документа
	ManufNumberFNKKT     string `json:"ЗаводскойНомерФН"`          //Заводской номер ФН ККТ
	ManufNumberKKT       string `json:"ЗаводскойНомерККТ"`         //заводской номер ККТ
	FiscalData           string `json:"ФискальныеДанные"`          //Строка с фискальными данными получаемая с кассы.(содержит фискальный признак документа ФПД)
	SumReceipt           string `json:"СуммаЧека"`                 //итоговая сумма чека ()
	Barcode              string `json:"Штрихкод"`                  // ШК - товара в нашей системе EAN13
	ExpireDate           string `json:"СрокГодности"`              //   - срок годности проданного товара
}

type UniFarmProduct struct {
	ProductName      string `json:"ТоварНаименование"`
	ManufacturerName string `json:"ПроизводительНаименование"`
	SupplierName     string `json:"ПоставщикИНН"`
	PartNumber       string `json:"Партия"`
	Serial           string `json:"Серия"`
	Stock            string
}

func UniFarm(username string, password string) *uniFarm {
	return &uniFarm{Username: username, Password: password}
}

func (uf *uniFarm) GetProduct(date time.Time) (products []UniFarmProduct) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://1c.unifarma.ru:180/unifarma_aptechka/hs/exchange?type=ПолучитьПродажиПартнеров&ДатаНачала=%s&ДатаОкончания=%s", date.Format("02.01.2006"), date.Format("02.01.2006")), nil)
	if err != nil {
		log.Printf("[ERR] Не можем получить данные с http://1c.unifarma.ru:180")
	}
	req.SetBasicAuth(uf.Username, uf.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERR] Не можем получить данные с http://1c.unifarma.ru:180")
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	uniFarmData := &UniFarmProductResponse{}
	err = json.Unmarshal(bodyText, uniFarmData)
	if err != nil {
		log.Printf("[ERR] не получается маппить данные: %v", err)
	}
	return uniFarmData.Response.Products
}

func (uf *uniFarm) GetReceipt(date time.Time) (receipts []UniFarmReceipt) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://1c.unifarma.ru:180/unifarma_aptechka/hs/exchange?type=ПолучитьПродажиПартнеров&ДатаНачала=%s&ДатаОкончания=%s", date.Format("02.01.2006"), date.Format("02.01.2006")), nil)
	if err != nil {
		log.Printf("[ERR] Не можем получить данные с http://1c.unifarma.ru:180")
	}
	req.SetBasicAuth(uf.Username, uf.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERR] Не можем получить данные с http://1c.unifarma.ru:180")
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	uniFarmData := &UniFarmReceiptResponse{}
	err = json.Unmarshal(bodyText, uniFarmData)
	if err != nil {
		log.Printf("[ERR] не получается маппить данные: %v", err)
	}
	return uniFarmData.Response.Products
}
