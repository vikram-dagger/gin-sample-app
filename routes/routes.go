package routes

import (
	"book/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/books", controllers.CreateBook)
		api.GET("/books", controllers.GetBooks)
		api.GET("/books/:id", controllers.GetBookByID)
		api.DELETE("/books/:id", controllers.DeleteBook)
	}
}
