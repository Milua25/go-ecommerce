package routes

// This file defines the routes for the application.
import (
	"github.com/gin-gonic/gin"
	"github.com/milua25/e-commerce-backend/controllers"
)

func UserRoutes(router gin.IRouter, app *controllers.Application) {
	router.POST("/users/signup", app.SignUp())
	router.POST("/users/login", app.Login())
	// router.POST("/admin/addproduct", app.AddProduct())
	// router.GET("/users/productview", app.GetProducts())
	router.GET("/users/search", app.SearchProductByQuery())
}
