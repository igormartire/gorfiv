package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Run(env *Env, apiToken string) {
	router := gin.Default()
	router.HandleMethodNotAllowed = true

	authorized := router.Group("/", tokenAuthMiddleware(apiToken))

	authorized.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/invoices")
	})

	authorized.GET("/invoices", prepareQueryOptions, env.invoicesIndex)
	authorized.GET("/invoices/:id", env.invoicesShow)
	authorized.POST("/invoices", env.invoicesPost)
	authorized.PUT("/invoices/:id", validatePostFormMiddleware, env.invoicesPut)
	authorized.DELETE("/invoices/:id", env.invoicesDelete)

	router.Run("localhost:3000")
}
