package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/milua25/e-commerce-backend/database"
	"github.com/milua25/e-commerce-backend/helpers"
	"github.com/milua25/e-commerce-backend/models"
	"github.com/milua25/e-commerce-backend/tokens"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var Validate = validator.New()

// SignUp handles user registration, including input validation, password hashing, and token generation
func (app *Application) SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sign-up logic goes here
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		var newUser CreateUserInput

		if err := c.BindJSON(&newUser); err != nil {
			log.Printf("Error binding JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request payload",
			})
			return
		}

		// Validate the input data
		err := Validate.Struct(newUser)
		if err != nil {
			log.Printf("Validation error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": helpers.ExtractValidationErrors(err),
			})
			return
		}
		// Check if a user with the same email already exists
		_, exists, err := database.FindUserByEmail(ctx, app.userCollection, newUser.Email)
		if err != nil {
			log.Printf("Error checking for existing user: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing user"})
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
			return
		}

		phoneExists, err := database.FindUserByPhone(ctx, app.userCollection, newUser.Phone)
		if err != nil {
			log.Printf("Error checking for existing phone: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing user"})
			return
		}
		if phoneExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User with this phone number already exists"})
			return
		}

		hashedPassword, err := helpers.HashPassword(newUser.Password)
		if err != nil {
			fmt.Printf("Error hashing password: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error processing password",
			})
			return
		}
		newUser.Password = hashedPassword
		newUser.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		newUser.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		newUser.ID = primitive.NewObjectID()
		newUser.UserID = newUser.ID.Hex()

		// token generation
		token, refreshToken, err := tokens.GenerateToken(newUser.Email, newUser.UserID, "24h", "168h", app.secretKey)
		if err != nil {
			fmt.Printf("Error generating token: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error generating token",
			})
			return
		}
		newUser.Token = &token
		newUser.RefreshToken = &refreshToken

		// Insert the new user into the database
		err = database.CreateUser(ctx, app.userCollection, newUser.ToModel())
		if err != nil {
			log.Printf("Error creating user: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error creating user",
			})
			return
		}

		// Prepare the response data
		UserResponse := UserResponse{
			ID:             newUser.ID,
			Username:       newUser.Username,
			Email:          newUser.Email,
			FirstName:      newUser.FirstName,
			LastName:       newUser.LastName,
			Phone:          &newUser.Phone,
			CreatedAt:      newUser.CreatedAt,
			UpdatedAt:      newUser.UpdatedAt,
			UserID:         newUser.UserID,
			AddressDetails: []AddressResponse{},
			UserCart:       []ProductUserResponse{},
			OrderStatus:    "No orders yet",
			Token:          &token,
			RefreshToken:   &refreshToken,
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"user":    UserResponse,
		})

	}
}

func (app *Application) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Login logic goes here
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		// Bind the JSON input to the LoginInput struct
		var loginInput LoginInput
		if err := c.BindJSON(&loginInput); err != nil {
			log.Printf("Error binding JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		// Validate the input data
		err := Validate.Struct(loginInput)
		if err != nil {
			log.Printf("Validation error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   helpers.ExtractValidationErrors(err),
				"message": "Validation failed",
			})
			return
		}

		var foundUser models.User
		foundUser, _, err = database.FindUserByEmail(ctx, app.userCollection, loginInput.Email)
		if err != nil {
			log.Printf("Error finding user: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
			return
		}
		// Verify password
		passwordValid, msg := helpers.VerifyPassword(foundUser.Password, loginInput.Password)
		if !passwordValid {
			log.Printf("Invalid password for user: %s\n", foundUser.Email)
			log.Printf("Password verification failed: %s\n", msg)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password",
			})
			return
		}

		// Generate new tokens
		token, refreshToken, err := tokens.GenerateToken(foundUser.Email, foundUser.UserID, "24h", "168h", app.secretKey)
		if err != nil {
			log.Printf("Error generating token: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			return
		}
		err = tokens.UpdateAllTokens(ctx, &token, &refreshToken, foundUser.UserID, app.userCollection)
		if err != nil {
			log.Printf("Error updating tokens: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Login successful",
			"token":         token,
			"refresh_token": refreshToken,
		})
	}
}

// Logout invalidates the user's tokens, effectively logging them out
func (app *Application) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Logout logic goes here
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		userId, ok := c.Get("uid")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		err := tokens.UpdateAllTokens(ctx, nil, nil, userId.(string), app.userCollection)
		if err != nil {
			log.Printf("Error clearing tokens: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
	}
}

// Update user details, such as name, email, or phone number
func (app *Application) UpdateUserDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		userQueryID := c.Param("id")
		if userQueryID == "" {
			log.Println("User ID query parameter is missing")
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		userId, ok := c.Get("uid")
		if !ok {
			log.Printf("User ID not found in context: %v\n", userId)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userValue, ok := userId.(string)
		if !ok {
			log.Printf("Error asserting user ID type: %v\n", userId)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Check if the user is trying to update their own details
		if userValue != userQueryID {
			log.Printf("User ID mismatch: token user ID %v does not match query user ID %v\n", userValue, userQueryID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only update your own details"})
			return
		}

		// Check if userId is empty string
		if userValue == "" {
			log.Printf("User ID from context is empty: %v\n", userId)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		primitiveUserID, err := primitive.ObjectIDFromHex(userValue)
		if err != nil {
			log.Printf("Error converting user ID to ObjectID: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// check if user exists
		user, exists, err := database.FindUserByID(ctx, app.userCollection, primitiveUserID)
		if err != nil {
			log.Printf("Error finding user: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			log.Printf("User not found: %v\n", userId)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Bind the JSON input to the UpdateUserInput struct
		var updateInput UpdateUserInput
		if err := c.BindJSON(&updateInput); err != nil {
			log.Printf("Error binding JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request payload",
			})
			return
		}

		// Validate the input data
		err = Validate.Struct(updateInput)
		if err != nil {
			log.Printf("Validation error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": helpers.ExtractValidationErrors(err),
			})
			return
		}

		if updateInput.Role != nil {
			// Check if the user is trying to update their role to a valid value
			if *updateInput.Role != "admin" && *updateInput.Role != "customer" {
				log.Printf("Invalid role value: %v\n", *updateInput.Role)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be either 'admin' or 'customer'"})
				return
			}
		}

		// Check if the user is trying to update their email to one that already exists
		if updateInput.Email != nil && *updateInput.Email != user.Email {
			_, exists, err := database.FindUserByEmail(ctx, app.userCollection, *updateInput.Email)
			if err != nil {
				log.Printf("Error checking for existing email: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing email"})
				return
			}
			if exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already in use"})
				return
			}
		}

		// Check if the user is trying to update their phone number to one that already exists
		if updateInput.Phone != nil && *updateInput.Phone != *user.Phone {
			phoneExists, err := database.FindUserByPhone(ctx, app.userCollection, *updateInput.Phone)
			if err != nil {
				log.Printf("Error checking for existing phone: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing phone"})
				return
			}
			if phoneExists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is already in use"})
				return
			}
		}

		if updateInput.FirstName == nil && updateInput.LastName == nil && updateInput.Phone == nil && updateInput.Email == nil && updateInput.Role == nil {
			log.Println("No fields to update")
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided for update"})
			return
		}

		user.FirstName = updateInput.FirstName
		user.LastName = updateInput.LastName
		user.Phone = updateInput.Phone
		if updateInput.Email != nil {
			user.Email = *updateInput.Email
		}
		if updateInput.Role != nil {
			user.Role = *updateInput.Role
		}
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		// update user details in the database
		err = database.UpdateUserDetails(ctx, app.userCollection, primitiveUserID, user)
		if err != nil {
			log.Printf("Error updating user details: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user details"})
			return
		}

		// Update user details logic goes here
		c.JSON(http.StatusOK, gin.H{"message": "Update user details endpoint"})
	}
}

// func ProductViewerAdmin() gin.HandlerFunc {}
func (app *Application) SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		products, err := database.FindAllProducts(ctx, app.productCollection)
		if err != nil {
			log.Printf("Error querying products: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please retry once again"})
		}

		productResponse := make([]ProductResponse, len(products))
		for i, p := range products {
			productResponse[i] = ProductResponse{
				ID:          p.ID,
				ProductName: p.ProductName,
				ProductID:   p.ProductID,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    p.Quantity,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
			}
		}
		c.JSON(http.StatusOK, productResponse)
	}
}

// SearchProductByQuery returns a list of products that match the search query
func (app *Application) SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for searching products by query goes here
		queryParam := c.Query("name")
		if queryParam == "" {
			log.Println("Search query is missing")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		searchProducts, err := database.FindProductByQuery(ctx, app.productCollection, queryParam)
		if err != nil {
			log.Printf("Error searching products: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please retry once again"})
		}

		searchProductResponses := make([]ProductResponse, len(searchProducts))
		for i, p := range searchProducts {
			searchProductResponses[i] = ProductResponse{
				ID:          p.ID,
				ProductName: p.ProductName,
				ProductID:   p.ProductID,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    p.Quantity,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
			}
		}

		log.Printf("Products matching query '%s' searched successfully", queryParam)
		c.JSON(http.StatusOK, searchProductResponses)
	}
}

func (app *Application) ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for viewing products goes here
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		user_id, ok := c.Get("uid")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		products, err := database.FindAllProducts(ctx, app.productCollection)
		if err != nil {
			log.Printf("Error querying products: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please retry once again"})
		}

		productResponse := make([]ProductResponse, len(products))
		for i, p := range products {
			productResponse[i] = ProductResponse{
				ID:          p.ID,
				ProductName: p.ProductName,
				ProductID:   p.ProductID,
				Description: p.Description,
				Price:       p.Price,
				Quantity:    p.Quantity,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
			}
		}
		log.Printf("Products viewed successfully by user %s", user_id)
		c.JSON(http.StatusOK, productResponse)

	}
}
