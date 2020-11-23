package format

import (
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
	"pixel/config"
	"pixel/core/model"
	"pixel/store/service"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Pixel структура
type Pixel struct {
	date             time.Time
	Config           *config.Config
	MP               service.MarketPlaceInterface
	Log              serviceLib.Logger
	matchingOfdCache *cache.Cache
	matchingErpCache *cache.Cache
	cache            *cache.Cache
}

// GetCache получение кеша по ключу
func (p *Pixel) GetCache(key string) (interface{}, bool) {
	if p.cache == nil {
		return nil, false
	}
	return p.cache.Get(key)
}

// GetOFDCache получение кеша по ключу
func (p *Pixel) GetOFDCache(key string) (interface{}, bool) {
	if p.matchingOfdCache == nil {
		return nil, false
	}
	return p.matchingOfdCache.Get(key)
}

// GetERPCache получение кеша по ключу
func (p *Pixel) GetERPCache(key string) (interface{}, bool) {
	if p.matchingErpCache == nil {
		return nil, false
	}
	return p.matchingErpCache.Get(key)
}

// SetOFDCache получение кеша по ключу
func (p *Pixel) SetOFDCache(key string, val interface{}, duration time.Duration) {
	if p.matchingOfdCache == nil {
		return
	}
	p.matchingOfdCache.Set(key, val, duration)
}

// SetERPCache установка кеша по ключу
func (p *Pixel) SetERPCache(key string, val interface{}, duration time.Duration) {
	if p.matchingErpCache == nil {
		return
	}
	p.matchingOfdCache.Set(key, val, duration)
}

// GetMP получение инстанса для маркетплейса
func (p *Pixel) GetMP() service.MarketPlaceInterface {
	return p.MP
}

// GetDate получение даты
func (p *Pixel) GetDate() time.Time {
	return p.date
}

// NewPixel формат Пикселя
func NewPixel(cf *config.Config, mp service.MarketPlaceInterface, ch *cache.Cache, l serviceLib.Logger, date time.Time) *Pixel {
	matchingOfdCache := cache.New(5*time.Minute, 10*time.Minute)
	matchingErpCache := cache.New(5*time.Minute, 10*time.Minute)

	return &Pixel{Config: cf, cache: ch, MP: mp, Log: l, matchingErpCache: matchingErpCache, matchingOfdCache: matchingOfdCache, date: date}
}

// Parse парсинг
func (p *Pixel) Parse() {
	p.matchingErpCache.Flush()
	p.matchingOfdCache.Flush()
	getMatchProducts(p)
	var err error

	// данные из кеша по чекам из ОФД
	ofdRecieptCacheItems := p.cache.Items()
	// чеки из ЕРП
	erpOrder, erpProducts := p.getFiles()

	receipts, checkOfdProductNames := nameMatching(p, erpOrder, ofdRecieptCacheItems)
	if len(receipts) > 0 {
		err = p.MP.SendReceipt(receipts)
		if err != nil {
			log.Printf("[ERROR] Ошикбка отправкии чеков")
		}
	}

	if len(checkOfdProductNames) > 0 {
		// отправляем товары из ofd на матчинг
		productsForMatching := OfdProductsForMatching(checkOfdProductNames)
		err = p.MP.SendOfdProducts(productsForMatching, true, false)
		if err != nil {
			log.Printf("[ERROR] Ошикбка отправкии продуктов")
		}
	}
	// отправляем товары из erp на матчинг
	productsForMatching := ErpProductsForMatching(p, p.convertProducts(erpProducts))
	err = p.MP.SendOfdProducts(productsForMatching, false, true)
	if err != nil {
		log.Printf("[ERROR] Ошикбка отправкии продуктов")
	}
}

func (p *Pixel) getFiles() (receipts []*model.Receipt, products []*model.Product) {
	files := p.readDir(p.Config.Files.SourceFolder)
	for _, file := range files {
		zipCompile := regexp.MustCompile(".zip")
		if zipCompile.Match([]byte(file.Name())) {
			date := p.date.AddDate(0, 0, p.Config.FormatDate)
			fileZip, _ := regexp.MatchString(date.Format("20060102"), file.Name())
			if fileZip {
				p.unzip(fmt.Sprintf("%s/%s", p.Config.Files.SourceFolder, file.Name()), p.Config.Files.WorkingFolder)
				r, pr := p.parseFiles()
				receipts = append(receipts, r...)
				products = append(products, pr...)
				if p.Config.Files.BackupFolder != "" {
					err := os.Rename(fmt.Sprintf("%s/%s", p.Config.Files.SourceFolder, file.Name()), fmt.Sprintf("%s/%s", p.Config.Files.BackupFolder, file.Name()))
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}
	return receipts, products
}

func (p *Pixel) parseFiles() ([]*model.Receipt, []*model.Product) {
	filesUnzip := p.readDir(p.Config.Files.WorkingFolder)
	AllOrders := make([]*model.Receipt, 0)
	AllProducts := make([]*model.Product, 0)
	for _, file := range filesUnzip {
		productCompile := regexp.MustCompile("products.csv")
		orderCompile := regexp.MustCompile("orders.csv")
		filePath := fmt.Sprintf("%s/%s", p.Config.Files.WorkingFolder, file.Name())
		if productCompile.Match([]byte(file.Name())) {
			products := p.parsingProductFile(filePath)
			AllProducts = append(AllProducts, products...)
			if err := os.Remove(filePath); err != nil {
				err = p.Log.Errorf("%v", err)
				if err != nil {
					log.Printf("[ERROR] Ошибка удаления файла продуктов%v", err)
				}
			}
		}
		if orderCompile.Match([]byte(file.Name())) {
			orders := p.parsingOrderFile(filePath)
			AllOrders = append(AllOrders, orders...)
			if err := os.Remove(filePath); err != nil {
				err = p.Log.Errorf("%v", err)
				if err != nil {
					log.Printf("[ERROR] Ошибка удаления файла чеков %v", err)
				}
			}
		}
	}
	return AllOrders, AllProducts
}

// convertProducts - конвертирует продукты для отправки на матчинг
func (p *Pixel) convertProducts(record []*model.Product) []service.Product {
	products := []service.Product{}

	for _, record := range record {
		var product service.Product
		product.Name = record.Name
		product.Manufacturer = record.Manufacturer
		product.PartNumber = record.ShipmentNumber
		product.Serial = record.Series
		product.Stock = record.Inventory
		product.WarehouseName = record.PharmacyID
		product.SupplerName = record.Supplier
		product.SupplerInn = record.SupplierINN
		product.CreatedAt = p.date.Format(time.RFC3339Nano)
		products = append(products, product)
	}
	return products
}

func (p *Pixel) parsingOrderFile(file string) (orders []*model.Receipt) {
	records, _ := p.convertCsvToStruct(file, 16)

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

func (p *Pixel) parsingProductFile(file string) (products []*model.Product) {
	records, _ := p.convertCsvToStruct(file, 11)
	for i, record := range records {
		if i == 0 {
			continue
		}
		record[9] = strings.Replace(record[9], ",", ".", -1)
		quantity, err := strconv.ParseFloat(record[9], 64)
		if err != nil {
			err = p.Log.Errorf("convert quantity string to float64: %v", err)
			if err != nil {
				log.Printf("convert quantity string to float64: %v", err)
			}
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

func (p *Pixel) convertCsvToStruct(file string, l int) (outRecord [][]string, err error) {
	readFile, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		err = p.Log.Errorf("Error read file: %v", err)
		if err != nil {
			log.Printf("Error read file: %v", err)
		}
	}
	sourceFile := string(readFile)
	sourceFile = strings.ReplaceAll(sourceFile, "\r", "")
	// лайфхак для лидерфармы
	sourceFile = strings.ReplaceAll(sourceFile, "\",\"", ";")
	sourceFile = strings.ReplaceAll(sourceFile, "\"", "")

	fileIO := strings.NewReader(sourceFile)
	r := csv.NewReader(fileIO)
	r.Comma = ';'
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	for i, record := range records {
		lenRecord := len(record)
		if lenRecord <= 1 {
			continue
		}
		if lenRecord < l {
			err = p.Log.Errorf("%s: не совпадает количество значений в строке. Ожидаемое количество: %v Пришло: %v Строка: %v", file, l, lenRecord, i)
			if err != nil {
				log.Printf("%s: не совпадает количество значений в строке. Ожидаемое количество: %v Пришло: %v Строка: %v", file, l, lenRecord, i)
			}
			continue
		}
		outRecord = append(outRecord, record)
	}
	return outRecord, err
}

func (p *Pixel) readDir(dir string) (files []os.FileInfo) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		err = p.Log.Errorf("[ERR] Ошибка чтения директории %s| %v", dir, err)
		if err != nil {
			log.Printf("[ERR] Ошибка чтения директории %s| %v", dir, err)
		}
	}
	return files
}

func (p *Pixel) unzip(fileName, targetDir string) {
	zipReader, _ := zip.OpenReader(fileName)
	for _, file := range zipReader.Reader.File {

		zippedFile, err := file.Open()
		if err != nil {
			err = p.Log.Errorf("[ERR] Ошибка чтения файла | %v", err)
			if err != nil {
				log.Printf("[ERR] Ошибка чтения файла | %v", err)
				return
			}
			return
		}

		// #nosec G305
		extractedFilePath := filepath.Join(targetDir, file.Name)
		if !strings.HasPrefix(extractedFilePath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			log.Printf("%s: illegal file path", extractedFilePath)
			return
		}

		if file.FileInfo().IsDir() {
			err = os.MkdirAll(extractedFilePath, file.Mode())
			if err != nil {
				log.Printf("[ERR] Ошибка оздания файлов | %v", err)
				return
			}
		} else {
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				err = p.Log.Errorf("[ERR] Ошибка распаковки файла | %v", err)
				if err != nil {
					log.Printf("[ERR] Ошибка распаковки файла | %v", err)
				}
			}
			// #nosec G110
			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				err = p.Log.Errorf("[ERR] Ошибка копирование содержимого архива | %v", err)
				if err != nil {
					log.Printf("[ERR] Ошибка копирование содержимого архива | %v", err)
				}
			}
		}
	}
	defer zipReader.Close()
}
