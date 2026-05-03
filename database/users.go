package database

import (
	"context"
	"errors"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func UserCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("ecommerce").Collection(collectionName)
}

func FindUserByID(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID) (models.User, bool, error) {
	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return user, false, nil
	}
	if err != nil {
		return user, false, err
	}
	return user, true, nil
}

func FindUserByEmail(ctx context.Context, collection *mongo.Collection, email string) (models.User, bool, error) {
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return user, false, nil
	}
	if err != nil {
		return user, false, err
	}
	return user, true, nil
}

func FindUserByPhone(ctx context.Context, collection *mongo.Collection, phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}
	var result bson.D
	err := collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeleteAddressByUserID(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID, filter bson.D, update bson.D) error {
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("no user found with the given ID")
	}
	return nil
}

func CreateUser(ctx context.Context, collection *mongo.Collection, user models.User) error {
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func IsUserAdmin(ctx context.Context, userCollection *mongo.Collection, userId primitive.ObjectID) (bool, error) {
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return false, err
	}
	return user.Role == models.UserRoleAdmin, nil
}

func UpdateUserDetails(ctx context.Context, userCollection *mongo.Collection, userId primitive.ObjectID, update models.User) error {

	result, err := userCollection.UpdateOne(ctx, bson.M{"_id": userId}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("no user found with the given ID")
	}
	return nil
}
