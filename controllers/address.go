package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/milua25/e-commerce-backend/database"
	"github.com/milua25/e-commerce-backend/helpers"
	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (app *Application) CreateAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("id")
		if userId == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}
		primitiveId, err := bson.ObjectIDFromHex(userId)
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
		log.Printf("New address model: %+v\n", newAddressModel)

		ctx, cancel := requestContext(c, DefaultTimeout)
		defer cancel()

		err = app.store.UserStoreCollection.AddAddressByUserID(ctx, primitiveId, newAddressModel)
		if err != nil {
			if err == database.ErrAddressMaxReached {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Maximum number of addresses reached"})
				return
			}
			log.Printf("Error adding address: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Address added successfully"})
	}

}

func (app *Application) DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for deleting an address goes here
		userId := c.Param("id")
		if userId == "" {
			c.JSON(400, gin.H{"error": "User ID is required"})
			return
		}
		addresses := make([]models.Address, 0)

		primitiveId, err := bson.ObjectIDFromHex(userId)
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

		err = app.store.UserStoreCollection.DeleteAddressByUserID(ctx, primitiveId, filter, update)
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
		primitiveId, err := bson.ObjectIDFromHex(userId)
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

		err = app.store.UserStoreCollection.UpdateHomeAddressByUserID(ctx, primitiveId, updatedAddress.ToModel())
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
		primitiveId, err := bson.ObjectIDFromHex(userId)
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

		err = app.store.UserStoreCollection.UpdateWorkAddressByUserID(ctx, primitiveId, updatedAddress.ToModel())
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
