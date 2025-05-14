package main

import (
	"book/database"
	"book/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	database.ConnectDatabase()
	routes.RegisterRoutes(r)
	r.Run(":8080")
}
