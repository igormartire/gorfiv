package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func New(env *Env, apiToken string) *gin.Engine {
	router := gin.Default()
	router.HandleMethodNotAllowed = true

	authorized := router.Group("/", tokenAuthMiddleware(apiToken))

	authorized.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/invoices")
	})

	authorized.GET("/invoices", prepareQueryOptions, env.invoicesIndex)
	authorized.GET("/invoices/:id", env.invoicesShow)
	authorized.POST("/invoices", validatePostFormMiddleware, env.invoicesPost)
	authorized.PUT("/invoices/:id", env.invoicesPut)
	authorized.DELETE("/invoices/:id", env.invoicesDelete)

	return router
}
