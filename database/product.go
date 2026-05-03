package database

import (
	"context"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func FindAllProducts(ctx context.Context, collection *mongo.Collection) ([]models.Product, error) {
	var productList []models.Product

	cursor, err := collection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &productList)
	if err != nil {
		return nil, err
	}

	return productList, nil
}

func FindProductByQuery(ctx context.Context, collection *mongo.Collection, query string) ([]models.Product, error) {
	var searchProducts []models.Product

	cursor, err := collection.Find(ctx, bson.M{"product_name": bson.M{"$regex": query, "$options": "i"}})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &searchProducts)
	if err != nil {
		return nil, err
	}

	return searchProducts, nil
}

func CreateProduct(ctx context.Context, productCollection *mongo.Collection, product models.Product) error {
	_, err := productCollection.InsertOne(ctx, product)
	return err
}
