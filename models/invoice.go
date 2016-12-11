package models

import (
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	DOCUMENT_MAX_LENGTH = 14
)

type Invoice struct {
	Id             int            `json:"id"`
	CreatedAt      time.Time      `json:"createdAt"`
	ReferenceMonth int            `json:"referenceMonth"`
	ReferenceYear  int            `json:"referenceYear"`
	Document       string         `json:"document"`
	Description    string         `json:"description"`
	Amount         float64        `json:"amount"`
	IsActive       bool           `json:"isActive"`
	DeactiveAt     mysql.NullTime `json:"deactiveAt"`
}

func (i1 *Invoice) Equals(i2 *Invoice) bool {
	return i1.Id == i2.Id &&
		i1.CreatedAt == i2.CreatedAt &&
		i1.ReferenceMonth == i2.ReferenceMonth &&
		i1.ReferenceYear == i2.ReferenceYear &&
		i1.Document == i2.Document &&
		i1.Description == i2.Description &&
		i1.Amount == i2.Amount &&
		i1.IsActive == i2.IsActive &&
		i1.DeactiveAt.Valid == i2.DeactiveAt.Valid &&
		!(i1.DeactiveAt.Valid && (i1.DeactiveAt.Time != i2.DeactiveAt.Time))
}
