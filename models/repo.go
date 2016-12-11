package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

type Repo interface {
	GetInvoices(opts *QueryOptions) ([]*Invoice, error)
	GetInvoiceById(id int) (Invoice, error)
	InsertInvoice(i Invoice) (id int64, err error)
	DeleteInvoice(id int) (nRows int64, err error)
	UpdateInvoice(id int, newDescription string) (nRows int64, err error)
	CountInvoices() (count int, err error)
}

type SQLRepo struct {
	db *sql.DB
}

func NewSQLRepo(db *sql.DB) *SQLRepo {
	return &SQLRepo{db: db}
}

func (r *SQLRepo) GetInvoiceById(id int) (invoice Invoice, err error) {
	err = r.db.
		QueryRow("SELECT * FROM Invoice WHERE IsActive=1 AND Id=?;", id).
		Scan(&invoice.Id, &invoice.CreatedAt, &invoice.ReferenceMonth,
			&invoice.ReferenceYear, &invoice.Document, &invoice.Description,
			&invoice.Amount, &invoice.IsActive, &invoice.DeactiveAt)
	return
}

func (r *SQLRepo) UpdateInvoice(id int, newDescription string) (nRows int64, err error) {
	stmt, err := r.db.Prepare("UPDATE Invoice SET Description=? WHERE IsActive=1 AND Id=?")
	if err != nil {
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(newDescription, id)
	if err != nil {
		return
	}

	nRows, err = res.RowsAffected()
	return
}

func (r *SQLRepo) DeleteInvoice(id int) (nRows int64, err error) {
	stmt, err := r.db.Prepare("UPDATE Invoice SET IsActive=0, DeactiveAt=? WHERE IsActive=1 AND Id=?")
	if err != nil {
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(time.Now(), id)
	if err != nil {
		return
	}

	nRows, err = res.RowsAffected()
	return
}

func (r *SQLRepo) InsertInvoice(i Invoice) (id int64, err error) {
	stmt, err := r.db.Prepare(`INSERT INTO Invoice SET
	                           CreatedAt=?, ReferenceMonth=?, ReferenceYear=?,
	                           Document=?, Description=?, Amount=?,
	                           IsActive=?, DeactiveAt=?`)
	if err != nil {
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.CreatedAt, i.ReferenceMonth, i.ReferenceYear,
		i.Document, i.Description, i.Amount, i.IsActive, nil)
	if err != nil {
		return
	}

	id, err = res.LastInsertId()
	return
}

func (r *SQLRepo) CountInvoices() (numActiveRecords int, err error) {
	err = r.db.QueryRow("SELECT COUNT(*) FROM Invoice WHERE IsActive=1").Scan(&numActiveRecords)
	return
}

type Pagination struct {
	Page    int
	PerPage int
}

func (p Pagination) LastPageNumber(numActiveRecords int) (lastPageNumber int) {
	lastPageNumber = int(math.Ceil(float64(numActiveRecords) / float64(p.PerPage)))
	return
}

type Sort struct {
	Field string
	Desc  bool
}

func (s Sort) String() (str string) {
	str = s.Field
	if s.Desc {
		str += " DESC"
	} else {
		str += " ASC"
	}
	return
}

type QueryOptions struct {
	Filters    map[string]string
	Sorts      []Sort
	Pagination Pagination
}

func (q *QueryOptions) QueryString() string {
	var queryStr bytes.Buffer
	for k, v := range q.Filters {
		queryStr.WriteString(" AND " + k + "=\"" + v + "\"")
	}

	if len(q.Sorts) > 0 {
		queryStr.WriteString(" ORDER BY ")
		var sortsStr = make([]string, len(q.Sorts))
		for i, sort := range q.Sorts {
			sortsStr[i] = sort.String()
		}
		queryStr.WriteString(strings.Join(sortsStr, ", "))
	}

	var limitClauseStr = fmt.Sprint(" LIMIT ",
		(q.Pagination.Page-1)*q.Pagination.PerPage, ", ", q.Pagination.PerPage)
	queryStr.WriteString(limitClauseStr)

	return queryStr.String()
}

func (r *SQLRepo) GetInvoices(opts *QueryOptions) (invoices []*Invoice, err error) {
	queryStr := "SELECT * FROM Invoice WHERE IsActive=1" + opts.QueryString()

	rows, err := r.db.Query(queryStr)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var invoice Invoice
		err = rows.Scan(&invoice.Id, &invoice.CreatedAt, &invoice.ReferenceMonth,
			&invoice.ReferenceYear, &invoice.Document, &invoice.Description,
			&invoice.Amount, &invoice.IsActive, &invoice.DeactiveAt)
		if err != nil {
			return
		}
		invoices = append(invoices, &invoice)
	}

	err = rows.Err()
	return
}
