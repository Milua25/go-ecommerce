package database

import (
	"context"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Storage struct {
	// MongoDB client and collections can be added here
	CartStoreCollection interface {
		GetProductForCart(ctx context.Context, collection *mongo.Collection, productId primitive.ObjectID) (models.ProductUser, error)
		AddProductToCart(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID, model models.ProductUser) error
		RemoveProductFromCart(ctx context.Context, userCollection *mongo.Collection, userId, productId primitive.ObjectID) error
		GetCartItems(ctx context.Context, userCollection *mongo.Collection, userId primitive.ObjectID) ([]models.ProductUser, error)
		BuyCartItems(ctx context.Context, userCollection *mongo.Collection, userId primitive.ObjectID) error
		InstantBuyFromCart(ctx context.Context, userCollection *mongo.Collection, userId primitive.ObjectID) error
	}
	// Other collections and methods can be added here
}

// func NewStorage() *Storage {
// 	return &Storage{
// 		CartStoreCollection: &CartStore{},
// 	}
// }
