package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type SQLRepo struct {
	db *sql.DB
}

func NewSQLRepo(db *sql.DB) *SQLRepo {
	return &SQLRepo{db: db}
}

func (r *SQLRepo) GetInvoiceById(id int) (invoice *Invoice, err error) {
	invoice = &Invoice{}
	err = r.db.
		QueryRow("SELECT * FROM Invoice WHERE IsActive=1 AND Id=?;", id).
		Scan(&invoice.Id, &invoice.CreatedAt, &invoice.ReferenceMonth,
			&invoice.ReferenceYear, &invoice.Document, &invoice.Description,
			&invoice.Amount, &invoice.IsActive, &invoice.DeactiveAt)
	if err == sql.ErrNoRows {
		err = InvoiceNotFound
	}
	return
}

func (r *SQLRepo) UpdateInvoice(id int, newDescription string) (nRows int64, err error) {
	fmt.Println(id) //DEBUG
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
	fmt.Println(nRows)
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

func (r *SQLRepo) CountInvoices(opts *QueryOptions) (count int, err error) {
	err = r.db.QueryRow("SELECT COUNT(*) FROM Invoice WHERE IsActive=1" + r.QueryStringWithoutLimit(opts)).Scan(&count)
	return
}

func (r *SQLRepo) GetInvoices(opts *QueryOptions) (invoices []*Invoice, err error) {
	queryStr := "SELECT * FROM Invoice WHERE IsActive=1" + r.QueryString(opts)
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

func (r *SQLRepo) QueryString(q *QueryOptions) string {
	var queryStr bytes.Buffer
	for k, v := range q.Filters {
		queryStr.WriteString(" AND " + k + "=\"" + v + "\"")
	}

	if len(q.Sorts) > 0 {
		queryStr.WriteString(" ORDER BY ")
		var sortsStr = make([]string, len(q.Sorts))
		for i, sort := range q.Sorts {
			sortsStr[i] = r.SortToString(sort)
		}
		queryStr.WriteString(strings.Join(sortsStr, ", "))
	}

	var limitClauseStr = fmt.Sprint(" LIMIT ",
		(q.Pagination.Page-1)*q.Pagination.PerPage, ", ", q.Pagination.PerPage)
	queryStr.WriteString(limitClauseStr)

	return queryStr.String()
}

func (r *SQLRepo) QueryStringWithoutLimit(q *QueryOptions) string {
	var queryStr bytes.Buffer
	for k, v := range q.Filters {
		queryStr.WriteString(" AND " + k + "=\"" + v + "\"")
	}

	if len(q.Sorts) > 0 {
		queryStr.WriteString(" ORDER BY ")
		var sortsStr = make([]string, len(q.Sorts))
		for i, sort := range q.Sorts {
			sortsStr[i] = r.SortToString(sort)
		}
		queryStr.WriteString(strings.Join(sortsStr, ", "))
	}

	return queryStr.String()
}

func (*SQLRepo) SortToString(s Sort) (str string) {
	str = s.Field
	if s.Desc {
		str += " DESC"
	} else {
		str += " ASC"
	}
	return
}
