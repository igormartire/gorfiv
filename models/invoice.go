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
