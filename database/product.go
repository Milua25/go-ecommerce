package database

import (
	"context"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProductStore struct {
	productCollection *mongo.Collection
}

func NewProductStore(collection *mongo.Collection) *ProductStore {
	return &ProductStore{productCollection: collection}
}

func (p *ProductStore) FindAllProducts(ctx context.Context) ([]models.Product, error) {
	var productList []models.Product

	cursor, err := p.productCollection.Find(ctx, bson.M{})

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

func (p *ProductStore) FindProductByQuery(ctx context.Context, query string) ([]models.Product, error) {
	var searchProducts []models.Product

	cursor, err := p.productCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": query, "$options": "i"}})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &searchProducts)
	if err != nil {
		return nil, err
	}

	return searchProducts, nil
}

func (p *ProductStore) CreateProduct(ctx context.Context, product models.Product) error {
	_, err := p.productCollection.InsertOne(ctx, product)
	return err
}
