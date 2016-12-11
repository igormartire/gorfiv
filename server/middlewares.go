package server

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/igormartire/gorfiv/models"
)

var documentMaxLengthErrorMsg = "parameter document cannot have length greater than " + strconv.Itoa(models.DOCUMENT_MAX_LENGTH) + " characters"

func respondWithError(c *gin.Context, code int, errorMsg string) {
	c.JSON(code, gin.H{"error": errorMsg})
	c.Abort()
}

func tokenAuthMiddleware(apiToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userToken := c.Request.FormValue("apiToken")

		if userToken == "" {
			respondWithError(c, http.StatusUnauthorized, "API token required")
			return
		}

		if userToken != apiToken {
			respondWithError(c, http.StatusUnauthorized, "Invalid API token")
			return
		}

		c.Next()
	}
}

func validatePostFormMiddleware(c *gin.Context) {
	document := c.PostForm("document")
	_, err := strconv.ParseFloat(c.PostForm("amount"), 64)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "amount parameter must be specified and must be a number")
		return
	}

	if document == "" {
		respondWithError(c, http.StatusBadRequest, "missing or empty document parameter")
		return
	}

	if utf8.RuneCountInString(document) > models.DOCUMENT_MAX_LENGTH {
		respondWithError(c, http.StatusBadRequest, documentMaxLengthErrorMsg)
		return
	}

	c.Next()
}

func prepareQueryOptions(c *gin.Context) {
	values := c.Request.Form
	errors := validateFormValuesForQueryOptions(values)
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": errors,
		})
		c.Abort()
	}

	var opts = models.QueryOptions{
		Filters: map[string]string{},
		Sorts:   []models.Sort{},
		Pagination: models.Pagination{
			Page:    1,
			PerPage: 5,
		},
	}

	for k, v := range values {
		switch k {
		case "document", "referenceMonth", "referenceYear":
			if len(v) == 1 {
				opts.Filters[k] = v[0]
			}
		case "page":
			if len(v) == 1 {
				opts.Pagination.Page, _ = strconv.Atoi(v[0])
			}
		case "perPage":
			if len(v) == 1 {
				opts.Pagination.PerPage, _ = strconv.Atoi(v[0])
			}
		case "sort":
			if len(v) == 1 {
				fields := strings.Split(v[0], ",")
				sorts := make([]models.Sort, len(fields))
				for i, field := range fields {
					if field[0] == '-' {
						sorts[i].Desc = true
						field = field[1:]
					}
					sorts[i].Field = field
				}
				opts.Sorts = sorts
			}
		}
	}

	c.Set("QueryOptions", &opts)
	c.Next()
}

func validateFormValuesForQueryOptions(values url.Values) (errors []string) {
	for k, v := range values {
		switch k {
		case "document", "referenceMonth", "referenceYear", "sort", "apiToken", "page", "perPage":
			if len(v) > 1 {
				errors = append(errors, "duplicate parameter "+k)
			}
		default:
			errors = append(errors, "invalid parameter "+k)
		}

		switch k {
		case "document":
			for _, value := range v {
				if utf8.RuneCountInString(value) > models.DOCUMENT_MAX_LENGTH {
					errors = append(errors, documentMaxLengthErrorMsg)
					break
				}
			}
		case "referenceMonth", "referenceYear", "page", "perPage":
			for _, value := range v {
				if _, err := strconv.Atoi(value); err != nil {
					errors = append(errors, "parameter "+k+" must be an integer")
					break
				}
			}
		case "sort":
			for _, value := range v {
				fields := strings.Split(value, ",")
				for _, f := range fields {
					if f == "" {
						errors = append(errors, "malformed sort query")
						continue
					}
					if f[0] == '-' {
						f = f[1:]
					}
					if f != "document" && f != "referenceMonth" && f != "referenceYear" {
						errors = append(errors, "malformed sort query. Correct syntax: sort=[-](document|referenceMonth|referenceYear)[,...]")
					}
				}
			}
		}
	}
	return
}
