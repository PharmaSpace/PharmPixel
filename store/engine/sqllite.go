package engine

import (
	"Pixel/store"
	"database/sql"
	"fmt"
	log "github.com/go-pkgz/lgr"
	"github.com/hashicorp/go-multierror"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"strconv"
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(fileName string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make sqllite for %s", fileName)
	}
	result := SQLiteDB{db: db}
	log.Printf("[DEBUG] sqllite store created for %s", fileName)
	return &result, nil
}

func (s *SQLiteDB) Create(file store.File) (fileName store.File, err error) {
	result, err := s.db.Exec("insert into file (name) values ($1)", file.Name)
	if err != nil {
		return fileName, err
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		return fileName, err
	}
	fileName.ID = strconv.FormatInt(lastId, 10)
	return fileName, nil
}

func (s *SQLiteDB) CreateProduct(product store.Product) (pr store.Product, err error) {
	var name string
	cRaw := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='products_new';")
	err = cRaw.Scan(&name)
	if err == sql.ErrNoRows {
		_, err := s.db.Exec("create table products_new(id integer primary key autoincrement unique, pharmspaceid varchar(255),name varchar(255), manufacture varchar(255), export integer(1), timestamp timestamp);", nil)
		if err != nil {
			return pr, err
		}
	}

	row := s.db.QueryRow("select pharmspaceid, name, manufacture, export from products_new where name=$1", product.Name)
	err = row.Scan(&pr.ID, &pr.Name, &pr.Manufacture, &pr.Export)
	if err != nil && err == sql.ErrNoRows {
		result, err := s.db.Exec("insert into products_new (pharmspaceid, name,manufacture, export) values ($1, $2, $3, $4)", product.ID, product.Name, product.Manufacture, product.Export)
		if err != nil {
			return pr, err
		}
		lastId, err := result.LastInsertId()
		if err != nil {
			return pr, err
		}
		pr.ID = strconv.FormatInt(lastId, 10)
	} else if err == nil {
		result, err := s.db.Exec("update products_new set export=$1 where id=$2", product.Export, product.Name)
		if err != nil {
			return pr, err
		}
		lastId, err := result.LastInsertId()
		if err != nil {
			return pr, err
		}
		pr.ID = strconv.FormatInt(lastId, 10)
		pr.Export = product.Export
	}
	return pr, err
}

func (s *SQLiteDB) CreateReceipt(receipt store.Receipt) (rReceipt store.Receipt, err error) {
	row := s.db.QueryRow("select createdAt, dateTime, fiscalDocumentNumber, inn, kktRegId, link, name,ofd, price, priceSellIn, productId, quantity, total,totalSum,updatedAt from receipt where fiscalDocumentNumber=$1", receipt.FiscalDocumentNumber)
	err = row.Scan(
		&rReceipt.CreatedAt,
		&rReceipt.DateTime,
		&rReceipt.FiscalDocumentNumber,
		&rReceipt.Inn,
		&rReceipt.KktRegId,
		&rReceipt.Link,
		&rReceipt.Name,
		&rReceipt.Ofd,
		&rReceipt.Price,
		&rReceipt.PriceSellIn,
		&rReceipt.ProductId,
		&rReceipt.Quantity,
		&rReceipt.Total,
		&rReceipt.TotalSum,
		&rReceipt.UpdatedAt,
	)
	if err != nil && err == sql.ErrNoRows {
		_, err = s.db.Exec("insert into receipt (createdAt, dateTime, fiscalDocumentNumber, inn, kktRegId, link, name,ofd, price, priceSellIn, productId, quantity, total,totalSum,updatedAt) values ($1, $2, $3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)",
			receipt.CreatedAt,
			receipt.DateTime,
			receipt.FiscalDocumentNumber,
			receipt.Inn,
			receipt.KktRegId,
			receipt.Link,
			receipt.Name,
			receipt.Ofd,
			receipt.Price,
			receipt.PriceSellIn,
			receipt.ProductId,
			receipt.Quantity,
			receipt.Total,
			receipt.TotalSum,
			receipt.UpdatedAt,
		)
	}

	if err != nil {
		return rReceipt, err
	}
	return rReceipt, err
}

func (s *SQLiteDB) CreateReceiptN(receipt store.ReceiptN) (rReceipt store.ReceiptN, err error) {
	_, err = s.db.Exec("insert into receiptn (datePay, manufacture, name, number, pointName, priceSellIn, priceSellOut, productId,provider,quantity,supplerName) values ($1, $2, $3,$4,$5,$6,$7,$8,$9,$10,$11)",
		receipt.DatePay,
		receipt.Manufacture,
		receipt.Name,
		receipt.Number,
		receipt.PointName,
		receipt.PriceSellIn,
		receipt.PriceSellOut,
		receipt.ProductId,
		receipt.Provider,
		receipt.Quantity,
		receipt.SupplerName,
	)
	if err != nil {
		return rReceipt, err
	}
	return rReceipt, err
}

func (s *SQLiteDB) Get(req GetRequest) (file store.File, err error) {
	stmt, err := s.db.Prepare("select id, name from file where name=?")
	if err != nil {
		return file, err
	}
	defer stmt.Close()
	f := store.File{}
	err = stmt.QueryRow(req.Name).Scan(&f.ID, &f.Name)
	if err != nil {
		return file, err
	}
	return f, nil
}

func (s *SQLiteDB) GetProduct(req GetRequest) (product store.Product, err error) {
	stmt, err := s.db.Prepare("select pharmspaceid, name, manufacture, export from products_new where name=?")
	if err != nil {
		return product, err
	}
	defer stmt.Close()
	p := store.Product{}
	err = stmt.QueryRow(req.Name).Scan(&p.ID, &p.Name, &p.Manufacture, &p.Export)
	if err != nil {
		return product, err
	}
	return p, nil
}

func (s *SQLiteDB) GetAllReceipt() (receipts []store.Receipt, err error) {
	rs, err := s.db.Query("select * from receipt")
	if err != nil {
		return receipts, err
	}

	for rs.Next() {
		r := store.Receipt{}
		err := rs.Scan(
			&r.CreatedAt,
			&r.DateTime,
			&r.FiscalDocumentNumber,
			&r.Inn,
			&r.KktRegId,
			&r.Link,
			&r.Name,
			&r.Ofd,
			&r.Price,
			&r.PriceSellIn,
			&r.ProductId,
			&r.Quantity,
			&r.Total,
			&r.TotalSum,
			&r.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		receipts = append(receipts, r)
	}
	return receipts, err
}

func (s *SQLiteDB) DeleteReceipt(receipt store.Receipt) error {
	_, err := s.db.Query("delete from receipt where fiscalDocumentNumber=$1 and kktRegId=$2", receipt.FiscalDocumentNumber, receipt.KktRegId)
	if err != nil {
		return err
	}
	return nil
}

// Close boltdb store
func (s *SQLiteDB) Close() error {
	errs := new(multierror.Error)
	err := errors.Wrapf(s.db.Close(), "can't close")
	errs = multierror.Append(errs, err)
	return errs.ErrorOrNil()
}
