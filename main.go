package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type Invoice struct {
	Id             int            `json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	ReferenceMonth int            `json:"reference_month"`
	ReferenceYear  int            `json:"reference_year"`
	Document       string         `json:"document"`
	Description    string         `json:"description"`
	Amount         float64        `json:"amount"`
	IsActive       bool           `json:"is_active"`
	DeactiveAt     mysql.NullTime `json:"deactive_at"`
}

func main() {
	db, err := sql.Open("mysql",
		"stone:password@/Stone?collation=utf8_general_ci&parseTime=true")
	checkErr(err)
	defer db.Close()

	// make sure connection is available
	err = db.Ping()
	checkErr(err)

	router := gin.Default()

	router.GET("/invoices/:id", func(c *gin.Context) {
		var invoice Invoice

		id := c.Param("id")
		row := db.QueryRow("SELECT * FROM Invoice WHERE Id=? AND IsActive=1;", id)
		err = row.Scan(&invoice.Id, &invoice.CreatedAt, &invoice.ReferenceMonth,
			&invoice.ReferenceYear, &invoice.Document, &invoice.Description,
			&invoice.Amount, &invoice.IsActive, &invoice.DeactiveAt)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "there is no resource with the specified id",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"entity": invoice,
			})
		}
	})

	router.GET("/invoices", func(c *gin.Context) {
		var (
			invoice  Invoice
			invoices []Invoice
		)

		rows, err := db.Query("SELECT * FROM Invoice WHERE IsActive=1")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&invoice.Id, &invoice.CreatedAt, &invoice.ReferenceMonth,
				&invoice.ReferenceYear, &invoice.Document, &invoice.Description,
				&invoice.Amount, &invoice.IsActive, &invoice.DeactiveAt)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			invoices = append(invoices, invoice)
		}

		if rows.Err() != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Header("X-Total-Count", strconv.Itoa(len(invoices)))
		c.JSON(http.StatusOK, gin.H{"collection": invoices})
	})

	router.POST("/invoices", func(c *gin.Context) {
		document := c.PostForm("document")
		description := c.PostForm("description")
		amount, err := strconv.ParseFloat(c.PostForm("amount"), 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "amount parameter must be specified and must be a number",
			})
			return
		}

		if document == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing or empty document parameter",
			})
			return
		}

		if utf8.RuneCountInString(document) > 14 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "parameter document cannot have length greater than 14 characters",
			})
			return
		}

		createdAt := time.Now()
		referenceMonth := int(time.Now().Month())
		referenceYear := time.Now().Year()
		isActive := true

		stmt, err := db.Prepare(`INSERT INTO Invoice SET
                             CreatedAt=?, ReferenceMonth=?, ReferenceYear=?,
                             Document=?, Description=?, Amount=?,
                             IsActive=?, DeactiveAt=?`)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer stmt.Close()

		res, err := stmt.Exec(createdAt, referenceMonth, referenceYear,
			document, description, amount, isActive, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Header("Location", fmt.Sprint("localhost:3000/invoices", id))
		c.Status(http.StatusCreated)
	})

	router.DELETE("/invoices/:id", func(c *gin.Context) {
		id := c.Param("id")
		stmt, err := db.Prepare("UPDATE Invoice SET IsActive=0, DeactiveAt=? WHERE Id=? AND IsActive=1")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer stmt.Close()

		res, err := stmt.Exec(time.Now(), id)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		nRows, err := res.RowsAffected()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if nRows == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "there is no resource with the specified id",
			})
		} else {
			c.Status(http.StatusNoContent)
		}
	})

	router.PUT("/invoices/:id", func(c *gin.Context) {
		id := c.Param("id")
		description, exist := c.GetPostForm("description")
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "paramater description must be specified",
			})
			return
		}

		stmt, err := db.Prepare("UPDATE Invoice SET Description=? WHERE Id=?")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer stmt.Close()

		res, err := stmt.Exec(description, id)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		nRows, err := res.RowsAffected()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if nRows == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "there is no resource with the specified id",
			})
		} else {
			c.Status(http.StatusNoContent)
		}
	})

	router.Run(":3000")
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
