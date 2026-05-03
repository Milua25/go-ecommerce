package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/milua25/e-commerce-backend/database"
	"github.com/milua25/e-commerce-backend/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (app *Application) AddProductToDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for adding products goes here
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		var newProduct CreateProductInput
		if err := c.BindJSON(&newProduct); err != nil {
			log.Printf("Error binding JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request payload",
			})
			return
		}

		err := Validate.Struct(newProduct)
		if err != nil {
			log.Printf("Validation error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": helpers.ExtractValidationErrors(err),
			})
			return
		}

		uidValue, ok := c.Get("uid")
		log.Printf("User ID from context: %v\n", uidValue)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userId, ok := uidValue.(string)
		if !ok || userId == "" {
			log.Printf("Error asserting user ID type: %v\n", uidValue)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		primitivUserID, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			log.Printf("Error converting user ID: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		// check if user is admin
		isAdmin, err := database.IsUserAdmin(ctx, app.userCollection, primitivUserID)
		if err != nil {
			log.Printf("Error checking user role: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
			return
		}

		err = database.CreateProduct(ctx, app.productCollection, newProduct.ToModel())
		if err != nil {
			log.Printf("Error creating product: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error creating product",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Product created successfully",
			"product": ProductResponse{
				ID:          newProduct.ID,
				ProductName: newProduct.ProductName,
				ProductID:   newProduct.ProductID,
				Description: newProduct.Description,
				Price:       newProduct.Price,
				Quantity:    newProduct.Quantity,
				CreatedAt:   newProduct.CreatedAt,
				UpdatedAt:   newProduct.UpdatedAt,
			},
		})

	}
}
