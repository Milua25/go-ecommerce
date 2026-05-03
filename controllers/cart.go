package controllers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/milua25/e-commerce-backend/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var DefaultTimeout = 100 * time.Second

type Application struct {
	productCollection *mongo.Collection
	userCollection    *mongo.Collection
	secretKey         string
	// cartCollection    *mongo.Collection
}

func NewApplication(productCol, userCol *mongo.Collection, secretKey string) *Application {
	return &Application{
		productCollection: productCol,
		userCollection:    userCol,
		secretKey:         secretKey,
		// cartCollection:    cartCol,
	}
}

func (app *Application) AddProductToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for adding a product to the cart goes here
		productQueryID := c.Query("id")
		if productQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("Product ID is required"), Type: gin.ErrorTypePublic})
			return
		}

		userQueryID := c.Query("user_id")
		if userQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("User ID is required"), Type: gin.ErrorTypePublic})
			return
		}

		productId, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}
		userId, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		// Check if the product exists
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		product, err := database.GetProductForCart(ctx, app.productCollection, productId)
		if err != nil {
			log.Printf("Error fetching product: %v\n", err)
			if err == database.ErrProductNotFound {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Product not found"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Call the database function to add the product to the cart
		err = database.AddProductToCart(ctx, app.userCollection, userId, product)
		if err != nil {
			log.Printf("Error adding product to cart: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product added to cart successfully"})
	}
}

func (app *Application) ViewCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for adding a product to the cart goes here
		productQueryID := c.Query("id")
		if productQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("Product ID is required"), Type: gin.ErrorTypePublic})
			return
		}

		userQueryID := c.Query("user_id")
		if userQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("User ID is required"), Type: gin.ErrorTypePublic})
			return
		}
		//
		productId, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}
		userId, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		// Call the database function to view the product in the cart
		err = database.ViewProductCart(ctx, app.productCollection, productId, userId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product viewed in cart successfully"})
	}
}

func (app *Application) RemoveItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for removing a product from the cart goes here
		productQueryID := c.Query("id")
		if productQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("Product ID is required"), Type: gin.ErrorTypePublic})
			return
		}

		userQueryID := c.Query("user_id")
		if userQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("User ID is required"), Type: gin.ErrorTypePublic})
			return
		}

		productId, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}
		userId, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()
		// Call the database function to remove the product from the cart
		err = database.RemoveProductFromCart(ctx, app.productCollection, userId, productId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product removed from cart successfully"})
	}
}

func (app *Application) ClearCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for adding a product to the cart goes here
		// productQueryID := c.Query("id")
		// if productQueryID == "" {
		// 	_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("Product ID is required"), Type: gin.ErrorTypePublic})
		// 	return
		// }

		userQueryID := c.Query("user_id")
		if userQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("User ID is required"), Type: gin.ErrorTypePublic})
			return
		}

		// productId, err := primitive.ObjectIDFromHex(productQueryID)
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		// 	return
		// }

		userId, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()
		// Call the database function to clear the cart
		err = database.ClearCart(ctx, app.productCollection, userId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for adding a product to the cart goes here
		productQueryID := c.Query("id")
		if productQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("Product ID is required"), Type: gin.ErrorTypePublic})
			return
		}

		userQueryID := c.Query("user_id")
		if userQueryID == "" {
			_ = c.AbortWithError(http.StatusBadRequest, gin.Error{Err: errors.New("User ID is required"), Type: gin.ErrorTypePublic})
			return
		}

		productId, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}
		userId, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()
		// Call the database function to add the product to the cart
		err = database.InstantBuy(ctx, app.productCollection, userId, productId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product purchased successfully"})
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for adding a product to the cart goes here
		userQueryID := c.Query("user_id")
		if userQueryID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()
		// userId, err := primitive.ObjectIDFromHex(userQueryID)
		userId, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		err = database.BuyItemFromCart(ctx, app.userCollection, userId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Items purchased successfully"})
	}
}

func (app *Application) GetItemsInCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")

		// check if user_id is provided
		if user_id == "" {
			log.Println("User ID is missing in the request")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		// Convert user_id to ObjectID
		userId, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			log.Printf("Error converting user ID to ObjectID: %v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		// find the user with the filled cart items
		filledCartItems, err := database.FindUserWithFilledCart(ctx, app.userCollection, userId)
		if err != nil {
			log.Printf("Error finding user with filled cart: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// aggregate the cart items
		databaseItems, err := database.AggregateCartItems(ctx, app.userCollection, userId)
		if err != nil {
			log.Printf("Error aggregating cart items: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, item := range databaseItems {
			log.Printf("Cart item: %v\n", item)
			if itemMap, ok := item.(map[string]interface{}); ok {
				c.JSON(http.StatusOK, itemMap["total"])
			}
			c.JSON(http.StatusOK, filledCartItems.UserCart)

		}

	}
}
