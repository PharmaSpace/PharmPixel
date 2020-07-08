package format

import (
	"Pixel/config"
	"Pixel/core/model"
	"Pixel/store"
	"Pixel/store/service"
	"archive/zip"
	"encoding/csv"
	"fmt"
	"github.com/PharmaSpace/OfdYa"
	"github.com/PharmaSpace/ofdru"
	"github.com/PharmaSpace/oneofd"
	"github.com/PharmaSpace/platformOfd"
	"github.com/PharmaSpace/sbis"
	"github.com/PharmaSpace/taxcom"
	serviceLib "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"golang.org/x/text/encoding/charmap"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type partner struct {
	Config      *config.Config
	DataService *service.DataStore
	MP          *service.Marketpalce
	Log         serviceLib.Logger
	cache       *cache.Cache
}

func Partner(c *config.Config, dataService *service.DataStore, mp *service.Marketpalce, cache *cache.Cache, log serviceLib.Logger) *partner {
	return &partner{Config: c, DataService: dataService, cache: cache, MP: mp, Log: log}
}

func (p *partner) Parse() {
	p.getMachProduct()
	files := p.readDir(p.Config.Files.SourceFolder)
	for _, file := range files {
		isZip, err := regexp.MatchString(".zip", file.Name())
		if isZip {
			_, err := p.DataService.Get(fmt.Sprintf("1%s", file.Name()))
			if err != nil {
				p.unzip(fmt.Sprintf("%s/%s", p.Config.Files.SourceFolder, file.Name()), p.Config.Files.WorkingFolder)
				_, err := p.DataService.Create(store.File{Name: fmt.Sprintf("1%s", file.Name())})
				if err != nil {
					p.Log.Errorf("Ошибка создания записи в хранилище: %v", err)
				}
			}
		}

		if err != nil {
			p.Log.Errorf("Ошибка сопостовления архивов: %v", err)
		}
	}
	filesUnzip := p.readDir(p.Config.Files.WorkingFolder)
	for _, file := range filesUnzip {
		filePath := fmt.Sprintf("%s/%s", p.Config.Files.WorkingFolder, file.Name())
		mov, _ := regexp.MatchString(".mov", strings.ToLower(file.Name()))
		ost, _ := regexp.MatchString(".ost", strings.ToLower(file.Name()))
		if mov {
			orders := p.parsingOrderFile(filePath)
			p.sendOrders(orders)
			err := os.Remove(filePath)
			if err != nil {
				p.Log.Errorf("Не возиожно удалить файл: %v", err)
			}
		}
		if ost {
			products := p.parsingProductFile(filePath)
			p.sendProduct(products)
			err := os.Remove(filePath)
			if err != nil {
				p.Log.Errorf("Не возиожно удалить файл: %v", err)
			}
		}
	}

	p.getMachProduct()
}

type PartnerType struct {
	DType               string
	NDok                string
	DDok                string
	Supplier            string
	SupplierInn         string
	NKkm                string
	NChek               string
	FIOChek             string
	DiskT               string
	DiskSum             string
	SumZak              string
	SumRozn             string
	PPTeg               string
	DrugCode            string
	DrugName            string
	DrugProducerCode    string
	DrugProducerName    string
	DrugProducerCountry string
	DrugBar             string
	CenaZak             string
	CenaRozn            string
	Quant               float64
	Serial              string
	Godn                string
	Barecode            string
}

func (p *partner) sendOrders(orders []PixelOrder) {
	receipts := []store.Receipt{}
	receiptsN := []store.ReceiptN{}

	for _, order := range orders {
		datePay, err := time.Parse(time.RFC3339, order.Date)
		receipt, checkReceiptErr := p.checkReceipt(order.Name, order.InvoiceNumber, datePay, order.TotalPrice)
		if err == nil && receipt.Link == "" {
			continue
		}
		product, err := p.DataService.GeProduct(order.Name)
		if err != nil {
			p.Log.Errorf("sendOrders->GeProduct: %s %v", order.Name, err.Error())
		}
		if checkReceiptErr == nil {
			dateTime := time.Unix(receipt.DateTime, 0)
			dateR := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), dateTime.Hour(), dateTime.Minute(), dateTime.Second(), 11, time.Local)
			rc := store.Receipt{
				DateTime:             dateR.Format(time.RFC3339Nano), //2020-02-05T10:11:08
				FiscalDocumentNumber: receipt.FiscalDocumentNumber,
				KktRegId:             receipt.KktRegId,
				Link:                 receipt.Link,
				Name:                 order.Name,
				Ofd:                  receipt.Ofd,
				Price:                receipt.ProductPrice,
				ProductId:            product.ID,
				Quantity:             order.TotalNumber,
				TotalSum:             receipt.TotalSum,
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

func (p *partner) sendProduct(record []PixelProduct) {
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

func (p *partner) parsingOrderFile(file string) (orders []PixelOrder) {
	records, _ := p.convertCsvToStruct(file, 25)

	for _, record := range records {
		record[21] = strings.Replace(record[21], ",", ".", -1)

		record[20] = strings.Replace(record[20], ",", ".", -1)
		priceWVat, err := strconv.ParseFloat(record[20], 64)
		if err != nil {
			p.Log.Errorf("Конвертация цены позиции %v", err)
		}
		record[11] = strings.Replace(record[11], ",", ".", -1)
		totalPrice, err := strconv.ParseFloat(record[11], 64)
		if err != nil {
			p.Log.Errorf("Конвертация цены итоговой цены %v", err)
		}

		date, err := time.Parse("02012006", record[2])
		if err != nil {
			p.Log.Errorf("Конвертация даты %v", err)
		}
		orders = append(orders, PixelOrder{
			Date:           date.Format(time.RFC3339Nano),
			KKM:            record[5],
			InvoiceNumber:  record[6],
			Manufacturer:   record[16],
			Supplier:       record[3],
			SupplierINN:    record[4],
			Name:           record[14],
			PriceWoVat:     0,
			PriceWVat:      int(priceWVat * float64(100)),
			Vat:            0,
			TotalPrice:     int(totalPrice * float64(100)),
			TotalNumber:    record[21],
			ShipmentNumber: "",
			Series:         record[22],
		})
	}

	return orders
}

func (p *partner) parsingProductFile(file string) (products []PixelProduct) {
	records, _ := p.convertCsvToStruct(file, 25)

	for _, record := range records {
		var quantity float64
		record[21] = strings.Replace(record[21], ",", ".", -1)
		quantity, err := strconv.ParseFloat(record[21], 64)
		if err != nil {
			p.Log.Errorf("convert string to float64: %v", err)
		}
		products = append(products, PixelProduct{
			Name:            record[14],
			Supplier:        record[3],
			SupplierINN:     record[4],
			Manufacturer:    record[16],
			CountryOfOrigin: record[17],
			ShipmentNumber:  record[1],
			Series:          record[22],
			Inventory:       quantity,
			EAN:             record[24],
		})
	}

	return products
}

func (p *partner) convertCsvToStruct(file string, l int) (outRecord [][]string, err error) {
	readFile, err := ioutil.ReadFile(file)
	sourceFile := string(readFile)
	sourceFile = strings.Replace(sourceFile, "\r", "", -1)
	if err != nil {
		p.Log.Errorf("Error read file: %v", err)
	}
	outFile, _ := charmap.Windows1251.NewDecoder().Bytes(readFile)

	fileIO := strings.NewReader(string(outFile))
	r := csv.NewReader(fileIO)
	r.Comma = ';'
	r.LazyQuotes = false
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if len(record) < l {
			p.Log.Errorf("Не совпадает количество значений: ", record)
			continue
		}
		if record[0] == "D_TYPE" {
			continue
		}
		if err != nil {
			p.Log.Errorf("Ошибка парсинга файла: ", err)
			continue
		}
		outRecord = append(outRecord, record)
	}
	return outRecord, err
}

func (p *partner) readDir(dir string) (files []os.FileInfo) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		p.Log.Errorf("[ERR] Ошибка чтения директории | %v", err)
	}
	return files
}

func (p *partner) unzip(fileName string, targetDir string) {
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

func (p *partner) checkReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	productName = p.cut(strings.ToLower(productName), 32)
	if receipts, ok := p.cache.Get(strings.ToUpper(productName)); ok {
		switch t := receipts.(type) {
		case []platformOfd.Receipt:
			for _, v := range t {
				if fd == v.FD {
					document.Link = v.Link
					document.TotalSum = v.Price
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
				}
			}
		case []ofdru.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
				}
			}
		case []OfdYa.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
				}
			}
		case []*sbis.Receipt:
			for _, v := range t {
				if fd == strconv.Itoa(v.RequestNumber) || totalPrice == v.TotalSum {
					for _, i := range v.Items {
						if i.Name == strings.ToUpper(productName) {
							date, _ := time.Parse("2006-01-02T15:04:05", v.ReceiveDateTime)
							document.DateTime = date.Unix()
							document.Link = v.Url
							document.FiscalDocumentNumber = v.FiscalDocumentNumber
							document.KktRegId = v.KktRegID
							document.ProductPrice = i.Price
							document.TotalSum = v.TotalSum
						}
					}
				}
			}
		case []taxcom.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
				}
			}
		default:
			p.Log.Errorf("Обработка данной ОФД не доступна")
		}
	}
	return document, err
}

func (p *partner) getMachProduct() {
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

func (p *partner) cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) >= limit {
		return string(runes[:limit])
	}
	return text
}
