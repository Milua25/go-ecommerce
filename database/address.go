package database

import (
	"context"
	"fmt"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func AddAddressByUserID(ctx context.Context, collection *mongo.Collection, primitiveId primitive.ObjectID, newAddress models.Address) error {
	// Define the filter and update for the MongoDB operation
	filter := bson.D{bson.E{Key: "$match", Value: bson.D{bson.E{Key: "_id", Value: primitiveId}}}}
	unwind := bson.D{{
		Key: "$unwind",
		Value: bson.D{
			bson.E{Key: "path", Value: "$address"},
		}}}
	group := bson.D{{
		Key: "$group",
		Value: bson.D{
			bson.E{Key: "_id", Value: "$address_id"},
			bson.E{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}}
	// addFields := bson.D{{
	// 	Key: "$addFields",
	// 	Value: bson.D{
	// 		bson.E{Key: "address", Value: bson.D{{Key: "$concatArrays", Value: bson.A{"$address", bson.A{newAddress}}}}},
	// 	}}}
	// run the aggregation pipeline
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{
		filter, unwind, group,
	})

	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		return err
	}

	// Check if the user has less than 3 addresses
	for _, result := range results {
		if count, ok := result["count"].(int32); ok && count >= 3 {
			return ErrAddressMaxReached // Return an error if the user already has 3 addresses
		}
	}

	filter = bson.D{bson.E{Key: "_id", Value: primitiveId}}
	update := bson.D{bson.E{Key: "$push", Value: bson.D{bson.E{Key: "address", Value: newAddress}}}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func UpdateHomeAddressByUserID(ctx context.Context, collection *mongo.Collection, primitiveId primitive.ObjectID, newAddress models.Address) error {
	return updateAddressByFixedIndex(ctx, collection, primitiveId, 0, newAddress)
}

func UpdateWorkAddressByUserID(ctx context.Context, collection *mongo.Collection, primitiveId primitive.ObjectID, newAddress models.Address) error {
	return updateAddressByFixedIndex(ctx, collection, primitiveId, 1, newAddress)
}

func updateAddressByFixedIndex(ctx context.Context, collection *mongo.Collection, primitiveId primitive.ObjectID, index int, newAddress models.Address) error {
	filter := bson.D{
		bson.E{Key: "_id", Value: primitiveId},
		bson.E{Key: fmt.Sprintf("address.%d", index), Value: bson.D{bson.E{Key: "$exists", Value: true}}},
	}
	setFields, err := buildAddressSetFields(index, newAddress)
	if err != nil {
		return err
	}

	update := bson.D{bson.E{Key: "$set", Value: setFields}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil

}
func buildAddressSetFields(index int, newAddress models.Address) (bson.D, error) {
	setFields := bson.D{}
	fieldPrefix := fmt.Sprintf("address.%d.", index)
	if newAddress.HouseNo != nil {
		setFields = append(setFields, bson.E{Key: fieldPrefix + "house_no", Value: *newAddress.HouseNo})
	}
	if newAddress.Street != nil {
		setFields = append(setFields, bson.E{Key: fieldPrefix + "street", Value: *newAddress.Street})
	}
	if newAddress.City != nil {
		setFields = append(setFields, bson.E{Key: fieldPrefix + "city", Value: *newAddress.City})
	}
	if newAddress.State != nil {
		setFields = append(setFields, bson.E{Key: fieldPrefix + "state", Value: *newAddress.State})
	}
	if newAddress.PostalCode != nil {
		setFields = append(setFields, bson.E{Key: fieldPrefix + "postal_code", Value: *newAddress.PostalCode})
	}
	if newAddress.Country != nil {
		setFields = append(setFields, bson.E{Key: fieldPrefix + "country", Value: *newAddress.Country})
	}

	if len(setFields) == 0 {
		return nil, ErrNoFieldsToUpdate
	}

	return setFields, nil
}
