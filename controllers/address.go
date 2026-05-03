package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/milua25/e-commerce-backend/database"
	"github.com/milua25/e-commerce-backend/helpers"
	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (app *Application) CreateAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Query("id")
		if userId == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}
		primitiveId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		var newAddress CreateAddressInput
		if err := c.ShouldBindJSON(&newAddress); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = Validate.Struct(newAddress)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": helpers.ExtractValidationErrors(err)})
			return
		}

		newAddressModel := newAddress.ToModel()

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		err = database.AddAddressByUserID(ctx, app.userCollection, primitiveId, newAddressModel)
		if err != nil {
			if err == database.ErrAddressMaxReached {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Maximum number of addresses reached"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Address added successfully"})
	}

}

// func GetAddress() gin.HandlerFunc {}

// func UpdateAddress() gin.HandlerFunc {}

func (app *Application) DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for deleting an address goes here
		userId := c.Query("id")
		if userId == "" {
			c.JSON(400, gin.H{"error": "User ID is required"})
			return
		}
		addresses := make([]models.Address, 0)

		primitiveId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}
		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		filter := bson.D{bson.E{Key: "_id", Value: primitiveId}}

		update := bson.D{{
			Key: "$set",
			Value: bson.D{
				bson.E{Key: "address", Value: addresses}},
		}}

		err = database.DeleteAddressByUserID(ctx, app.userCollection, primitiveId, filter, update)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Address deleted successfully"})
	}
}

func (app *Application) UpdateHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for editing the home address goes here
		userId := c.Query("id")
		if userId == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return

		}
		// change id to primitive id
		primitiveId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}
		var updatedAddress UpdateAddressInput
		if err := c.ShouldBindJSON(&updatedAddress); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = Validate.Struct(updatedAddress)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": helpers.ExtractValidationErrors(err)})
			return
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		err = database.UpdateHomeAddressByUserID(ctx, app.userCollection, primitiveId, updatedAddress.ToModel())
		if err != nil {
			if err == database.ErrNoFieldsToUpdate {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No fields provided for update"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update home address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Home address updated successfully"})
	}
}

func (app *Application) UpdateWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for editing the work address goes here
		userId := c.Query("id")
		if userId == "" {
			log.Printf("User ID is required\n")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return

		}
		// change id to primitive id
		primitiveId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			log.Printf("Invalid user ID: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}
		var updatedAddress UpdateAddressInput
		if err := c.ShouldBindJSON(&updatedAddress); err != nil {
			log.Printf("Error binding JSON: %v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = Validate.Struct(updatedAddress)
		if err != nil {
			log.Printf("Validation error: %v\n", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": helpers.ExtractValidationErrors(err)})
			return
		}

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		err = database.UpdateWorkAddressByUserID(ctx, app.userCollection, primitiveId, updatedAddress.ToModel())
		if err != nil {
			log.Printf("Error updating work address: %v\n", err)
			if err == database.ErrNoFieldsToUpdate {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No fields provided for update"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update work address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Work address updated successfully"})
	}
}
