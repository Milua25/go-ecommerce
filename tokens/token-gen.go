package tokens

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// token details struct to hold the token information
type TokenDetails struct {
	Email     string
	FirstName string
	LastName  string
	Uid       string
	jwt.RegisteredClaims
}

func GenerateToken(email, uid, expires, refreshExpires, secretkey string) (string, string, error) {

	expiresDuration, err := time.ParseDuration(expires)
	if err != nil {
		return "", "", fmt.Errorf("invalid expiration duration: %w", err)
	}
	refreshDuration, err := time.ParseDuration(refreshExpires)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh expiration duration: %w", err)
	}

	claims := TokenDetails{
		Email: email,
		// FirstName: firstName,
		// LastName:  lastName,
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "e-commerce-backend",
		},
	}

	refreshClaims := TokenDetails{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "e-commerce-backend",
		},
	}

	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	signedToken, err := token.SignedString([]byte(secretkey))
	if err != nil {
		return "", "", fmt.Errorf("error signing token: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	signedRefreshToken, err := refreshToken.SignedString([]byte(secretkey))
	if err != nil {
		return "", "", fmt.Errorf("error signing refresh token: %w", err)
	}
	return signedToken, signedRefreshToken, nil
}

func ValidateToken(signedToken, secretkey string) (*TokenDetails, error) {
	token, err := jwt.ParseWithClaims(signedToken, &TokenDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretkey), nil
	})
	if token == nil {
		return nil, fmt.Errorf("token is nil")
	}

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token has expired")
		}
		return nil, fmt.Errorf("error validating token: %w", err)
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	claims, ok := token.Claims.(*TokenDetails)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.ExpiresAt.Unix() < time.Now().Local().Unix() {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

func UpdateAllTokens(ctx context.Context, signedToken, refreshToken *string, userId string, userData *mongo.Collection) error {
	var updateObj bson.M

	updateObj = bson.M{}
	if signedToken != nil {
		updateObj["token"] = *signedToken
	} else {
		updateObj["token"] = nil
	}
	if refreshToken != nil {
		updateObj["refresh_token"] = *refreshToken
	} else {
		updateObj["refresh_token"] = nil
	}

	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj["updated_at"] = updatedAt

	filter := bson.M{"user_id": userId}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := userData.UpdateOne(ctx, filter, bson.M{"$set": updateObj}, opts)
	if err != nil {
		return fmt.Errorf("error updating tokens: %w", err)
	}

	return nil
}
