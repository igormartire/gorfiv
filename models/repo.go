package models

import (
	"errors"
	"math"
)

type Repo interface {
	GetInvoices(opts *QueryOptions) (invoices []*Invoice, err error)
	GetInvoiceById(id int) (*Invoice, error)
	InsertInvoice(i Invoice) (id int64, err error)
	DeleteInvoice(id int) (nRows int64, err error)
	UpdateInvoice(id int, newDescription string) (nRows int64, err error)
	CountInvoices(opts *QueryOptions) (count int, err error)
}

var InvoiceNotFound = errors.New("id not found")

type QueryOptions struct {
	Filters    map[string]string
	Sorts      []Sort
	Pagination Pagination
}

type Sort struct {
	Field string
	Desc  bool
}

type Pagination struct {
	Page    int
	PerPage int
}

func (p Pagination) LastPageNumber(numActiveRecords int) (lastPageNumber int) {
	lastPageNumber = int(math.Ceil(float64(numActiveRecords) / float64(p.PerPage)))
	return
}
