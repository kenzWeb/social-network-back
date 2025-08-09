package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	g := gin.New()
	g.Use(gin.Logger(), gin.Recovery())

	v1 := g.Group("/api/v1")

	{
		v1.POST()
	}

	return g
}
