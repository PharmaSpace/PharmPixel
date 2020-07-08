package format

import (
	"Pixel/config"
	"Pixel/core/model"
	"Pixel/store"
	"Pixel/store/service"
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/PharmaSpace/OfdYa"
	"github.com/PharmaSpace/ofdru"
	"github.com/PharmaSpace/oneofd"
	"github.com/PharmaSpace/platformOfd"
	"github.com/PharmaSpace/sbis"
	"github.com/PharmaSpace/taxcom"
	serviceLib "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type pixel struct {
	Config      *config.Config
	DataService *service.DataStore
	MP          *service.Marketpalce
	Log         serviceLib.Logger
	cache       *cache.Cache
}

type PixelProduct struct {
	PharmacyID      string
	PharmacyAddress string
	Name            string
	Supplier        string
	SupplierINN     string
	Manufacturer    string
	CountryOfOrigin string
	ShipmentNumber  string
	Series          string
	Inventory       float64
	EAN             string
}

type PixelOrder struct {
	PharmacyID      string
	PharmacyAddress string
	Date            string
	KKM             string
	InvoiceNumber   string
	Manufacturer    string
	Supplier        string
	SupplierINN     string
	Name            string
	PriceWoVat      int
	PriceWVat       int
	Vat             int
	TotalPrice      int
	TotalNumber     string
	ShipmentNumber  string
	Series          string
}

func Pixel(c *config.Config, dataService *service.DataStore, mp *service.Marketpalce, cache *cache.Cache, log serviceLib.Logger) *pixel {
	return &pixel{Config: c, DataService: dataService, cache: cache, MP: mp, Log: log}
}

func (p *pixel) Parse() {
	p.getMachProduct()
	files := p.readDir(p.Config.Files.SourceFolder)
	for _, file := range files {
		isZip, err := regexp.MatchString(".zip", file.Name())
		if isZip {
			_, err := p.DataService.Get(file.Name())
			if err != nil {
				p.unzip(fmt.Sprintf("%s/%s", p.Config.Files.SourceFolder, file.Name()), p.Config.Files.WorkingFolder)
				_, err := p.DataService.Create(store.File{Name: file.Name()})
				if err != nil {
					p.Log.Errorf("Ошибка фиксации файла", err)
				}
			}
		}

		if err != nil {
			log.Print(err)
		}
	}
	filesUnzip := p.readDir(p.Config.Files.WorkingFolder)
	for _, file := range filesUnzip {
		product, _ := regexp.MatchString("products.csv", file.Name())
		order, _ := regexp.MatchString("orders.csv", file.Name())
		filePath := fmt.Sprintf("%s/%s", p.Config.Files.WorkingFolder, file.Name())
		if product {
			products := p.parsingProductFile(filePath)
			p.sendProduct(products)
			err := os.Remove(filePath)
			if err != nil {
				p.Log.Errorf("Ошибка удаления файла продуктов %+v", err)
			}
		}
		if order {
			orders := p.parsingOrderFile(filePath)
			p.sendOrders(orders)
			err := os.Remove(filePath)
			if err != nil {
				p.Log.Errorf("Ошибка удаления файла заказов %+v", err)
			}
		}
	}
	p.getMachProduct()
}

func (p *pixel) sendOrders(orders []PixelOrder) {
	receipts := []store.Receipt{}
	receiptsN := []store.ReceiptN{}

	for _, order := range orders {
		datePay, err := time.Parse("02.01.2006 15:04:05", order.Date)
		if err != nil {
			datePay, err = time.Parse("02.01.2006 15:04", order.Date)
			if err != nil {
				p.Log.Errorf("[pixel]ошибка преобразования даты %s | %+v", order.Date, err)
			}
		}
		receipt, checkReceiptErr := p.checkReceipt(order.Name, order.InvoiceNumber, datePay, order.TotalPrice)
		if err == nil && receipt.Link == "" {
			continue
		}
		product, err := p.DataService.GeProduct(order.Name)
		if err != nil {
			p.Log.Errorf("sendOrders->GeProduct: %s %v", order.Name, err.Error())
		}

		if checkReceiptErr == nil {
			quantity, err := strconv.Atoi(receipt.ProductQuantity.String())
			if err != nil {
				quantity = 1
			}
			rc := store.Receipt{
				DateTime:             datePay.Local().Format(time.RFC3339Nano), //2020-02-05T10:11:08
				FiscalDocumentNumber: receipt.FiscalDocumentNumber,
				KktRegId:             receipt.KktRegId,
				Link:                 receipt.Link,
				Name:                 order.Name,
				Ofd:                  receipt.Ofd,
				Price:                receipt.ProductPrice,
				ProductId:            product.ID,
				Quantity:             receipt.ProductQuantity.String(),
				TotalSum:             receipt.TotalSum,
				Total:                receipt.ProductPrice * quantity,
				SupplerName:          order.SupplierINN,
				PointName:            order.PharmacyID,
				Series:               order.Series,
				CreatedAt:            time.Now().Format(time.RFC3339Nano),
				UpdatedAt:            time.Now().Format(time.RFC3339Nano),
			}
			if product.ID != "" && product.Export {
				receipts = append(receipts, rc)
			} else {
				_, err := p.DataService.CreateReceipt(rc)
				if err != nil {
					p.Log.Errorf("sendOrders->CreateReceipt: %v", err)
				}
			}
		} else {
			rc := store.ReceiptN{
				DatePay:      order.Date,
				Manufacture:  order.Manufacturer,
				Name:         order.Name,
				Number:       order.InvoiceNumber,
				PointName:    order.PharmacyID,
				PriceSellOut: order.PriceWVat,
				Series:       order.Series,
				ProductId:    product.ID,
				Quantity:     order.TotalNumber,
				SupplerName:  order.SupplierINN,
			}
			if product.ID != "" && product.Export {
				receiptsN = append(receiptsN, rc)
			} else {
				_, err := p.DataService.CreateReceiptN(rc)
				if err != nil {
					p.Log.Errorf("sendOrders->CreateReceiptN: %v", err)
				}
			}
		}
	}
	if len(receipts) > 0 {
		p.MP.SendReceipt(receipts)
	}
	if len(receiptsN) > 0 {
		p.MP.SendReceiptN(receiptsN)
	}
}
func (p *pixel) sendProduct(record []PixelProduct) {
	products := []service.Product{}

	for _, record := range record {
		products = append(products, service.Product{
			Name:          record.Name,
			Manufacturer:  record.Manufacturer,
			Stock:         record.Inventory,
			PartNumber:    record.ShipmentNumber,
			Serial:        record.Series,
			WarehouseName: record.PharmacyID,
			SupplerName:   record.Supplier,
		})
	}
	p.MP.SendProduct(products)
}

func (p *pixel) parsingOrderFile(file string) (orders []PixelOrder) {
	records, _ := p.convertCsvToStruct(file, 15)

	for i, record := range records {
		if i == 0 {
			continue
		}

		priceWoVat, _ := strconv.ParseFloat(strings.Replace(record[9], ",", ".", -1), 64)
		priceWVat, _ := strconv.ParseFloat(strings.Replace(record[10], ",", ".", -1), 64)
		vat, _ := strconv.ParseFloat(strings.Replace(record[11], ",", ".", -1), 64)
		totalPrice, _ := strconv.ParseFloat(strings.Replace(record[12], ",", ".", -1), 64)
		orders = append(orders, PixelOrder{
			PharmacyID:      record[0],
			PharmacyAddress: record[1],
			Date:            record[2],
			KKM:             record[3],
			InvoiceNumber:   record[4],
			Manufacturer:    record[5],
			Supplier:        record[6],
			SupplierINN:     record[7],
			Name:            record[8],
			PriceWoVat:      int(priceWoVat),
			PriceWVat:       int(priceWVat),
			Vat:             int(vat),
			TotalPrice:      int(totalPrice),
			TotalNumber:     record[13],
			ShipmentNumber:  record[14],
			Series:          record[15],
		})
	}

	return orders
}

func (p *pixel) parsingProductFile(file string) (products []PixelProduct) {

	records, _ := p.convertCsvToStruct(file, 7)
	for i, record := range records {
		if i == 0 {
			continue
		}
		record[9] = strings.Replace(record[9], ",", ".", -1)
		quantity, err := strconv.ParseFloat(record[9], 64)
		if err != nil {
			p.Log.Errorf("convert string to float64: %v", err)
		}
		products = append(products, PixelProduct{
			PharmacyID:      record[0],
			PharmacyAddress: record[1],
			Name:            record[2],
			Supplier:        record[3],
			SupplierINN:     record[4],
			Manufacturer:    record[5],
			CountryOfOrigin: record[6],
			ShipmentNumber:  record[7],
			Series:          record[8],
			Inventory:       quantity,
			EAN:             record[10],
		})
	}

	return products
}

func (p *pixel) convertCsvToStruct(file string, l int) (outRecord [][]string, err error) {
	readFile, err := ioutil.ReadFile(file)
	if err != nil {
		p.Log.Errorf("Error read file: %v", err)
	}
	sourceFile := string(readFile)
	sourceFile = strings.Replace(sourceFile, "\r", "", -1)

	fileIO := strings.NewReader(sourceFile)
	r := csv.NewReader(fileIO)
	r.Comma = ','
	r.LazyQuotes = true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if len(record) < l {
			p.Log.Errorf("Не совпадает количество значений в файле: %s", file)
			continue
		}
		if err != nil {
			p.Log.Errorf("Ошибка парсинга файла: %s | %v", file, err)
			continue
		}
		outRecord = append(outRecord, record)
	}
	return outRecord, err
}

func (p *pixel) readDir(dir string) (files []os.FileInfo) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		p.Log.Errorf("[ERR] Ошибка чтения директории %s| %v", dir, err)
	}
	return files
}

func (p *pixel) unzip(fileName string, targetDir string) {
	zipReader, _ := zip.OpenReader(fileName)
	for _, file := range zipReader.Reader.File {

		zippedFile, err := file.Open()
		if err != nil {
			p.Log.Errorf("[ERR] Ошибка чтения файла | %v", err)
		}
		defer zippedFile.Close()

		extractedFilePath := filepath.Join(
			targetDir,
			file.Name,
		)

		if file.FileInfo().IsDir() {
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				p.Log.Errorf("[ERR] Ошибка распаковки файла | %v", err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				p.Log.Errorf("[ERR] Ошибка копирование содержимого архива | %v", err)
			}
		}
	}
}

func (p *pixel) checkReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	productName = p.cut(strings.ToLower(productName), 32)
	if receipts, ok := p.cache.Get(productName); ok {
		if config.Cfg.Debug {
			f, err := os.OpenFile("receipt.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
			}
			defer f.Close()
			j, _ := json.Marshal(receipts)
			if _, err := f.WriteString(string(j)); err != nil {
				log.Println(err)
			}
		}
		switch t := receipts.(type) {
		case []platformOfd.Receipt:
			for _, v := range t {
				if fd == v.FD {
					document.Link = v.Link
					document.TotalSum = v.Price
					document.Ofd = "platformofd"
				}
			}
		case []oneofd.Receipt:
			for _, v := range t {
				fdInt, _ := strconv.Atoi(fd)
				fd = fmt.Sprintf("%d", fdInt)
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
					document.Ofd = "1ofd"
					document.FiscalDocumentNumber = fdInt
					document.KktRegId = v.KktRegId
					for _, i := range v.Products {
						if p.cut(strings.ToLower(i.Name), 32) == productName {
							document.ProductPrice = i.Price
							document.ProductQuantity = json.Number(fmt.Sprintf("%d", i.Quantity))
						}
					}
				}
			}
		case []ofdru.Receipt:
			for _, v := range t {
				fdInt, _ := strconv.Atoi(fd)
				fd = fmt.Sprintf("%d", fdInt)
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
					document.Ofd = "ofdru"
					document.FiscalDocumentNumber = fdInt
					document.KktRegId = v.KktRegId
					for _, i := range v.Products {
						if p.cut(strings.ToLower(i.Name), 32) == productName {
							document.ProductPrice = i.Price
							document.ProductQuantity = json.Number(fmt.Sprintf("%d", i.Quantity))
						}
					}
				}
			}
		case []OfdYa.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
					document.Ofd = "ofd-ya"
				}
			}
		case []*sbis.Receipt:
			for _, v := range t {
				if fd == strconv.Itoa(v.RequestNumber) || fd == strconv.Itoa(v.FiscalDocumentNumber) || totalPrice == v.TotalSum {
					for _, i := range v.Items {
						if p.cut(strings.ToLower(i.Name), 32) == productName {
							date, _ := time.Parse("2006-01-02T15:04:05", v.ReceiveDateTime)
							document.DateTime = date.Unix()
							document.Link = v.Url
							document.FiscalDocumentNumber = v.FiscalDocumentNumber
							document.KktRegId = v.KktRegID
							document.ProductPrice = i.Price
							document.TotalSum = v.TotalSum
							document.Ofd = "sbis"
						}
					}
				}
			}
		case []taxcom.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP || v.Price == totalPrice {
					fiscalDocumentNumber, _ := strconv.Atoi(v.FD)
					document.KktRegId = v.KktRegId
					document.Link = v.Link
					document.TotalSum = v.Price
					document.FiscalDocumentNumber = fiscalDocumentNumber
					document.Ofd = "taxcom"
					for _, i := range v.Products {
						if p.cut(strings.ToLower(i.Name), 32) == productName {
							document.ProductQuantity = json.Number(fmt.Sprintf("%d", i.Quantity))
							document.ProductPrice = i.Price
							document.ProductName = i.Name
							document.ProductTotalPrice = i.TotalPrice
						}
					}
				}
			}
		default:
			p.Log.Errorf("Обработка данной ОФД не доступна")
		}
	}
	return document, err
}

func (p *pixel) cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) >= limit {
		return string(runes[:limit])
	}
	return text
}

func (p *pixel) getMachProduct() {
	matchProducts := p.MP.GetMatchProduct()
	for _, matchProduct := range matchProducts.Data {
		product := store.Product{
			ID:          matchProduct.ID,
			Name:        matchProduct.Name,
			Export:      matchProduct.Export,
			Manufacture: matchProduct.Manufacturer,
		}
		p.DataService.CreateProduct(product)
	}
}
