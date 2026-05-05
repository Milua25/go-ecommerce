package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/milua25/e-commerce-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserStore struct {
	userCollection *mongo.Collection
}

func NewUserStore(collection *mongo.Collection) *UserStore {
	return &UserStore{userCollection: collection}
}

func UserCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("ecommerce").Collection(collectionName)
}

func (u *UserStore) FindUserByID(ctx context.Context, userId bson.ObjectID) (models.User, bool, error) {
	var user models.User
	err := u.userCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return user, false, nil
	}
	if err != nil {
		return user, false, err
	}
	return user, true, nil
}

func (u *UserStore) FindUserByEmail(ctx context.Context, email string) (models.User, bool, error) {
	var user models.User

	err := u.userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return user, false, nil
	}
	if err != nil {
		return user, false, err
	}
	return user, true, nil
}

func (u *UserStore) FindUserByPhone(ctx context.Context, phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}
	var result bson.D
	err := u.userCollection.FindOne(ctx, bson.M{"phone": phone}).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *UserStore) DeleteAddressByUserID(ctx context.Context, userId bson.ObjectID, filter bson.D, update bson.D) error {
	result, err := u.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("no user found with the given ID")
	}
	return nil
}

func (u *UserStore) CreateUser(ctx context.Context, user models.User) error {
	_, err := u.userCollection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserStore) IsUserAdmin(ctx context.Context, userId bson.ObjectID) (bool, error) {
	var user models.User
	err := u.userCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return false, err
	}
	return user.Role == models.UserRoleAdmin, nil
}

func (u *UserStore) UpdateUserDetails(ctx context.Context, userId bson.ObjectID, update models.User) error {

	result, err := u.userCollection.UpdateOne(ctx, bson.M{"_id": userId}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("no user found with the given ID")
	}
	return nil
}

func (u *UserStore) CountUsers(ctx context.Context) (int64, error) {
	opts := options.Count().SetHint("_id_")
	count, err := u.userCollection.CountDocuments(ctx, bson.D{}, opts)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (u *UserStore) AddAddressByUserID(ctx context.Context, userId bson.ObjectID, newAddress models.Address) error {
	// Normalize legacy documents where address is null/non-array to an empty array.
	filter1 := bson.D{bson.E{Key: "_id", Value: userId}}
	_, err := u.userCollection.UpdateOne(
		ctx,
		filter1,
		mongo.Pipeline{
			bson.D{bson.E{Key: "$set", Value: bson.D{
				bson.E{Key: "address", Value: bson.D{bson.E{Key: "$cond", Value: bson.A{
					bson.D{bson.E{Key: "$isArray", Value: "$address"}},
					"$address",
					bson.A{},
				}}}},
			}}},
		},
	)
	if err != nil {
		return err
	}

	// Define the filter and update for the MongoDB operation
	filter := bson.D{bson.E{Key: "$match", Value: bson.D{bson.E{Key: "_id", Value: userId}}}}
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

	// run the aggregation pipeline
	cursor, err := u.userCollection.Aggregate(ctx, mongo.Pipeline{
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

	filter = bson.D{bson.E{Key: "_id", Value: userId}}
	update := bson.D{bson.E{Key: "$push", Value: bson.D{bson.E{Key: "address", Value: newAddress}}}}

	result, err := u.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (u *UserStore) UpdateHomeAddressByUserID(ctx context.Context, userId bson.ObjectID, newAddress models.Address) error {
	return u.updateAddressByFixedIndex(ctx, userId, 0, newAddress)
}

func (u *UserStore) UpdateWorkAddressByUserID(ctx context.Context, userId bson.ObjectID, newAddress models.Address) error {
	return u.updateAddressByFixedIndex(ctx, userId, 1, newAddress)
}

func (u *UserStore) updateAddressByFixedIndex(ctx context.Context, userId bson.ObjectID, index int, newAddress models.Address) error {
	setFields, err := u.buildAddressSetFields(index, newAddress)
	if err != nil {
		return err
	}

	filter := bson.D{
		bson.E{Key: "_id", Value: userId},
		bson.E{Key: fmt.Sprintf("address.%d", index), Value: bson.D{bson.E{Key: "$exists", Value: true}}},
	}
	result, err := u.userCollection.UpdateOne(ctx, filter, bson.D{bson.E{Key: "$set", Value: setFields}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		err = u.userCollection.FindOne(ctx, bson.M{"_id": userId}).Err()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrUserNotFound
		}
		if err != nil {
			return err
		}
		return ErrAddressNotFound
	}
	return nil
}

func (u *UserStore) buildAddressSetFields(index int, newAddress models.Address) (bson.D, error) {
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

// func (u *UserStore) testupdateAddressByFixedIndex(ctx context.Context, userId bson.ObjectID, index int, newAddress models.Address) error {
// 	setFields, err := u.buildAddressSetFields(index, newAddress)
// 	if err != nil {
// 		return err
// 	}

// 	sesssion, err := u.userCollection.Database().Client().StartSession()
// 	if err != nil {
// 		return err
// 	}
// 	defer sesssion.EndSession(ctx)

// 	err = withTransaction(*sesssion, ctx, func(ctx context.Context) error {
// 		filter := bson.D{
// 			bson.E{Key: "_id", Value: userId},
// 			bson.E{Key: fmt.Sprintf("address.%d", index), Value: bson.D{bson.E{Key: "$exists", Value: true}}},
// 		}
// 		result, err := u.userCollection.UpdateOne(ctx, filter, bson.D{bson.E{Key: "$set", Value: setFields}})
// 		if err != nil {
// 			return err
// 		}
// 		if result.MatchedCount == 0 {
// 			err = u.userCollection.FindOne(ctx, bson.M{"_id": userId}).Err()
// 			if errors.Is(err, mongo.ErrNoDocuments) {
// 				return ErrUserNotFound
// 			}
// 			if err != nil {
// 				return err
// 			}
// 			return ErrAddressNotFound
// 		}
// 		return nil
// 	})

// 	return err
// }
