package database

import (
	"context"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProductStoreCollection interface {
	FindAllProducts(ctx context.Context) ([]models.Product, error)
	FindProductByQuery(ctx context.Context, query string) ([]models.Product, error)
	CreateProduct(ctx context.Context, product models.Product) error
}

type CartStoreCollection interface {
	GetProductForCart(ctx context.Context, productId bson.ObjectID) (models.ProductUser, error)
	AddProductToCart(ctx context.Context, userId bson.ObjectID, model models.ProductUser) error
	RemoveProductFromCart(ctx context.Context, userId, productId bson.ObjectID) error
	GetCartContents(ctx context.Context, userId bson.ObjectID) error
	BuyItemFromCart(ctx context.Context, userId bson.ObjectID) error
	InstantBuyFromCart(ctx context.Context, productId, userId bson.ObjectID) error
	FindUserWithFilledCart(ctx context.Context, userId bson.ObjectID) (models.User, error)
	AggregateCartItems(ctx context.Context, userId bson.ObjectID) (bson.M, error)
	ViewProductCart(ctx context.Context, productId bson.ObjectID, userId bson.ObjectID) error
	ClearCart(ctx context.Context, userId bson.ObjectID) error
}

type UserStoreCollection interface {
	FindUserByID(ctx context.Context, userId bson.ObjectID) (models.User, bool, error)
	FindUserByEmail(ctx context.Context, email string) (models.User, bool, error)
	FindUserByPhone(ctx context.Context, phone string) (bool, error)
	DeleteAddressByUserID(ctx context.Context, userId bson.ObjectID, filter bson.D, update bson.D) error
	CreateUser(ctx context.Context, user models.User) error
	IsUserAdmin(ctx context.Context, userId bson.ObjectID) (bool, error)
	UpdateUserDetails(ctx context.Context, userId bson.ObjectID, update models.User) error
	CountUsers(ctx context.Context) (int64, error)
	AddAddressByUserID(ctx context.Context, userId bson.ObjectID, newAddress models.Address) error
	UpdateHomeAddressByUserID(ctx context.Context, userId bson.ObjectID, newAddress models.Address) error
	UpdateWorkAddressByUserID(ctx context.Context, userId bson.ObjectID, newAddress models.Address) error
}

type Storage struct {
	// productCollection      *mongo.Collection
	// userCollection         *mongo.Collection
	ProductStoreCollection ProductStoreCollection
	CartStoreCollection    CartStoreCollection
	UserStoreCollection    UserStoreCollection
}

func NewStorage(productCol, userCol *mongo.Collection) *Storage {
	return &Storage{
		// productCollection:      productCol,
		// userCollection:         userCol,
		ProductStoreCollection: &ProductStore{productCollection: productCol},
		CartStoreCollection:    &CartStore{productCollection: productCol, userCollection: userCol},
		UserStoreCollection:    &UserStore{userCollection: userCol},
	}
}

// need a tx wrapper for cart operations that involve both cart and product collections to ensure atomicity and consistency
func withTransaction(sess mongo.Session, ctx context.Context, fn func(context.Context) error) error {
	err := mongo.WithSession(ctx, &sess, func(sessCtx context.Context) error {
		if err := sess.StartTransaction(); err != nil {
			return err
		}

		// Execute the provided callback within the transaction
		if err := fn(sessCtx); err != nil {
			if abortErr := sess.AbortTransaction(sessCtx); abortErr != nil {
				return abortErr
			}
			return err
		}

		if err := sess.CommitTransaction(sessCtx); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}
