package service

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestUniFarm_GetProduct(t *testing.T) {
	uniFarmTestData := getStruct("product")
	uni := UniFarm("", "")

	products := uni.getProducts(uniFarmTestData)
	assert.True(t, len(products) > 0)
	if len(products) > 0 {
		assert.Equal(t, "ПРЕЗИДЕНТ СПРЕЙ ДЛЯ ПОЛОСТИ РТА МАНДАРИНОВЫЙ ВКУС Б/СПИРТА 20МЛ", products[0].ProductName)
		assert.Equal(t, "АВЕН ВОДА ТЕРМАЛЬНАЯ 300МЛ 119742", products[1].ProductName)
	}
}

func TestUniFarm_GetReceipt(t *testing.T) {
	uniFarmTestData := getStruct("receipt")
	uni := UniFarm("", "")

	receipts := uni.getReceipts(&uniFarmTestData)
	assert.True(t, len(receipts) > 0)
	if len(receipts) > 0 {
		assert.Equal(t, "2019-06-29T11:55:49", receipts[0].Date)
		assert.Equal(t, "ДИП РИЛИФ ГЕЛЬ 100Г", receipts[0].ProductName)
		assert.Equal(t, "000000002518447", receipts[0].PartNumber)
		assert.Equal(t, "459.64", receipts[0].PriceSellIn)
		assert.Equal(t, "547", receipts[0].PriceSellOut)
		assert.Equal(t, "34452", receipts[0].Serial)
		assert.Equal(t, "Mentholatum Company Ltd", receipts[0].ManufacturerName)
		assert.Equal(t, "ПРОТЕК ЗАО", receipts[0].ProviderName)
		assert.Equal(t, "00106208856992", receipts[0].NumberKKT)
		assert.Equal(t, "8.21", receipts[0].DiscountPrice)
		assert.Equal(t, "UM011, г. Щербинка, ул. Барышевская роща, д.10", receipts[0].WarehouseName)
		assert.Equal(t, "1", receipts[0].Quantity)
		assert.Equal(t, "519.92", receipts[0].SumPriceSellIn)
		assert.Equal(t, "577", receipts[0].SumPriceSellOut)
		assert.Equal(t, "6-0000126245", receipts[0].DocumentNumber)
		assert.Equal(t, "49352", receipts[0].NumberFiscalDocument)
		assert.Equal(t, "9284000100064105", receipts[0].ManufNumberFNKKT)
		assert.Equal(t, "00106208856992", receipts[0].ManufNumberKKT)
		assert.Equal(t, "411;49 352;2;29.06.2019 11:55:49;3634276327;www.nalog.ru;000", receipts[0].FiscalData)
		assert.Equal(t, "1714.75", receipts[0].SumReceipt)
		assert.Equal(t, "5011501011353", receipts[0].Barcode)
		assert.Equal(t, "2021-09-01T00:00:00", receipts[0].ExpireDate)

		assert.Equal(t, "2019-06-29T12:42:39", receipts[1].Date)
		assert.Equal(t, "ТЕРАЛИДЖЕН ВАЛЕНТА ТАБЛ. П/ПЛЕН/ОБ. 5МГ №50", receipts[1].ProductName)
		assert.Equal(t, "11-000002601566", receipts[1].PartNumber)
		assert.Equal(t, "702.72", receipts[1].PriceSellIn)
		assert.Equal(t, "709.5", receipts[1].PriceSellOut)
		assert.Equal(t, "591117", receipts[1].Serial)
		assert.Equal(t, "Валента Фарм", receipts[1].ManufacturerName)
		assert.Equal(t, "ЕАПТЕКА ООО", receipts[1].ProviderName)
		assert.Equal(t, "00106204515143", receipts[1].NumberKKT)
		assert.Equal(t, "12.18", receipts[1].DiscountPrice)
		assert.Equal(t, "UM029, ул. Красного Маяка, д. 9", receipts[0].WarehouseName)
		assert.Equal(t, "1", receipts[1].Quantity)
		assert.Equal(t, "702.72", receipts[1].SumPriceSellIn)
		assert.Equal(t, "812", receipts[1].SumPriceSellOut)
		assert.Equal(t, "11-100073376", receipts[1].DocumentNumber)
		assert.Equal(t, "27216", receipts[1].NumberFiscalDocument)
		assert.Equal(t, "9289000100166159", receipts[1].ManufNumberFNKKT)
		assert.Equal(t, "00106204515143", receipts[1].ManufNumberKKT)
		assert.Equal(t, "213;27 216;2;29.06.2019 12:42:39;3228979574;www.nalog.ru;000", receipts[1].FiscalData)
		assert.Equal(t, "799.82", receipts[1].SumReceipt)
		assert.Equal(t, "4602193011656", receipts[1].Barcode)
		assert.Equal(t, "2019-12-01T00:00:00", receipts[1].ExpireDate)
	}
}

func getStruct(t string) UniFarmResponse {
	// Open our jsonFile
	jsonFile, err := os.Open("./data/testReceipt_20200325.json")
	if t == "product" {
		jsonFile, err = os.Open("./data/testProduct.json")
	}
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	var uniFarmResponse UniFarmResponse
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &uniFarmResponse)

	return uniFarmResponse
}
