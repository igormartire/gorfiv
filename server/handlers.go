package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/igormartire/gorfiv/models"
)

type Env struct {
	repo models.Repo
}

func NewEnv(r models.Repo) *Env {
	return &Env{repo: r}
}

func (env *Env) invoicesShow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parameter id should be an integer",
		})
		return
	}

	invoice, err := env.repo.GetInvoiceById(id)

	if err != nil {
		if err == models.InvoiceNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "there is no resource with the specified id",
			})
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"item": invoice,
		})
	}
}

func (env *Env) invoicesDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parameter id should be an integer",
		})
		return
	}

	nRows, err := env.repo.DeleteInvoice(id)
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
}

func (env *Env) invoicesPut(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parameter id should be an integer",
		})
		return
	}

	newDescription, exist := c.GetPostForm("description")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "paramater description must be specified",
		})
		return
	}

	_, err = env.repo.UpdateInvoice(id, newDescription)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (env *Env) invoicesPost(c *gin.Context) {
	amount, _ := strconv.ParseFloat(c.PostForm("amount"), 64) //err already checked in middleware

	id, err := env.repo.InsertInvoice(models.Invoice{
		Document:       c.PostForm("document"),
		Description:    c.PostForm("description"),
		Amount:         amount,
		CreatedAt:      time.Now(),
		ReferenceMonth: int(time.Now().Month()),
		ReferenceYear:  time.Now().Year(),
		IsActive:       true,
	})

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Location", fmt.Sprint(c.Request.Host, "/invoices/", id))
	c.Status(http.StatusCreated)
}

func (env *Env) invoicesIndex(c *gin.Context) {
	getValue, exist := c.Get("QueryOptions")
	if !exist {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Couldn't get QueryOptions @ handlers.invoicesIndex"))
		return
	}

	opts := getValue.(*models.QueryOptions)

	totalCount, err := env.repo.CountInvoices(opts)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if totalCount == 0 {
		c.Header("X-Total-Count", "0")
		c.JSON(http.StatusOK, gin.H{"items": []*models.Invoice{}})
		return
	}

	var lastPageNumber = opts.Pagination.LastPageNumber(totalCount)
	if opts.Pagination.Page < 1 || opts.Pagination.Page > lastPageNumber {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid page number passed as parameter.",
		})
		return
	}

	invoices, err := env.repo.GetInvoices(opts)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var linksHeader []string
	linkPrefix := "<" + c.Request.Host + "/invoices?"
	values := c.Request.URL.Query()
	if opts.Pagination.Page < lastPageNumber {
		//next
		values.Set("page", strconv.Itoa(opts.Pagination.Page+1))
		linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"next\"")
		//last
		values.Set("page", strconv.Itoa(lastPageNumber))
		linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"last\"")
	}
	if opts.Pagination.Page > 1 {
		//first
		values.Set("page", "1")
		linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"first\"")
		//prev
		values.Set("page", strconv.Itoa(opts.Pagination.Page-1))
		linksHeader = append(linksHeader, linkPrefix+values.Encode()+">; rel=\"prev\"")
	}

	c.Header("X-Total-Count", strconv.Itoa(totalCount))
	c.Header("Link", strings.Join(linksHeader, ", "))
	c.JSON(http.StatusOK, gin.H{"items": invoices})
}
