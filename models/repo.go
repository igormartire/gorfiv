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

// var (
// 	invoice           models.Invoice
// 	orderByStrs       []string
// 	queryStr          string
// 	filterClauseStr   string
// 	numResultsPerPage int = 5
// 	pageNumber        int = 1
// 	linksHeader       []string
// 	numActiveInvoices int
// )
//
// dbErr := db.QueryRow("SELECT COUNT(*) FROM Invoice WHERE IsActive=1").Scan(&numActiveInvoices)
// if dbErr != nil {
// 	c.AbortWithError(http.StatusInternalServerError, err)
// 	return
// }
//
// perPageStr, isSet := c.GetQuery("perPage")
// if isSet {
// 	perPageVal, err := strconv.Atoi(perPageStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Malformed perPage parameter. It must be an integer.",
// 		})
// 		return
// 	}
// 	numResultsPerPage = perPageVal
// }
//
// lastPageNumber := int(math.Ceil(float64(numActiveInvoices) / float64(numResultsPerPage)))
//
// pageNumStr, isSet := c.GetQuery("page")
// if isSet {
// 	pageNumVal, err := strconv.Atoi(pageNumStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Malformed page parameter. It must be an integer.",
// 		})
// 		return
// 	}
// 	if pageNumVal < 1 || pageNumVal > lastPageNumber {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid page number passed as parameter.",
// 		})
// 		return
// 	}
// 	pageNumber = pageNumVal
// }
//
// linkPrefix := "<" + c.Request.Host + "/invoices?"
// values := c.Request.URL.Query()
// if pageNumber < lastPageNumber {
// 	//next
// 	values.Set("page", strconv.Itoa(pageNumber+1))
// 	linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"next\"")
// 	//last
// 	values.Set("page", strconv.Itoa(lastPageNumber))
// 	linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"last\"")
// }
// if pageNumber > 1 {
// 	//first
// 	values.Set("page", "1")
// 	linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"first\"")
// 	//prev
// 	values.Set("page", strconv.Itoa(pageNumber-1))
// 	linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"prev\"")
// }
//
// documentFilter, filteringDocument := c.GetQuery("document")
// if filteringDocument {
// 	filterClauseStr += " AND Document=\"" + documentFilter + "\""
// }
// monthFilter, filteringMonth := c.GetQuery("referenceMonth")
// if filteringMonth {
// 	filterClauseStr += " AND ReferenceMonth=\"" + monthFilter + "\""
// }
// yearFilter, filteringYear := c.GetQuery("referenceYear")
// if filteringYear {
// 	filterClauseStr += " AND ReferenceYear=\"" + yearFilter + "\""
// }
//
// sortQuery, sorting := c.GetQuery("sort")
// if sorting {
// 	if sortQuery == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "malformed sort query",
// 		})
// 		return
// 	}
// 	fields := strings.Split(sortQuery, ",")
// 	for _, field := range fields {
// 		sort_order := "ASC"
// 		if len(field) == 0 {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "malformed sort query",
// 			})
// 			return
// 		}
// 		if field[0] == '-' {
// 			sort_order = "DESC"
// 			field = field[1:]
// 		}
// 		if field == "document" || field == "referenceMonth" || field == "referenceYear" {
// 			orderByStrs = append(orderByStrs, fmt.Sprintf("%v %v", field, sort_order))
// 		} else {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "malformed sort query: can only order by document, referenceMonth and referenceYear",
// 			})
// 			return
// 		}
// 	}
// }
