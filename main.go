package main

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
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

const (
	RootPath = "localhost:3000"
)

func main() {
	db, err := sql.Open("mysql",
		"stone:password@/Stone?collation=utf8_general_ci&parseTime=true")
	checkErr(err)
	defer db.Close()

	// make sure connection is available
	err = db.Ping()
	checkErr(err)

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/invoices")
	})

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
			invoice           Invoice
			invoices          []Invoice
			orderByStrs       []string
			queryStr          string
			filterClauseStr   string
			numResultsPerPage int = 5
			pageNumber        int = 1
			linksHeader       []string
			numActiveInvoices int
		)

		dbErr := db.QueryRow("SELECT COUNT(*) FROM Invoice WHERE IsActive=1").Scan(&numActiveInvoices)
		if dbErr != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		perPageStr, isSet := c.GetQuery("perPage")
		if isSet {
			perPageVal, err := strconv.Atoi(perPageStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Malformed perPage parameter. It must be an integer.",
				})
				return
			}
			numResultsPerPage = perPageVal
		}

		lastPageNumber := int(math.Ceil(float64(numActiveInvoices) / float64(numResultsPerPage)))

		pageNumStr, isSet := c.GetQuery("page")
		if isSet {
			pageNumVal, err := strconv.Atoi(pageNumStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Malformed page parameter. It must be an integer.",
				})
				return
			}
			if pageNumVal < 1 || pageNumVal > lastPageNumber {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid page number passed as parameter.",
				})
				return
			}
			pageNumber = pageNumVal
		}

		linkPrefix := "<" + RootPath + "/invoices?"
		values := c.Request.URL.Query()
		if pageNumber < lastPageNumber {
			//next
			values.Set("page", strconv.Itoa(pageNumber+1))
			fmt.Println(values.Encode())
			linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"next\"")
			//last
			values.Set("page", strconv.Itoa(lastPageNumber))
			fmt.Println(values.Encode())
			linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"last\"")
		}
		if pageNumber > 1 {
			//first
			values.Set("page", "1")
			linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"first\"")
			//prev
			values.Set("page", strconv.Itoa(pageNumber-1))
			fmt.Println(values.Encode())
			linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"prev\"")
		}

		documentFilter, filteringDocument := c.GetQuery("document")
		if filteringDocument {
			filterClauseStr += " AND Document=\"" + documentFilter + "\""
		}
		monthFilter, filteringMonth := c.GetQuery("referenceMonth")
		if filteringMonth {
			filterClauseStr += " AND ReferenceMonth=\"" + monthFilter + "\""
		}
		yearFilter, filteringYear := c.GetQuery("referenceYear")
		if filteringYear {
			filterClauseStr += " AND ReferenceYear=\"" + yearFilter + "\""
		}

		sortQuery, sorting := c.GetQuery("sort")
		if sorting {
			if sortQuery == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "malformed sort query",
				})
				return
			}
			fields := strings.Split(sortQuery, ",")
			for _, field := range fields {
				sort_order := "ASC"
				if len(field) == 0 {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "malformed sort query",
					})
					return
				}
				if field[0] == '-' {
					sort_order = "DESC"
					field = field[1:]
				}
				if field == "document" || field == "referenceMonth" || field == "referenceYear" {
					orderByStrs = append(orderByStrs, fmt.Sprintf("%v %v", field, sort_order))
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "malformed sort query: can only order by document, referenceMonth and referenceYear",
					})
					return
				}
			}
		}

		queryStr = "SELECT * FROM Invoice WHERE IsActive=1" + filterClauseStr
		if len(orderByStrs) > 0 {
			queryStr += " ORDER BY " + strings.Join(orderByStrs, ", ")
		}
		queryStr += fmt.Sprint(" LIMIT ", (pageNumber-1)*numResultsPerPage, ", ", numResultsPerPage)

		fmt.Println(queryStr)

		rows, err := db.Query(queryStr)

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

		c.Header("X-Total-Count", strconv.Itoa(numActiveInvoices))
		c.Header("Link", strings.Join(linksHeader, ", "))
		c.JSON(http.StatusOK, gin.H{"items": invoices})
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

		c.Header("Location", fmt.Sprint(RootPath, "/invoices/", id))
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
