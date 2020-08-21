package format

import (
	"Pixel/config"
	"Pixel/core/model"
	"Pixel/store/service"
	"archive/zip"
	"encoding/csv"
	"fmt"
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
	date             time.Time
	Config           *config.Config
	MP               service.MarketPlaceInterface
	Log              serviceLib.Logger
	matchingOfdCache *cache.Cache
	matchingErpCache *cache.Cache
	cache            *cache.Cache
}

func (p *pixel) GetCache(key string) (interface{}, bool) {
	if p.cache == nil {
		return nil, false
	}
	return p.cache.Get(key)
}

func (p *pixel) GetOFDCache(key string) (interface{}, bool) {
	if p.matchingOfdCache == nil {
		return nil, false
	}
	return p.matchingOfdCache.Get(key)
}

func (p *pixel) GetERPCache(key string) (interface{}, bool) {
	if p.matchingErpCache == nil {
		return nil, false
	}
	return p.matchingErpCache.Get(key)
}

func (p *pixel) SetOFDCache(key string, val interface{}, duration time.Duration) {
	if p.matchingOfdCache == nil {
		return
	}
	p.matchingOfdCache.Set(key, val, duration)
}


func (p *pixel) SetERPCache(key string, val interface{}, duration time.Duration) {
	if p.matchingErpCache == nil {
		return
	}
	p.matchingOfdCache.Set(key, val, duration)
}

func (p *pixel) GetMP() service.MarketPlaceInterface {
	return p.MP
}

func (p *pixel) GetDate() time.Time {
	return p.date
}

func Pixel(cf *config.Config, mp service.MarketPlaceInterface, ch *cache.Cache, log serviceLib.Logger, date time.Time) *pixel {
	matchingOfdCache := cache.New(5*time.Minute, 10*time.Minute)
	matchingErpCache := cache.New(5*time.Minute, 10*time.Minute)

	return &pixel{Config: cf, cache: ch, MP: mp, Log: log, matchingErpCache: matchingErpCache, matchingOfdCache: matchingOfdCache, date: date}
}

func (p *pixel) Parse() {
	getMatchProducts(p)

	// данные из кеша по чекам из ОФД
	ofdRecieptCacheItems := p.cache.Items()
	// чеки из ЕРП
	erpOrder, erpProducts := p.parseFiles()

	receipts, checkOfdProductNames := nameMatching(p, erpOrder, ofdRecieptCacheItems)
	if len(receipts) > 0 {
		p.MP.SendReceipt(receipts)
	}

	if len(checkOfdProductNames) > 0 {
		// отправляем товары из ofd на матчинг
		productsForMatching := OfdProductsForMatching(checkOfdProductNames)
		p.MP.SendOfdProducts(productsForMatching, true, false)
	}
	// отправляем товары из erp на матчинг
	productsForMatching := ErpProductsForMatching(p, p.convertProducts(erpProducts))
	p.MP.SendOfdProducts(productsForMatching, false, true)
}

func (p *pixel) getFiles(files []os.FileInfo) {
	for _, file := range files {
		isZip, err := regexp.MatchString(".zip", file.Name())
		if isZip {
			_, ok := p.GetOFDCache(file.Name())
			if ok {
				p.unzip(fmt.Sprintf("%s/%s", p.Config.Files.SourceFolder, file.Name()), p.Config.Files.WorkingFolder)
			}
		}

		if err != nil {
			log.Print(err)
		}
	}
}

func (p *pixel) parseFiles() ([]*model.Receipt, []*model.Product) {
	files := p.readDir(p.Config.Files.SourceFolder)
	p.getFiles(files)

	filesUnzip := p.readDir(p.Config.Files.WorkingFolder)
	AllOrders := make([]*model.Receipt, 0)
	AllProducts := make([]*model.Product, 0)
	for _, file := range filesUnzip {
		product, _ := regexp.MatchString("products.csv", file.Name())
		order, _ := regexp.MatchString("orders.csv", file.Name())
		filePath := fmt.Sprintf("%s/%s", p.Config.Files.WorkingFolder, file.Name())
		doneFilePath := fmt.Sprintf("%s/%s/%s", p.Config.Files.WorkingFolder, "done", file.Name())
		if product {
			products := p.parsingProductFile(filePath)
			err := os.Rename(filePath, doneFilePath)
			if err != nil {
				p.Log.Errorf("Ошибка удаления файла продуктов %+v", err)
			}
			AllProducts = append(AllProducts, products...)
		}
		if order {
			orders := p.parsingOrderFile(filePath)
			err := os.Rename(filePath, doneFilePath)
			if err != nil {
				p.Log.Errorf("Ошибка удаления файла заказов %+v", err)
			}
			AllOrders = append(AllOrders, orders...)
		}
	}
	return AllOrders, AllProducts
}

// convertProducts - конвертирует продукты для отправки на матчинг
func (p *pixel) convertProducts(record []*model.Product) []service.Product {
	products := []service.Product{}

	for _, record := range record {
		products = append(products, service.Product{
			Name:          record.Name,
			Manufacturer:  record.Manufacturer,
			PartNumber:    record.ShipmentNumber,
			Serial:        record.Series,
			Stock:         record.Inventory,
			WarehouseName: record.PharmacyID,
			SupplerName:   record.Supplier,
			SupplerInn:    record.SupplierINN,
		})
	}
	return products
}

func (p *pixel) parsingOrderFile(file string) (orders []*model.Receipt) {
	records, _ := p.convertCsvToStruct(file, 15)

	for i, record := range records {
		if i == 0 {
			continue
		}

		orders = append(orders, &model.Receipt{
			PharmacyID:      record[0],
			PharmacyAddress: record[1],
			Date:            record[2],
			KKM:             record[3],
			InvoiceNumber:   record[4],
			Manufacturer:    record[5],
			Supplier:        record[6],
			SupplierINN:     record[7],
			Name:            record[8],
			PriceWoVat:      record[9],
			PriceWVat:       record[10],
			Vat:             record[11],
			TotalPrice:      record[12],
			TotalNumber:     record[13],
			ShipmentNumber:  record[14],
			Series:          record[15],
		})
	}

	return orders
}

func (p *pixel) parsingProductFile(file string) (products []*model.Product) {
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
		products = append(products, &model.Product{
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
