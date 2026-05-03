package database

import (
	"context"
	"errors"
	"time"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrProductNotFound   = errors.New("Product not found")
	ErrCartNotFound      = errors.New("Cart not found")
	ErrUserIdNotFound    = errors.New("User ID not found")
	ErrInvalidProductID  = errors.New("Invalid product ID")
	ErrInvalidUserID     = errors.New("Invalid user ID")
	ErrCartEmpty         = errors.New("Cart is empty")
	ErrCantBuyCartItems  = errors.New("Cannot buy item in the cart")
	ErrInsufficientStock = errors.New("Insufficient stock for the product")
	ErrUnauthorized      = errors.New("Unauthorized access")
	ErrInternalServer    = errors.New("Internal server error")
	ErrAddressMaxReached = errors.New("Maximum number of addresses reached")
	ErrNoFieldsToUpdate  = errors.New("no fields provided for update")
	ErrCantUpdateUser    = errors.New("cannot update user with provided data")
	ErrCantUserorder     = errors.New("cannot place order for the user")
)

type CartStore struct {
}

func GetProductForCart(ctx context.Context, collection *mongo.Collection, productId primitive.ObjectID) (models.ProductUser, error) {
	var product models.ProductUser
	err := collection.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: productId}}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.ProductUser{}, ErrProductNotFound
		}
		return models.ProductUser{}, err
	}
	return product, nil
}

func AddProductToCart(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID, model models.ProductUser) error {
	filter := bson.D{bson.E{Key: "_id", Value: userId}}
	update := bson.D{bson.E{Key: "$push", Value: bson.D{bson.E{Key: "user_cart", Value: bson.D{bson.E{Key: "$each", Value: bson.A{model}}}}}}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantUpdateUser
	}
	if result.MatchedCount == 0 {
		return ErrUserIdNotFound
	}
	return nil
}

func RemoveProductFromCart(ctx context.Context, userCollection *mongo.Collection, userId, productId primitive.ObjectID) error {

	// Implementation for removing a product from the cart goes here
	filter := bson.D{bson.E{Key: "_id", Value: userId}}
	update := bson.D{bson.E{Key: "$pull", Value: bson.D{bson.E{Key: "user_cart", Value: bson.D{bson.E{Key: "_id", Value: productId}}}}}}

	result, err := userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return ErrCantUpdateUser
	}
	if result.MatchedCount == 0 {
		return ErrUserIdNotFound
	}
	return nil
}

func GetCartContents(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID) error {
	return nil
}

func ClearCart(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID) error {
	return nil
}

func InstantBuy(ctx context.Context, collection *mongo.Collection, userId, productId primitive.ObjectID) error {
	return nil
}

func BuyItemFromCart(ctx context.Context, usercollection *mongo.Collection, userId primitive.ObjectID) error {
	// Implementation for buying items from the cart goes here
	// fetch the user's cart, check stock for each item, process payment, and update inventory accordingly
	var getcartItems models.User
	var ordercart models.Order

	ordercart.ID = primitive.NewObjectID()
	ordercart.OrderedAt = time.Now()
	ordercart.OrderCart = make([]models.ProductUser, 0)
	ordercart.PaymentMethod.COD = true

	// Fetch the user's cart
	unwind := bson.D{bson.E{Key: "$unwind", Value: bson.D{bson.E{Key: "path", Value: "$user_cart"}}}}
	grouping := bson.D{bson.E{Key: "$group", Value: bson.D{
		bson.E{Key: "_id", Value: "$_id"},
		bson.E{Key: "total", Value: bson.D{bson.E{Key: "$sum", Value: "$user_cart.price"}}},
	}}}

	cursor, err := usercollection.Aggregate(ctx, mongo.Pipeline{
		unwind, grouping,
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var getusercar []bson.M
	if err := cursor.All(ctx, &getusercar); err != nil {
		return err
	}
	// Calculate total price and prepare order cart
	var totalPrice int32
	for _, item := range getusercar {
		price := item["total"].(int32)
		totalPrice = price
	}
	ordercart.Price = uint64(totalPrice)

	filter := bson.D{bson.E{Key: "_id", Value: userId}}
	update := bson.D{bson.E{Key: "$push", Value: bson.D{bson.E{Key: "orders", Value: ordercart}}}}

	result, err := usercollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return ErrCantUserorder
	}
	if result.MatchedCount == 0 {
		return ErrUserIdNotFound
	}

	err = usercollection.FindOne(ctx, filter).Decode(&getcartItems)
	if err != nil {
		return err
	}

	filter2 := bson.D{bson.E{Key: "_id", Value: userId}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": bson.M{"$each": getcartItems.UserCart}}}

	_, err = usercollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		return ErrCantUserorder
	}

	// Clear the user's cart after successful order placement
	clearCart := bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: "user_cart", Value: []models.ProductUser{}}}}}
	_, err = usercollection.UpdateMany(ctx, filter, clearCart)
	if err != nil {
		return ErrCantUpdateUser
	}

	return nil
}

func FindUserWithFilledCart(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID) (models.User, error) {

	result := collection.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: userId}})
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return models.User{}, ErrUserIdNotFound
		}
		return models.User{}, result.Err()
	}

	var user models.User
	err := result.Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	if len(user.UserCart) == 0 {
		return models.User{}, ErrCartEmpty
	}

	return user, nil
}

func AggregateCartItems(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID) (bson.M, error) {
	// Define the aggregation pipeline
	filter_match := bson.D{bson.E{Key: "$match", Value: bson.D{bson.E{Key: "_id", Value: userId}}}}
	unwind_cart := bson.D{bson.E{Key: "$unwind", Value: "$userCart"}}
	grouping := bson.D{bson.E{Key: "$group", Value: bson.D{
		bson.E{Key: "_id", Value: "$_id"},
		bson.E{Key: "total", Value: bson.D{bson.E{Key: "$sum", Value: "$userCart.price"}}},
		bson.E{Key: "cartItems", Value: bson.D{bson.E{Key: "$push", Value: "$userCart"}}},
	}}}

	// Execute the aggregation pipeline
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind_cart, grouping})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result bson.M

	// var results []struct {
	// 	ID    primitive.ObjectID `bson:"_id"`
	// 	Total float64            `bson:"total"`
	// }

	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, ErrCartNotFound
	}

	return result, nil
}

func ViewProductCart(ctx context.Context, collection *mongo.Collection, productId primitive.ObjectID, userId primitive.ObjectID) error {
	return nil
}

func InstantBuyFromCart(ctx context.Context, productCollection, userCollection *mongo.Collection, productId, userId primitive.ObjectID) error {
	var productDetails models.ProductUser
	var ordercart models.Order

	ordercart.ID = primitive.NewObjectID()
	ordercart.OrderedAt = time.Now()
	ordercart.OrderCart = make([]models.ProductUser, 0)
	ordercart.PaymentMethod.COD = true

	// Fetch the product details
	err := productCollection.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: productId}}).Decode(&productDetails)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrProductNotFound
		}
		return err
	}
	ordercart.Price = productDetails.Price
	ordercart.OrderCart = append(ordercart.OrderCart, productDetails)

	// Process the order placement logic here (e.g., update inventory, process payment, etc.)
	filter := bson.D{bson.E{Key: "_id", Value: productId}}
	update := bson.D{bson.E{Key: "$inc", Value: bson.D{bson.E{Key: "quantity", Value: -1}}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// add the order to the user's order history
	orderFilter := bson.D{bson.E{Key: "_id", Value: userId}}
	orderUpdate := bson.D{bson.E{Key: "$push", Value: bson.D{bson.E{Key: "orders", Value: ordercart}}}}

	result, err := userCollection.UpdateOne(ctx, orderFilter, orderUpdate)
	if err != nil {
		return ErrCantUserorder
	}
	if result.MatchedCount == 0 {
		return ErrUserIdNotFound
	}

	return nil
}
